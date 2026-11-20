package herokux

import (
	"context"
	"log"
	"regexp"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceHerokuxSpaceApps() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHerokuxSpaceAppsRead,
		Schema: map[string]*schema.Schema{
			"space_regex": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},

			"apps": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"web_url": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"stack": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceHerokuxSpaceAppsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var appsFiltered []heroku.App

	platformAPI := meta.(*Config).API.Platform
	spaceRegex := d.Get("space_regex").(string)
	log.Printf("[DEBUG] space_regex: %s", spaceRegex)

	allApps, appsListErr := platformAPI.ListApps(ctx)
	if appsListErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to retrieve apps",
			Detail:   appsListErr.Error(),
		})
		return diags
	}

	regex := regexp.MustCompile(spaceRegex)

	log.Printf("[DEBUG] Found %d apps", len(allApps))

	for _, a := range allApps {
		if a.Space == nil {
			continue
		}

		if regex.MatchString(a.Space.ID) || regex.MatchString(a.Space.Name) {
			appsFiltered = append(appsFiltered, a)
		}
	}

	log.Printf("[DEBUG] Number of apps after filtering by space name: %d", len(appsFiltered))

	var apps []map[string]string

	for _, app := range appsFiltered {
		apps = append(apps, map[string]string{
			"id":      app.ID,
			"name":    app.Name,
			"web_url": app.WebURL,
			"stack":   app.Stack.Name,
			"region":  app.Region.Name,
		})
	}

	if err := d.Set("apps", apps); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("space_regex:" + spaceRegex)

	return nil
}
