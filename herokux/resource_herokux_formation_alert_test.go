package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

// IMPORTANT: These test only works on an APP that hadn't had an alert set previously.

func TestAccHerokuxFormationAlertLatency_Basic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	processType := "web"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxFormationAlert_basic(appID, processType, "LATENCY",
					"1202", 10, 1440),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "process_type", processType),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "threshold", "1202"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "sensitivity", "10"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "is_active", "true"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "notification_frequency", "1440"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "notification_channels.0", "app"),
				),
			},
			{
				Config: testAccCheckHerokuxFormationAlert_basic(appID, processType, "LATENCY",
					"89", 10, 5),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "process_type", processType),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "threshold", "89"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "sensitivity", "10"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "is_active", "true"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "notification_frequency", "5"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "notification_channels.0", "app"),
				),
			},
		},
	})
}

func TestAccHerokuxFormationAlertErrorRate_Basic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	processType := "web"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxFormationAlert_basic(appID, processType, "ERROR_RATE",
					"0.042", 5, 60),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "process_type", processType),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "threshold", "0.042"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "sensitivity", "5"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "is_active", "true"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "notification_frequency", "60"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "notification_channels.0", "app"),
				),
			},
			{
				Config: testAccCheckHerokuxFormationAlert_basic(appID, processType, "ERROR_RATE",
					"0.42", 10, 5),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "process_type", processType),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "threshold", "0.42"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "sensitivity", "10"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "is_active", "true"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "notification_frequency", "5"),
					resource.TestCheckResourceAttr(
						"herokux_formation_alert.foobar", "notification_channels.0", "app"),
				),
			},
		},
	})
}

func testAccCheckHerokuxFormationAlert_basic(appID, processType, alertName, threshold string, sensitivity, emailReminderFrequency int) string {
	return fmt.Sprintf(`
resource "herokux_formation_alert" "foobar" {
	app_id = "%s"
	process_type = "%s"
	name = "%s"
	threshold = "%s"
	sensitivity = %d
	is_active = true
	notification_frequency = %d
	notification_channels = ["app"]
}
`, appID, processType, alertName, threshold, sensitivity, emailReminderFrequency)
}
