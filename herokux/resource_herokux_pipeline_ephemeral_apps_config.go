package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/platform"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
)

func resourceHerokuxPipelineEphemeralAppsConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPipelineEphemeralAppsConfigCreate,
		ReadContext:   resourceHerokuxPipelineEphemeralAppsConfigRead,
		UpdateContext: resourceHerokuxPipelineEphemeralAppsConfigUpdate,
		DeleteContext: resourceHerokuxPipelineEphemeralAppsConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"pipeline_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"permissions": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(api.DefaultMemberPermissions, false),
				},
			},

			"pipeline_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func updatePipelineEphemeralAppsPermission(d *schema.ResourceData, meta interface{}) (*platform.Pipeline, error) {
	client := meta.(*Config).API
	opts := &platform.PipelineEphemeralAppsConfigUpdateOpts{}

	pipelineID := getPipelineID(d)
	opts.Permissions = getPermissions(d)

	// Set the other fields on opts to `true`.
	opts.Enabled = true
	opts.Synchronization = true

	log.Printf("[DEBUG] Setting ephemeral apps permissions for pipeline %s", pipelineID)

	p, _, setErr := client.Platform.UpdatePipelineEphemeralAppsConfig(pipelineID, opts)
	if setErr != nil {
		return nil, setErr
	}

	log.Printf("[DEBUG] Set ephemeral apps permissions for pipeline %s", pipelineID)

	return p, nil
}

func resourceHerokuxPipelineEphemeralAppsConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	p, setErr := updatePipelineEphemeralAppsPermission(d, meta)
	if setErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set ephemeral apps permission on %s", getPipelineID(d)),
			Detail:   setErr.Error(),
		})
		return diags
	}

	d.SetId(p.ID)

	return resourceHerokuxPipelineEphemeralAppsConfigRead(ctx, d, meta)
}

func resourceHerokuxPipelineEphemeralAppsConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	_, setErr := updatePipelineEphemeralAppsPermission(d, meta)
	if setErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set ephemeral apps permission on %s", getPipelineID(d)),
			Detail:   setErr.Error(),
		})
		return diags
	}

	return resourceHerokuxPipelineEphemeralAppsConfigRead(ctx, d, meta)
}

func resourceHerokuxPipelineEphemeralAppsConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	p, _, readErr := client.Platform.GetPipelineEphemeralAppsConfig(d.Id())
	if readErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to retrieve pipeline permissions for %s", d.Id()),
			Detail:   readErr.Error(),
		})
		return diags
	}

	d.Set("pipeline_id", p.ID)
	d.Set("pipeline_name", p.Name)
	d.Set("owner_id", p.Owner.ID)
	setPermissionsInState(d, p.GetEphemeralApps().Permissions)

	return diags
}

func resourceHerokuxPipelineEphemeralAppsConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	opts := &platform.PipelineEphemeralAppsConfigUpdateOpts{
		Enabled:         true,
		Synchronization: false,
	}

	log.Printf("[DEBUG] Unsetting ephemeral apps permissions for pipeline %s", d.Id())

	// Delete the resource by disabling the permission(s).
	_, _, deleteErr := client.Platform.UpdatePipelineEphemeralAppsConfig(d.Id(), opts)
	if deleteErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to disable ephemeral apps permissions for pipeline %s", d.Id()),
			Detail:   deleteErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Unset ephemeral apps permissions for pipeline %s", d.Id())

	d.SetId("")

	return diags
}
