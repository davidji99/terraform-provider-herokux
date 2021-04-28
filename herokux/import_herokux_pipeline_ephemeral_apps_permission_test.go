package herokux

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPipelineEphemeralAppsPermission_importBasic(t *testing.T) {
	pipelineID := testAccConfig.GetPipelineIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPipelineEphemeralAppsPermission_basic(pipelineID,
					"\"view\", \"deploy\", \"operate\""),
			},
			{
				ResourceName:      "herokux_pipeline_ephemeral_apps_permission.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
