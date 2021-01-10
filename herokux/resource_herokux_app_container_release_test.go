package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxAppContainerRelease_Basic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	imageID := testAccConfig.GetImageIDorSkip(t)
	processType := "web"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAppContainerRelease_basic(appID, imageID, processType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_app_container_release.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_app_container_release.foobar", "image_id", imageID),
					resource.TestCheckResourceAttr(
						"herokux_app_container_release.foobar", "process_type", "web"),
				),
			},
		},
	})
}

func testAccCheckHerokuxAppContainerRelease_basic(appID, imageID, processType string) string {
	return fmt.Sprintf(`
resource "herokux_app_container_release" "foobar" {
	app_id = "%s"
	image_id = "%s"
	process_type = "%s"
}
`, appID, imageID, processType)
}
