package herokux

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceHerokuxAddons() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHerokuxAddonsRead,
		Schema: map[string]*schema.Schema{
			"app_name_regex": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"addon_name_regex"},
				ValidateFunc:  validation.StringIsValidRegExp,
			},

			"addon_name_regex": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"app_name_regex"},
				ValidateFunc:  validation.StringIsValidRegExp,
			},

			"addons": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"app_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"app_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceHerokuxAddonsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	allAddons, err := meta.(*Config).API.Platform.ListAddons(ctx, "")
	if err != nil {
		return diag.FromErr(err)
	}
	id := "all"
	var appFilter, addonFilter *regexp.Regexp
	if v, ok := d.GetOk("app_name_regex"); ok {
		appFilter = regexp.MustCompile(v.(string))
		id = "app_name_regex:" + v.(string)
	}
	if v, ok := d.GetOk("addon_name_regex"); ok {
		addonFilter = regexp.MustCompile(v.(string))
		id = "addon_name_regex:" + v.(string)
	}

	var addons []map[string]string
	for _, addon := range allAddons {
		if (appFilter != nil && !appFilter.MatchString(addon.App.Name)) || (addonFilter != nil && !addonFilter.MatchString(addon.Name)) {
			continue
		}
		addons = append(addons, map[string]string{
			"app_id":   addon.App.ID,
			"app_name": addon.App.Name,
			"name":     addon.Name,
			"state":    addon.State,
		})
	}

	if err := d.Set("addons", addons); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	return nil
}
