package herokux

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceHerokuxRegistryImage() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHerokuxRegistryImageRead,
		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},

			"process_type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"docker_tag": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "latest",
			},

			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"schema_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"digest": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"number_of_layers": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceHerokuxRegistryImageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Config).API

	appID := getAppID(d)
	processType := d.Get("process_type").(string)
	dockerTag := d.Get("docker_tag").(string)

	image, _, getErr := client.Registry.GetAppProcessManifests(appID, processType, dockerTag)
	if getErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary: fmt.Sprintf("Unable to retrieve image info for app %s, process %s, docker tag %s",
				appID, processType, dockerTag),
			Detail: getErr.Error(),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", appID, processType, dockerTag))
	d.Set("app_id", appID)
	d.Set("process_type", processType)
	d.Set("docker_tag", dockerTag)
	d.Set("size", image.GetConfig().GetSize())
	d.Set("schema_version", image.GetSchemaVersion())
	d.Set("digest", image.GetConfig().GetDigest())
	d.Set("number_of_layers", len(image.Layers))

	return diags
}
