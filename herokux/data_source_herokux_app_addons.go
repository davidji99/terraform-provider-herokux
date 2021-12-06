package herokux

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceHerokuxAppAddons() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHerokuxAppAddonsRead,
		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"addon_service_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"addons": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceHerokuxAppAddonsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	platformAPI := meta.(*Config).PlatformAPI

	appID := getAppID(d)
	addOnServiceName := d.Get("addon_service_name")
	addOns := make(map[string]string)

	addOnList, listErr := platformAPI.AddOnListByApp(ctx, appID, nil)
	if listErr != nil {
		var diags diag.Diagnostics
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to retrieve add-ons for app %s", appID),
			Detail:   listErr.Error(),
		})
		return diags
	}
	for _, addOn := range addOnList {
		if addOnServiceName == "" || addOnServiceName == addOn.AddonService.Name {
			addOns[addOn.ID] = addOn.Name
		}
	}
	if len(addOns) > 0 {
		d.SetId(fmt.Sprintf("%s:%s", appID, addOnServiceName))
		d.Set("addons", addOns)
	} else {
		return diag.Errorf("Could not find any '%s' add-on installed in %s", addOnServiceName, appID)
	}
	return nil
}
