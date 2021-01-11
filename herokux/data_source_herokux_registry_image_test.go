package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccDatasourceHerokuxRegistryImage_Basic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	processType := "web"
	dockerTag := "latest"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxRegistryImageDataSource_Basic(appID, processType, dockerTag),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.herokux_registry_image.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"data.herokux_registry_image.foobar", "process_type", processType),
					resource.TestCheckResourceAttr(
						"data.herokux_registry_image.foobar", "docker_tag", dockerTag),
					resource.TestCheckResourceAttrSet(
						"data.herokux_registry_image.foobar", "size"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_registry_image.foobar", "schema_version"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_registry_image.foobar", "digest"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_registry_image.foobar", "number_of_layers"),
				),
			},
		},
	})
}

func TestAccDatasourceHerokuxRegistryImage_NotFound(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	processType := "web123"
	dockerTag := "latest"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckHerokuxRegistryImageDataSource_Basic(appID, processType, dockerTag),
				ExpectError: regexp.MustCompile(`Unable to retrieve image info`),
			},
		},
	})
}

func testAccCheckHerokuxRegistryImageDataSource_Basic(appID, processType, dockerTag string) string {
	return fmt.Sprintf(`
data "herokux_registry_image" "foobar" {
  app_id = "%s"
  process_type = "%s"
  docker_tag = "%s"
}
`, appID, processType, dockerTag)
}
