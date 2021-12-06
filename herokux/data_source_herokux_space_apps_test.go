package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

// NOTE:!!!!
// All tests in this file have hardcoded expected number of apps in spaces.
// The tests may fail depending on how many space apps the account has
// at the time of running the test.

func TestAccDatasourceHerokuxSpaceApps_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxSpaceApps_Basic(".*vald.*"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.herokux_space_apps.foobar", "space_regex", ".*vald.*"),
					resource.TestCheckResourceAttr(
						"data.herokux_space_apps.foobar", "apps.#", "9"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_space_apps.foobar", "apps.0.id"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_space_apps.foobar", "apps.0.name"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_space_apps.foobar", "apps.0.web_url"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_space_apps.foobar", "apps.0.stack"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_space_apps.foobar", "apps.0.region"),
				),
			},
		},
	})
}

func TestAccDatasourceHerokuxSpaceApps_FilterNotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxSpaceApps_Basic(`.*thisshouldbeinvalid-.*`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.herokux_space_apps.foobar", "addons.#", "0"),
				),
			},
		},
	})
}

func testAccCheckHerokuxSpaceApps_Basic(regex string) string {
	return fmt.Sprintf(`
data "herokux_space_apps" "foobar" {
	space_regex = "%s"
}
`, regex)
}
