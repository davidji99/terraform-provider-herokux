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

func resourceHerokuxPipelineMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPipelineMemberCreate,
		ReadContext:   resourceHerokuxPipelineMemberRead,
		UpdateContext: resourceHerokuxPipelineMemberUpdate,
		DeleteContext: resourceHerokuxPipelineMemberDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPipelineMemberImport,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"pipeline_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
		},
	}
}

func resourceHerokuxPipelineMemberImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API

	compositeID, parseErr := parseCompositeIDCustom(d.Id(), ":", 2)
	if parseErr != nil {
		return nil, parseErr
	}

	pipelineID := compositeID[0]
	email := compositeID[1]

	membership, _, readErr := client.Platform.FindPipelineMembersByEmail(pipelineID, email)
	if readErr != nil {
		return nil, readErr
	}

	d.SetId(membership.GetID())

	d.Set("pipeline_id", membership.GetPipeline().ID)
	d.Set("email", membership.GetUser().GetEmail())
	setPermissionsInState(d, membership.Permissions)

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPipelineMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	opts := &platform.PipelineMembershipRequestOpts{}

	opts.PipelineID = getPipelineID(d)
	opts.Permissions = getPermissions(d)
	opts.Email = getEmail(d)

	log.Printf("[DEBUG] Adding %s to pipeline %s", opts.Email, opts.PipelineID)

	membership, _, addErr := client.Platform.AddPipelineMember(opts)
	if addErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[DEBUG] Unable to add %s to pipeline %s", opts.Email, opts.PipelineID),
			Detail:   addErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Added %s to pipeline %s", opts.Email, opts.PipelineID)

	// Set resource to the membership ID
	d.SetId(membership.GetID())

	return resourceHerokuxPipelineMemberRead(ctx, d, meta)
}

func resourceHerokuxPipelineMemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	if d.HasChange("permissions") {
		permissions := getPermissions(d)

		log.Printf("[DEBUG] Updating permissions for membership %s", d.Id())

		_, _, updatErr := client.Platform.UpdatePipelineMemberPermissions(d.Id(), permissions)
		if updatErr != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("[DEBUG] Unable to update permissions for membership %s", d.Id()),
				Detail:   updatErr.Error(),
			})
			return diags
		}

		log.Printf("[DEBUG] Updated permissions for membership %s", d.Id())
	}

	return resourceHerokuxPipelineMemberRead(ctx, d, meta)
}

func resourceHerokuxPipelineMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	email := getEmail(d)
	pipelineID := getPipelineID(d)

	membership, _, readErr := client.Platform.FindPipelineMembersByEmail(pipelineID, email)
	if readErr != nil {
		_, notFound := readErr.(platform.PermissionNotFoundError)
		if notFound {
			// Remove resource from state if by chance the user was removed from the pipeline manually.
			log.Printf("[DEBUG] No permissions found for %s on pipeline %s", email, pipelineID)
			d.SetId("")
			return nil
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to retrieve %s's membership on pipeline %s", email, pipelineID),
			Detail:   readErr.Error(),
		})
		return diags
	}

	d.Set("pipeline_id", membership.GetPipeline().ID)
	d.Set("email", membership.GetUser().GetEmail())
	setPermissionsInState(d, membership.Permissions)

	return diags
}

func resourceHerokuxPipelineMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	log.Printf("[DEBUG] Removing membership %s from pipeline", d.Id())

	_, deleteErr := client.Platform.RemovePipelineMember(d.Id())
	if deleteErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to remove membership %s", d.Id()),
			Detail:   deleteErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Removed membership %s from pipeline", d.Id())

	d.SetId("")

	return diags
}
