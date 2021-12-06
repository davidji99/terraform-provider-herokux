package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccDatasourceHerokuxAppAddons_Basic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	addonServiceName := ""

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAppAddons_Basic(appID, addonServiceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.herokux_app_addons.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"data.herokux_app_addons.foobar", "addon_service_name", addonServiceName),
					resource.TestCheckResourceAttrSet(
						"data.herokux_app_addons.foobar", "addons"),
				),
			},
		},
	})
}

func TestAccDatasourceHerokuxAppAddons_Filter(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	addonServiceName := "heroku-postgresql"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAppAddons_Basic(appID, addonServiceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.herokux_app_addons.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"data.herokux_app_addons.foobar", "addon_service_name", addonServiceName),
					resource.TestCheckResourceAttrSet(
						"data.herokux_app_addons.foobar", "addons"),
				),
			},
		},
	})
}

func TestAccDatasourceHerokuxAppAddons_NotFound(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	addonServiceName := "non-existent"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckHerokuxAppAddons_Basic(appID, addonServiceName),
				ExpectError: regexp.MustCompile(`Could not find the requested add-ons installed`),
			},
		},
	})
}

func testAccCheckHerokuxAppAddons_Basic(appID, addonServiceName string) string {
	return fmt.Sprintf(`
data "herokux_app_addons" "foobar" {
  app_id = "%s"
  addon_service_name = "%s"
}
`, appID, addonServiceName)
}
