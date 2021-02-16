package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

// IMPORTANT: this test only works on an APP that wasn't autoscaled previously.
func TestAccHerokuxFormationAutoscaling_Basic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	formationName := "web"
	minQuantity := acctest.RandIntRange(1, 8)
	maxQuantity := minQuantity + 2
	p95ResponseTime := acctest.RandIntRange(500, 1000)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxFormationAutoscaling_basic(appID, formationName, minQuantity, maxQuantity, p95ResponseTime),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "formation_name", formationName),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "is_active", "true"),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "min_quantity", fmt.Sprintf("%d", minQuantity)),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "max_quantity", fmt.Sprintf("%d", maxQuantity)),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "desired_p95_response_time", fmt.Sprintf("%d", p95ResponseTime)),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "dyno_type", "performance-l"),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "notification_channels.0", "app"),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "period", "1"),
				),
			},
			{
				Config: testAccCheckHerokuxFormationAutoscaling_basicNoNotifications(appID, formationName, minQuantity+1, maxQuantity, p95ResponseTime),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "formation_name", formationName),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "is_active", "true"),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "min_quantity", fmt.Sprintf("%d", minQuantity+1)),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "max_quantity", fmt.Sprintf("%d", maxQuantity)),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "desired_p95_response_time", fmt.Sprintf("%d", p95ResponseTime)),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "dyno_type", "performance-l"),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "notification_channels.#", "0"),
					resource.TestCheckResourceAttr(
						"herokux_formation_autoscaling.foobar", "period", "1"),
				),
			},
		},
	})
}

func testAccCheckHerokuxFormationAutoscaling_basic(appID, formationName string, min, max, p95 int) string {
	return fmt.Sprintf(`
resource "herokux_formation_autoscaling" "foobar" {
	app_id = "%s"
	formation_name = "%s"
	is_active = true
	min_quantity = %d
	max_quantity = %d
	desired_p95_response_time = %d
	dyno_type = "performance-l"
	notification_channels = ["app"]
}
`, appID, formationName, min, max, p95)
}

func testAccCheckHerokuxFormationAutoscaling_basicNoNotifications(appID, formationName string, min, max, p95 int) string {
	return fmt.Sprintf(`
resource "herokux_formation_autoscaling" "foobar" {
	app_id = "%s"
	formation_name = "%s"
	is_active = true
	min_quantity = %d
	max_quantity = %d
	desired_p95_response_time = %d
	dyno_type = "performance-l"
	notification_channels = []
}
`, appID, formationName, min, max, p95)
}
