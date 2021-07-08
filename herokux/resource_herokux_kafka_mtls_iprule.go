package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/general"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"time"
)

func resourceHerokuxKafkaMTLSIPRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxKafkaMTLSIPRuleCreate,
		ReadContext:   resourceHerokuxKafkaMTLSIPRuleRead,
		DeleteContext: resourceHerokuxKafkaMTLSIPRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxKafkaMTLSIPRuleImport,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"kafka_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"cidr": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDR,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHerokuxKafkaMTLSIPRuleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API

	parsedImportID, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return nil, parseErr
	}

	kafkaID := parsedImportID[0]
	cidr := parsedImportID[1]

	// Get all IP rules
	ipRules, _, listErr := client.Kafka.ListMTLSIPRules(kafkaID)
	if listErr != nil {
		return nil, listErr
	}

	// Loop through the existing rules to find the rule ID
	var ipRule *general.MtlsIPRule
	for _, r := range ipRules {
		if r.GetCIDR() == cidr {
			ipRule = r
		}
	}

	if ipRule == nil {
		return nil, fmt.Errorf("no existing IP rule found with CIDR: %s", cidr)
	}

	// Set the ID
	d.SetId(ipRule.GetID())

	// Set state
	d.Set("kafka_id", kafkaID)
	d.Set("cidr", ipRule.GetCIDR())
	d.Set("description", ipRule.GetDescription())
	d.Set("status", ipRule.GetStatus().ToString())

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxKafkaMTLSIPRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := meta.(*Config)
	client := config.API
	opts := &general.MTLSIPRuleRequest{}
	kafkaID := getKafkaID(d)

	if v, ok := d.GetOk("cidr"); ok {
		opts.CIDR = v.(string)
		log.Printf("[DEBUG] Kafka MTLS IP rule cidr is : %s", opts.CIDR)
	}

	if v, ok := d.GetOk("description"); ok {
		opts.Description = v.(string)
		log.Printf("[DEBUG] Kafka MTLS IP rule description is : %s", opts.Description)
	}

	// Enable MTLS
	log.Printf("[DEBUG] Creating MTLS IP rule on kafka %s", kafkaID)

	ipRule, _, createErr := client.Kafka.CreateMTLSIPRule(kafkaID, opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to create MTLS IP rule on kafka %s", kafkaID),
			Detail:   createErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Waiting for MTLS IP rule %s on kafka %s to be authorized", ipRule.GetID(), kafkaID)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{general.MTLSIPRuleStatuses.AUTHORIZING.ToString()},
		Target:       []string{general.MTLSIPRuleStatuses.AUTHORIZED.ToString()},
		Refresh:      KafkaMtlsIPRuleStateRefreshFunc(client, kafkaID, ipRule.GetID()),
		Timeout:      time.Duration(config.MTLSIPRuleCreateVerifyTimeout) * time.Minute,
		PollInterval: StateRefreshPollInterval,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for MTLS IP rule %s to be authorized on kafka %s: %s",
			ipRule.GetID(), kafkaID, err.Error())
	}

	log.Printf("[DEBUG] Created MTLS IP rule on kafka %s", kafkaID)

	d.SetId(ipRule.GetID())

	return resourceHerokuxKafkaMTLSIPRuleRead(ctx, d, meta)
}

func resourceHerokuxKafkaMTLSIPRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	kafkaID := getKafkaID(d)

	ipRule, response, getErr := client.Kafka.GetMTLSIPRule(kafkaID, d.Id())
	if getErr != nil {
		if response.StatusCode == 404 {
			log.Printf("[WARN] Kafka MTLS IP rule %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to retrieve Kafka MTLS IP rule %s", d.Id()),
			Detail:   getErr.Error(),
		})
		return diags
	}

	d.Set("kafka_id", kafkaID)
	d.Set("cidr", ipRule.GetCIDR())
	d.Set("description", ipRule.GetDescription())
	d.Set("status", ipRule.GetStatus().ToString())

	return nil
}

func resourceHerokuxKafkaMTLSIPRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := meta.(*Config)
	client := config.API
	kafkaID := getKafkaID(d)

	log.Printf("[DEBUG] Deleting Kafka MTLS IP rule %s", d.Id())
	_, deleteErr := client.Kafka.DeleteMTLSIPRule(kafkaID, d.Id())
	if deleteErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to delete Kafka MTLS IP rule %s", d.Id()),
			Detail:   deleteErr.Error(),
		})
		return diags
	}

	d.SetId("")

	return nil
}

func KafkaMtlsIPRuleStateRefreshFunc(client *api.Client, kafkaID, ipRuleID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ipRule, _, getErr := client.Kafka.GetMTLSIPRule(kafkaID, ipRuleID)
		if getErr != nil {
			return nil, general.MTLSIPRuleStatuses.UNKNOWN.ToString(), getErr
		}

		if *ipRule.GetStatus() == general.MTLSIPRuleStatuses.AUTHORIZING {
			log.Printf("[DEBUG] Still waiting for Kafka MTLS IP rule %s to be authorized", ipRuleID)
			return ipRule, ipRule.GetStatus().ToString(), nil
		}

		return ipRule, ipRule.GetStatus().ToString(), nil
	}
}
