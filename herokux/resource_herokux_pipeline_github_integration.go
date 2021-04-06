package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api/kolkrabbi"
	"github.com/google/go-github/v34/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/oauth2"
	"log"
	"regexp"
	"strings"
)

func resourceHerokuxPipelineGithubIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPipelineGithubIntegrationCreate,
		ReadContext:   resourceHerokuxPipelineGithubIntegrationRead,
		DeleteContext: resourceHerokuxPipelineGithubIntegrationDelete,

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

			"github_org_repo": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\S+\/\S+$`),
					"Invalid attribute value. Value must be org/repo."),
			},

			"github_repository_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"creator_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"integration_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHerokuxPipelineGithubIntegrationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPipelineGithubIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var pipelineID string
	herokuClient := meta.(*Config).API
	opts := &kolkrabbi.PipelineGHIntegrationRequest{}

	if v, ok := d.GetOk("pipeline_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] pipeline_id: %s", vs)
		pipelineID = vs
	}

	if orgRepoRaw, ok := d.GetOk("github_org_repo"); ok {
		orgRepo := strings.Split(orgRepoRaw.(string), "/")
		log.Printf("[DEBUG] org_repo: %v", orgRepo)

		org := orgRepo[0]
		repo := orgRepo[1]

		// Retrieve Github token from the Heroku integration
		ghIntegrationData, _, accountErr := herokuClient.Kolkrabbi.GetAccountInfo()
		if accountErr != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to retrieve Heroku/Github integration data. Did you connect Heroku with your Github account?",
				Detail:   accountErr.Error(),
			})
			return diags
		}

		// Initialize Github API client
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: ghIntegrationData.GetGithub().GetToken()},
		)
		tc := oauth2.NewClient(ctx, ts)
		ghClient := github.NewClient(tc)

		// Retrieve the Github ID for the repo
		repoData, _, repoErr := ghClient.Repositories.Get(ctx, org, repo)
		if repoErr != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf("Unable to retrieve Github repository ID. Is %s a valid repository?",
					orgRepoRaw.(string)),
				Detail: repoErr.Error(),
			})
			return diags
		}

		opts.Repository = repoData.GetID()
		log.Printf("[DEBUG] repository ID: %v", opts.Repository)
	}

	log.Printf("[DEBUG] Creating integration with Heroku pipeline %s and Github repository %v", pipelineID, opts.Repository)

	integationData, _, createErr := herokuClient.Kolkrabbi.CreatePipelineGithubIntegration(pipelineID, opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary: fmt.Sprintf("Unable to create integration with Heroku pipeline %s and Github repository %v", pipelineID,
				opts.Repository),
			Detail: createErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Created integration with Heroku pipeline %s and Github repository %v", pipelineID, opts.Repository)

	// Set resource ID to the pipeline ID
	d.SetId(integationData.GetPipeline().GetID())

	return resourceHerokuxPipelineGithubIntegrationRead(ctx, d, meta)
}

func resourceHerokuxPipelineGithubIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	iData, _, readErr := client.Kolkrabbi.GetPipelineGithubIntegration(d.Id())
	if readErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to retrieve Github integration for pipeline %s", d.Id()),
			Detail:   readErr.Error(),
		})
		return diags
	}

	d.Set("pipeline_id", iData.GetPipeline().GetID())
	d.Set("github_org_repo", iData.GetRepository().GetName())
	d.Set("github_repository_id", iData.GetRepository().GetID())
	d.Set("creator_id", iData.GetCreator().GetID())
	d.Set("owner_id", iData.GetOwner().GetID())
	d.Set("integration_id", iData.GetID())

	return diags
}

func resourceHerokuxPipelineGithubIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	_, deleteErr := client.Kolkrabbi.DeletePipelineGithubIntegration(d.Id())
	if deleteErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to delete Github integration for pipeline %s", d.Id()),
			Detail:   deleteErr.Error(),
		})
		return diags
	}

	return diags
}
