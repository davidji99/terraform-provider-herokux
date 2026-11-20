package herokux

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceHerokuxAppAddons() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHerokuxAppAddonsRead,
		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
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
	platformAPI := meta.(*Config).API.Platform

	appID := getAppID(d)
	addOnServiceName := d.Get("addon_service_name").(string)
	addOns := make(map[string]string)

	addOnList, listErr := platformAPI.ListAddons(ctx, appID)
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
	if len(addOns) == 0 && addOnServiceName != "" {
		return diag.Errorf("Could not find any '%s' add-on installed in %s", addOnServiceName, appID)
	}
	if err := d.Set("addons", addOns); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%s:%s", appID, addOnServiceName))
	return nil
}
