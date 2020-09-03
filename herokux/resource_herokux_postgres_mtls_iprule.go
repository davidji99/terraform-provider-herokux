package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"time"
)

func resourceHerokuxPostgresMTLSIPRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresMTLSIPRuleCreate,
		ReadContext:   resourceHerokuxPostgresMTLSIPRuleRead,
		DeleteContext: resourceHerokuxPostgresMTLSIPRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPostgresMTLSIPRuleImport,
		},

		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

func resourceHerokuxPostgresMTLSIPRuleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API

	parsedImportID, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return nil, parseErr
	}

	dbName := parsedImportID[0]
	cidr := parsedImportID[1]

	// Get all IP rules
	ipRules, _, listErr := client.Postgres.ListMTLSIPRules(dbName)
	if listErr != nil {
		return nil, listErr
	}

	// Loop through the existing rules to find the rule ID
	var rule *postgres.MTLSIPRule
	for _, r := range ipRules {
		if r.GetCIDR() == cidr {
			rule = r
		}
	}

	if rule == nil {
		return nil, fmt.Errorf("no existing IP rule found with CIDR: %s", cidr)
	}

	// Set the ID
	d.SetId(fmt.Sprintf("%s:%s", dbName, rule.GetID()))

	// Set state
	d.Set("database_name", dbName)
	d.Set("cidr", rule.GetCIDR())
	d.Set("description", rule.GetDescription())
	d.Set("status", rule.GetStatus().ToString())

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPostgresMTLSIPRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	opts := &postgres.MTLSIPRuleRequest{}

	dbName := getDatabaseName(d)

	if v, ok := d.GetOk("cidr"); ok {
		opts.CIDR = v.(string)
		log.Printf("[DEBUG] cidr is : %s", opts.CIDR)
	}

	if v, ok := d.GetOk("description"); ok {
		opts.Description = v.(string)
		log.Printf("[DEBUG] description is : %s", opts.Description)
	}

	// Enable MTLS
	log.Printf("[DEBUG] Creating MTLS IP rule on database %s", dbName)
	ipRule, _, createErr := client.Postgres.CreateMTLSIPRule(dbName, opts)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Printf("[DEBUG] Waiting for MTLS IP rule on %s to be authorized", dbName)
	stateConf := &resource.StateChangeConf{
		Pending:      []string{postgres.MTLSIPRuleStatuses.AUTHORIZING.ToString()},
		Target:       []string{postgres.MTLSIPRuleStatuses.AUTHORIZED.ToString()},
		Refresh:      MTLSSIPRuleStateRefreshFunc(client, dbName, ipRule.GetID()),
		Timeout:      time.Duration(config.MTLSIPRuleCreateTimeout) * time.Minute,
		PollInterval: 15 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for MTLS IP rule to be authorized on %s: %s", dbName, err.Error())
	}

	d.SetId(fmt.Sprintf("%s:%s", dbName, ipRule.GetID()))

	return resourceHerokuxPostgresMTLSIPRuleRead(ctx, d, meta)
}

func resourceHerokuxPostgresMTLSIPRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	ids, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	ipRule, response, getErr := client.Postgres.GetMTLSIPRule(ids[0], ids[1])
	if getErr != nil {
		if response.StatusCode == 404 {
			log.Printf("[WARN] MTLS IP rule for %s not found, removing from state", ids[0])
			d.SetId("")
			return nil
		}
		return diag.FromErr(getErr)
	}

	d.Set("database_name", ids[0])
	d.Set("cidr", ipRule.GetCIDR())
	d.Set("description", ipRule.GetDescription())
	d.Set("status", ipRule.GetStatus().ToString())

	return nil
}

func resourceHerokuxPostgresMTLSIPRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	ids, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	log.Printf("[DEBUG] Deleting MTLS IP rule on database %s", d.Id())
	_, deleteErr := client.Postgres.DeleteMTLSIPRule(ids[0], ids[1])
	if deleteErr != nil {
		return diag.FromErr(deleteErr)
	}

	d.SetId("")

	return nil
}

func MTLSSIPRuleStateRefreshFunc(client *api.Client, dbName, ipRuleID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ipRule, _, getErr := client.Postgres.GetMTLSIPRule(dbName, ipRuleID)
		if getErr != nil {
			return nil, postgres.MTLSConfigStatuses.UNKNOWN.ToString(), getErr
		}

		if *ipRule.GetStatus() == postgres.MTLSIPRuleStatuses.AUTHORIZING {
			log.Printf("[DEBUG] Still waiting for MTLS IP rule on %s to be authorized", dbName)
			return ipRule, ipRule.GetStatus().ToString(), nil
		}

		return ipRule, ipRule.GetStatus().ToString(), nil
	}
}
