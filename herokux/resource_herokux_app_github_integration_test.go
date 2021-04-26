package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxAppGithubIntegration_Basic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAppGithubIntegration_basic(appID, "main",
					"true", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_app_github_integration.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_app_github_integration.foobar", "branch", "main"),
					resource.TestCheckResourceAttr(
						"herokux_app_github_integration.foobar", "auto_deploy", "true"),
					resource.TestCheckResourceAttr(
						"herokux_app_github_integration.foobar", "wait_for_ci", "true"),
					resource.TestCheckResourceAttrSet(
						"herokux_app_github_integration.foobar", "repository"),
					resource.TestCheckResourceAttrSet(
						"herokux_app_github_integration.foobar", "repository_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_app_github_integration.foobar", "integration_id"),
				),
			},
			{
				Config: testAccCheckHerokuxAppGithubIntegration_basic(appID, "bump-versions",
					"true", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_app_github_integration.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_app_github_integration.foobar", "branch", "bump-versions"),
					resource.TestCheckResourceAttr(
						"herokux_app_github_integration.foobar", "auto_deploy", "true"),
					resource.TestCheckResourceAttr(
						"herokux_app_github_integration.foobar", "wait_for_ci", "false"),
					resource.TestCheckResourceAttrSet(
						"herokux_app_github_integration.foobar", "repository"),
					resource.TestCheckResourceAttrSet(
						"herokux_app_github_integration.foobar", "repository_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_app_github_integration.foobar", "integration_id"),
				),
			},
		},
	})
}

func testAccCheckHerokuxAppGithubIntegration_basic(appID, branch, autoDeploy, waitForCI string) string {
	return fmt.Sprintf(`
resource "herokux_app_github_integration" "foobar" {
	app_id = "%s"
	branch = "%s"
	auto_deploy = %s
	wait_for_ci = %s
}
`, appID, branch, autoDeploy, waitForCI)
}
