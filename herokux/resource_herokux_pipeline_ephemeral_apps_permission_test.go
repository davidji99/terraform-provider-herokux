package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPipelineEphemeralAppsPermission_Basic(t *testing.T) {
	pipelineID := testAccConfig.GetPipelineIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPipelineEphemeralAppsPermission_basic(pipelineID, "\"view\""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_pipeline_ephemeral_apps_permission.foobar", "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(
						"herokux_pipeline_ephemeral_apps_permission.foobar", "permissions.#", "1"),
					resource.TestCheckResourceAttrSet(
						"herokux_pipeline_ephemeral_apps_permission.foobar", "pipeline_name"),
					resource.TestCheckResourceAttrSet(
						"herokux_pipeline_ephemeral_apps_permission.foobar", "owner_id"),
				),
			},
			{
				Config: testAccCheckHerokuxPipelineEphemeralAppsPermission_basic(pipelineID, "\"view\", \"deploy\", \"operate\", \"manage\""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_pipeline_ephemeral_apps_permission.foobar", "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(
						"herokux_pipeline_ephemeral_apps_permission.foobar", "permissions.#", "4"),
					resource.TestCheckResourceAttrSet(
						"herokux_pipeline_ephemeral_apps_permission.foobar", "pipeline_name"),
					resource.TestCheckResourceAttrSet(
						"herokux_pipeline_ephemeral_apps_permission.foobar", "owner_id"),
				),
			},
			{
				Config: testAccCheckHerokuxPipelineEphemeralAppsPermission_basic(pipelineID, "\"view\", \"deploy\", \"operate\""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_pipeline_ephemeral_apps_permission.foobar", "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(
						"herokux_pipeline_ephemeral_apps_permission.foobar", "permissions.#", "3"),
					resource.TestCheckResourceAttrSet(
						"herokux_pipeline_ephemeral_apps_permission.foobar", "pipeline_name"),
					resource.TestCheckResourceAttrSet(
						"herokux_pipeline_ephemeral_apps_permission.foobar", "owner_id"),
				),
			},
		},
	})
}

func testAccCheckHerokuxPipelineEphemeralAppsPermission_basic(pipelineID, orgRepo string) string {
	return fmt.Sprintf(`
resource "herokux_pipeline_ephemeral_apps_permission" "foobar" {
	pipeline_id = "%s"
	permissions = [%s]
}
`, pipelineID, orgRepo)
}
