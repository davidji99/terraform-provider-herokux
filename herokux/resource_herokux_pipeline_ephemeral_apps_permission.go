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

func resourceHerokuxPipelineEphemeralAppsPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPipelineEphemeralAppsPermissionCreate,
		ReadContext:   resourceHerokuxPipelineEphemeralAppsPermissionRead,
		UpdateContext: resourceHerokuxPipelineEphemeralAppsPermissionUpdate,
		DeleteContext: resourceHerokuxPipelineEphemeralAppsPermissionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
	opts := &platform.PipelinePermissionConfigUpdateOpts{}

	pipelineID := getPipelineID(d)
	opts.Permissions = getPermissions(d)

	// Set the other fields on opts to `true`.
	opts.Enabled = true
	opts.Synchronization = true

	log.Printf("[DEBUG] Setting ephemeral apps permissions for pipeline %s", pipelineID)

	p, _, setErr := client.Platform.UpdatePipelinePermissionConfig(pipelineID, opts)
	if setErr != nil {
		return nil, setErr
	}

	log.Printf("[DEBUG] Set ephemeral apps permissions for pipeline %s", pipelineID)

	return p, nil
}

func resourceHerokuxPipelineEphemeralAppsPermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	return resourceHerokuxPipelineEphemeralAppsPermissionRead(ctx, d, meta)
}

func resourceHerokuxPipelineEphemeralAppsPermissionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	return resourceHerokuxPipelineEphemeralAppsPermissionRead(ctx, d, meta)
}

func resourceHerokuxPipelineEphemeralAppsPermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	p, _, readErr := client.Platform.GetPipelinePermissionConfig(d.Id())
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

func resourceHerokuxPipelineEphemeralAppsPermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	opts := &platform.PipelinePermissionConfigUpdateOpts{
		Enabled:         true,
		Synchronization: false,
	}

	log.Printf("[DEBUG] Unsetting ephemeral apps permissions for pipeline %s", d.Id())

	// Delete the resource by disabling the permission(s).
	_, _, deleteErr := client.Platform.UpdatePipelinePermissionConfig(d.Id(), opts)
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
