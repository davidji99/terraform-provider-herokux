package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxAppContainerRelease_importBasic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	imageID := testAccConfig.GetImageIDorSkip(t)
	processType := "web"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAppContainerRelease_basic(appID, imageID, processType),
			},
			{
				ResourceName:      "herokux_app_container_release.foobar",
				ImportStateId:     fmt.Sprintf("%s|%s|%s", appID, imageID, processType),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
