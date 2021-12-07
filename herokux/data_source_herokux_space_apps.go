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

	platformAPI := meta.(*Config).PlatformAPI
	spaceRegex := d.Get("space_regex").(string)
	log.Printf("[DEBUG] space_regex: %s", spaceRegex)

	// Return all addons that the authenticated token can access
	// TODO: figure out a way to paginate.
	allApps, appsListErr := platformAPI.AppList(ctx, &heroku.ListRange{Max: 1000, Field: "id"})
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

	// Set data source to be a random ID given this data source encompasses multiple appsFiltered.
	d.SetId(time.Now().String())

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

	d.Set("apps", apps)

	return nil
}
