package herokux

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceHerokuxMTLSIPRules() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHerokuxKafkaMTLSIPRulesRead,
		Schema: map[string]*schema.Schema{
			"kafka_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},

			"rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceHerokuxKafkaMTLSIPRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	kafkaID := getKafkaID(d)
	rules := make([]map[string]string, 0)

	ipRules, _, listErr := client.Kafka.ListMTLSIPRules(kafkaID)
	if listErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to retrieve MTLS IP rules for kafka %s", kafkaID),
			Detail:   listErr.Error(),
		})
		return diags
	}

	for _, ipRule := range ipRules {
		rule := map[string]string{
			"id":          ipRule.GetID(),
			"cidr":        ipRule.GetCIDR(),
			"description": ipRule.GetDescription(),
			"status":      ipRule.GetStatus().ToString(),
		}
		rules = append(rules, rule)
	}

	d.SetId(kafkaID)
	d.Set("kafka_id", kafkaID)
	d.Set("rules", rules)

	return diags
}
