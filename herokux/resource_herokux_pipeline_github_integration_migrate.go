package herokux

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"regexp"
)

func resourceHerokuxPipelineGithubIntegrationResourceV0() *schema.Resource {
	return &schema.Resource{
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

func resourceHerokuxPipelineGithubIntegrationStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		return nil, nil
	}

	if val, ok := rawState["github_org_repo"]; ok {
		// Set org_repo attribute value to be the same as github_org_repo.
		rawState["org_repo"] = val

		// Delete the github_org_repo attribute from state
		delete(rawState, "github_org_repo")
	}

	return rawState, nil
}
