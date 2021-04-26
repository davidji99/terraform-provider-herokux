package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api/kolkrabbi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
)

func resourceHerokuxAppGithubIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxAppGithubIntegrationCreate,
		ReadContext:   resourceHerokuxAppGithubIntegrationRead,
		UpdateContext: resourceHerokuxAppGithubIntegrationUpdate,
		DeleteContext: resourceHerokuxAppGithubIntegrationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"branch": {
				Type:     schema.TypeString,
				Required: true,
			},

			"auto_deploy": {
				Type:     schema.TypeBool,
				Required: true,
			},

			"wait_for_ci": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"repository": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"repository_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"integration_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHerokuxAppGithubIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	appID := getAppID(d)
	opts := &kolkrabbi.AppGhIntegrationRequest{}

	if v, ok := d.GetOk("branch"); ok {
		vb := v.(string)
		log.Printf("[DEBUG] branch: %s", vb)
		opts.Branch = vb
	}

	autoDeploy := d.Get("auto_deploy").(bool)
	log.Printf("[DEBUG] auto_deploy: %v", autoDeploy)
	opts.AutoDeploy = &autoDeploy

	if v, ok := d.GetOkExists("wait_for_ci"); ok {
		vb := v.(bool)
		log.Printf("[DEBUG] wait_for_ci: %v", vb)
		opts.WaitForCI = &vb
	}

	log.Printf("[DEBUG] Creating integration with Heroku app %s and Github", appID)

	integrationData, _, createErr := client.Kolkrabbi.UpdateAppGithubIntegration(appID, opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to create integration with Heroku app %s and Github", appID),
			Detail:   createErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Created integration with Heroku app %s and Github", appID)

	// Set resource ID to be the app ID
	d.SetId(integrationData.GetAppID())

	return resourceHerokuxAppGithubIntegrationRead(ctx, d, meta)
}

func resourceHerokuxAppGithubIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	opts := &kolkrabbi.AppGhIntegrationRequest{}

	if ok := d.HasChange("branch"); ok {
		vs := d.Get("branch").(string)
		log.Printf("[DEBUG] updated branch: %s", vs)
		opts.Branch = vs
	}

	if ok := d.HasChange("auto_deploy"); ok {
		vb := d.Get("auto_deploy").(bool)
		log.Printf("[DEBUG] updated auto_deploy: %v", vb)
		opts.AutoDeploy = &vb
	}

	if ok := d.HasChange("wait_for_ci"); ok {
		vb := d.Get("wait_for_ci").(bool)
		log.Printf("[DEBUG] updated wait_for_ci: %v", vb)
		opts.WaitForCI = &vb
	}

	log.Printf("[DEBUG] Updating integration with Heroku app %s and Github", d.Id())

	_, _, updateErr := client.Kolkrabbi.UpdateAppGithubIntegration(d.Id(), opts)
	if updateErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to update integration with Heroku app %s and Github", d.Id()),
			Detail:   updateErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Updated integration with Heroku app %s and Github", d.Id())

	return resourceHerokuxAppGithubIntegrationRead(ctx, d, meta)
}

func resourceHerokuxAppGithubIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	iData, _, readErr := client.Kolkrabbi.GetAppGithubIntegration(d.Id())
	if readErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to retrieve Github integration for app %s", d.Id()),
			Detail:   readErr.Error(),
		})
		return diags
	}

	d.Set("app_id", iData.GetAppID())
	d.Set("branch", iData.GetBranch())
	d.Set("auto_deploy", iData.GetAutoDeploy())
	d.Set("wait_for_ci", iData.GetWaitForCI())
	d.Set("repository", iData.GetRepo())
	d.Set("repository_id", iData.GetRepoID())
	d.Set("integration_id", iData.GetID())

	return diags
}

func resourceHerokuxAppGithubIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	// Upon resource deletion, set everything to `false`.
	allFalse := false
	opts := &kolkrabbi.AppGhIntegrationRequest{
		AutoDeploy: &allFalse,
		WaitForCI:  &allFalse,
	}

	_, _, deleteErr := client.Kolkrabbi.UpdateAppGithubIntegration(d.Id(), opts)
	if deleteErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to delete Github integration for app %s", d.Id()),
			Detail:   deleteErr.Error(),
		})
		return diags
	}

	d.SetId("")

	return diags
}
