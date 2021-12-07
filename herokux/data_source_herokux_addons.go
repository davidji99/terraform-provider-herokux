package herokux

import (
	"context"
	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
	"time"
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
	var diags diag.Diagnostics

	platformAPI := meta.(*Config).PlatformAPI

	var isFilterByApp, isFilterByAddonName bool
	var filterByAppRegex, filterByAddonName string

	if v, ok := d.GetOk("app_name_regex"); ok {
		isFilterByApp = true
		filterByAppRegex = v.(string)
		log.Printf("[DEBUG] app_name_regex: %s", filterByAppRegex)
	}

	if v, ok := d.GetOk("addon_name_regex"); ok {
		isFilterByAddonName = true
		filterByAddonName = v.(string)
		log.Printf("[DEBUG] addon_name_regex: %s", filterByAddonName)
	}

	// Return all addons that the authenticated token can access
	// TODO: figure out a way to paginate.
	allAddons, addonListErr := platformAPI.AddOnList(ctx, &heroku.ListRange{Max: 1000, Field: "id"})
	if addonListErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to retrieve addons",
			Detail:   addonListErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Found %d addons", len(allAddons))

	// Set data source to be a random ID given this data source encompasses multiple addons.
	d.SetId(time.Now().String())

	if !isFilterByApp && !isFilterByAddonName {
		// If no filtering, set all addons in state
		setDataAddonsInState(d, allAddons)
		return nil
	}

	if isFilterByApp {
		var allAddonsFiltered heroku.AddOnListResult
		regex := regexp.MustCompile(filterByAppRegex)

		for _, addon := range allAddons {
			if regex.MatchString(addon.App.Name) {
				allAddonsFiltered = append(allAddonsFiltered, addon)
			}
		}

		log.Printf("[DEBUG] Number of addons after filtering by app name: %d", len(allAddonsFiltered))

		setDataAddonsInState(d, allAddonsFiltered)

		return nil
	}

	if isFilterByAddonName {
		var allAddonsFiltered heroku.AddOnListResult
		regex := regexp.MustCompile(filterByAddonName)

		for _, addon := range allAddons {
			if regex.MatchString(addon.Name) {
				allAddonsFiltered = append(allAddonsFiltered, addon)
			}
		}

		log.Printf("[DEBUG] Number of addons after filtering by addon name: %d", len(allAddonsFiltered))

		setDataAddonsInState(d, allAddonsFiltered)

		return nil
	}

	return nil
}

func setDataAddonsInState(d *schema.ResourceData, allAddons heroku.AddOnListResult) {
	var addons []map[string]string

	for _, addon := range allAddons {
		addons = append(addons, map[string]string{
			"app_id":   addon.App.ID,
			"app_name": addon.App.Name,
			"name":     addon.Name,
			"state":    addon.State,
		})
	}

	d.Set("addons", addons)
}
