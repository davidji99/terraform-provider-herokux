package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

// NOTE:!!!!
// All tests in this file have hardcoded expected number of addons.
// The tests may fail depending on how many addons the account has
// at the time of running the test.

func TestAccDatasourceHerokuxAddons_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAddons_Basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.herokux_addons.foobar", "addons.#", "9"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.0.app_id"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.0.app_name"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.0.name"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.0.state"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.8.app_id"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.8.app_name"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.8.name"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.8.state"),
				),
			},
		},
	})
}

func TestAccDatasourceHerokuxAddons_FilterByApp(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAddons_FilterByApp(`.*dj.*`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.herokux_addons.foobar", "addons.#", "1"),
					resource.TestCheckResourceAttr(
						"data.herokux_addons.foobar", "app_name_regex", ".*dj.*"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.0.app_id"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.0.app_name"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.0.name"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.0.state"),
				),
			},
		},
	})
}

func TestAccDatasourceHerokuxAddons_FilterByAddonName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAddons_FilterByAddonName(`.*scheduler-.*`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.herokux_addons.foobar", "addons.#", "3"),
					resource.TestCheckResourceAttr(
						"data.herokux_addons.foobar", "addon_name_regex", ".*scheduler-.*"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.0.app_id"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.0.app_name"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.0.name"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.0.state"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.2.app_id"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.2.app_name"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.2.name"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_addons.foobar", "addons.2.state"),
				),
			},
		},
	})
}

func TestAccDatasourceHerokuxAddons_FilterNotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAddons_FilterByAddonName(`.*thisshouldbeinvalid-.*`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.herokux_addons.foobar", "addons.#", "0"),
				),
			},
		},
	})
}

func testAccCheckHerokuxAddons_Basic() string {
	return `
data "herokux_addons" "foobar" {}
`
}

func testAccCheckHerokuxAddons_FilterByApp(regex string) string {
	return fmt.Sprintf(`
data "herokux_addons" "foobar" {
	app_name_regex = "%s"
}
`, regex)
}

func testAccCheckHerokuxAddons_FilterByAddonName(regex string) string {
	return fmt.Sprintf(`
data "herokux_addons" "foobar" {
	addon_name_regex = "%s"
}
`, regex)
}
