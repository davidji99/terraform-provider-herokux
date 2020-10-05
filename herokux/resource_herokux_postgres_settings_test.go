package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"testing"
)

func TestAccHerokuxPostgresSettings_Basic(t *testing.T) {
	postgresID := testAccConfig.GetAddonIDorSkip(t)
	wait := false
	connections := true
	duration := randInt(-1, 1999)
	statement := "mod"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresSettings_basic(postgresID, wait, connections, duration, statement),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_settings.foobar", "postgres_id", postgresID),
					resource.TestCheckResourceAttr(
						"herokux_postgres_settings.foobar", "log_lock_waits", "false"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_settings.foobar", "log_connections", "true"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_settings.foobar", "log_min_duration_statement", strconv.Itoa(duration)),
					resource.TestCheckResourceAttr(
						"herokux_postgres_settings.foobar", "log_statement", statement),
				),
			},
		},
	})
}

func testAccCheckHerokuxPostgresSettings_basic(postgresID string, logLocksWaits,
	logConnections bool, logMinDuration int, logStatement string) string {
	return fmt.Sprintf(`
resource "herokux_postgres_settings" "foobar" {
	postgres_id = "%s"
	log_lock_waits = %v
	log_connections = %v
	log_min_duration_statement = %v
	log_statement = "%s"
}
`, postgresID, logLocksWaits, logConnections, logMinDuration, logStatement)
}
