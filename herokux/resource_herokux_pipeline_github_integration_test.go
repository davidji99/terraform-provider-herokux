package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccHerokuxPipelineGithubIntegration_Basic(t *testing.T) {
	pipelineID := testAccConfig.GetPipelineIDorSkip(t)
	orgRepo := testAccConfig.GetGithubOrgRepoorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPipelineGithubIntegration_basic(pipelineID, orgRepo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_pipeline_github_integration.foobar", "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(
						"herokux_pipeline_github_integration.foobar", "org_repo", orgRepo),
					resource.TestCheckResourceAttrSet(
						"herokux_pipeline_github_integration.foobar", "repository_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_pipeline_github_integration.foobar", "creator_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_pipeline_github_integration.foobar", "owner_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_pipeline_github_integration.foobar", "integration_id"),
				),
			},
		},
	})
}

func TestAccHerokuxPipelineGithubIntegration_BasicInvalid(t *testing.T) {
	pipelineID := testAccConfig.GetPipelineIDorSkip(t)
	orgRepo := "myrepo"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckHerokuxPipelineGithubIntegration_basic(pipelineID, orgRepo),
				ExpectError: regexp.MustCompile(`Value must be org/repo`),
			},
		},
	})
}

func testAccCheckHerokuxPipelineGithubIntegration_basic(pipelineID, orgRepo string) string {
	return fmt.Sprintf(`
resource "herokux_pipeline_github_integration" "foobar" {
	pipeline_id = "%s"
	org_repo = "%s"
}
`, pipelineID, orgRepo)
}
