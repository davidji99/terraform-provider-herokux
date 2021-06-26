package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
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

func TestAccHerokuxAppContainerRelease_DoubleRelease(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	imageID := testAccConfig.GetImageIDorSkip(t)
	processType := "web"
	processType2 := "worker"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAppContainerRelease_doubleRelease(appID, imageID, processType, processType2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_app_container_release.web", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_app_container_release.web", "image_id", imageID),
					resource.TestCheckResourceAttr(
						"herokux_app_container_release.web", "process_type", "web"),
					resource.TestCheckResourceAttr(
						"herokux_app_container_release.worker", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_app_container_release.worker", "image_id", imageID),
					resource.TestCheckResourceAttr(
						"herokux_app_container_release.worker", "process_type", "worker"),
				),
			},
		},
	})
}

func TestAccHerokuxAppContainerRelease_InvalidImageID(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	imageID := "some_invalid_image_id"
	processType := "web"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckHerokuxAppContainerRelease_basic(appID, imageID, processType),
				ExpectError: regexp.MustCompile(`invalid image ID`),
			},
		},
	})
}

func TestAccHerokuxAppContainerRelease_BasicUsingDataSource(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	imageID := testAccConfig.GetImageIDorSkip(t)
	processType := "web"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAppContainerRelease_basicDataSource(appID, processType),
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

func testAccCheckHerokuxAppContainerRelease_doubleRelease(appID, imageID, processType1, processType2 string) string {
	return fmt.Sprintf(`
resource "herokux_app_container_release" "web" {
	app_id = "%[1]s"
	image_id = "%[2]s"
	process_type = "%[3]s"
}

resource "herokux_app_container_release" "worker" {
	app_id = "%[1]s"
	image_id = "%[2]s"
	process_type = "%[4]s"
}
`, appID, imageID, processType1, processType2)
}

func testAccCheckHerokuxAppContainerRelease_basicDataSource(appID, processType string) string {
	return fmt.Sprintf(`
data "herokux_registry_image" "foobar" {
  app_id = "%[1]s"
  process_type = "%[2]s"
}

resource "herokux_app_container_release" "foobar" {
	app_id = "%[1]s"
	image_id = data.herokux_registry_image.foobar.digest
	process_type = "%[2]s"
}
`, appID, processType)
}
