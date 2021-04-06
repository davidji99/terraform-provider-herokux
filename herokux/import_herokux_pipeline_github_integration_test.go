package herokux

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPipelineGithubIntegration_importBasic(t *testing.T) {
	pipelineID := testAccConfig.GetPipelineIDorSkip(t)
	orgRepo := testAccConfig.GetGithubOrgRepoorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPipelineGithubIntegration_basic(pipelineID, orgRepo),
			},
			{
				ResourceName:      "herokux_pipeline_github_integration.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
