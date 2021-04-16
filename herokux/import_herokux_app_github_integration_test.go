package herokux

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxAppGithubIntegration_importBasic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAppGithubIntegration_basic(appID, "main",
					"true", "true"),
			},
			{
				ResourceName:      "herokux_app_github_integration.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
