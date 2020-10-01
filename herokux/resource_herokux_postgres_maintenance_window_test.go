package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

func testAccCheckHerokuxPostgresMaintenanceWindow_basic(postgresID, window string) string {
	return fmt.Sprintf(`
resource "herokux_postgres_maintenance_window" "foobar" {
	postgres_id = "%s"
	window = "%s"
}
`, postgresID, window)
}
