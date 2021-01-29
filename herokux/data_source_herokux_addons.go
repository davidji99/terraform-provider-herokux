package herokux

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceHerokuxAddons() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHerokuxAddonsRead,
		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"addon_service_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"addon_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"addon_names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceHerokuxAddonsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	var appID string
	var addonIDs, addonNames []string
	addonServiceName := d.Get("addon_service_name")
	if v, ok := d.GetOk("app_id"); ok {
		appID = v.(string)
		result, _, setErr := client.Platform.ListAppAddons(appID)
		if setErr != nil {
			return diag.FromErr(setErr)
		}
		if len(result) > 0 {
			d.SetId(result[0].GetID())
		}
		for _, addon := range result {
			if addonServiceName == "" || addon.GetAddonService().GetName() == addonServiceName {
				addonIDs = append(addonIDs, addon.GetID())
				addonNames = append(addonNames, addon.GetName())
			}
		}
	}
	if len(addonIDs) == 0 {
		return diag.Errorf("Could not find the requested add-ons installed in %s", appID)
	}

	d.Set("addon_ids", addonIDs)
	d.Set("addon_names", addonNames)

	return nil
}
