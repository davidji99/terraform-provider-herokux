package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestAccHerokuxPostgresMaintenanceWindow_Basic(t *testing.T) {
	postgresID := testAccConfig.GetAddonIDorSkip(t)
	window := fmt.Sprintf("%ss 10:30", strings.Title(time.Weekday(randInt(0, 6)).String()))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresMaintenanceWindow_basic(postgresID, window),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_maintenance_window.foobar", "postgres_id", postgresID),
					resource.TestCheckResourceAttr(
						"herokux_postgres_maintenance_window.foobar", "window", window),
				),
			},
		},
	})
}

func TestAccHerokuxPostgresMaintenanceWindow_BasicInvalidWindow(t *testing.T) {
	redisID := testAccConfig.GetRedisIDorSkip(t)
	window := fmt.Sprintf("%ss1 10:30", strings.Title(time.Weekday(randInt(0, 6)).String()))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckHerokuxRedisMaintenanceWindow_basic(redisID, window),
				ExpectError: regexp.MustCompile(`maintenance window format should be 'Days HH:MM' where where MM is 00 or 30`),
			},
		},
	})
}

func testAccCheckHerokuxPostgresMaintenanceWindow_basic(postgresID, window string) string {
	return fmt.Sprintf(`
resource "herokux_postgres_maintenance_window" "foobar" {
	postgres_id = "%s"
	window = "%s"
}
`, postgresID, window)
}
