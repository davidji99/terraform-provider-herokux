package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresBackupSchedule_Basic(t *testing.T) {
	postgresID := testAccConfig.GetPostgresIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresBackupSchedule_basic(postgresID, "Africa/Abidjan", 5),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_backup_schedule.foobar", "postgres_id", postgresID),
					resource.TestCheckResourceAttr(
						"herokux_postgres_backup_schedule.foobar", "hour", "5"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_backup_schedule.foobar", "timezone", "Africa/Abidjan"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_backup_schedule.foobar", "name"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_backup_schedule.foobar", "retain_weeks"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_backup_schedule.foobar", "retain_months"),
				),
			},
			{
				Config: testAccCheckHerokuxPostgresBackupSchedule_basic(postgresID, "Asia/Barnaul", 20),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_backup_schedule.foobar", "postgres_id", postgresID),
					resource.TestCheckResourceAttr(
						"herokux_postgres_backup_schedule.foobar", "hour", "20"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_backup_schedule.foobar", "timezone", "Asia/Barnaul"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_backup_schedule.foobar", "name"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_backup_schedule.foobar", "retain_weeks"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_backup_schedule.foobar", "retain_months"),
				),
			},
		},
	})
}

func TestAccHerokuxPostgresBackupSchedule_BasicTimezoneWithUnderscore(t *testing.T) {
	postgresID := testAccConfig.GetPostgresIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresBackupSchedule_basic(postgresID, "America/New_York", 3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_backup_schedule.foobar", "postgres_id", postgresID),
					resource.TestCheckResourceAttr(
						"herokux_postgres_backup_schedule.foobar", "hour", "3"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_backup_schedule.foobar", "timezone", "America/New_York"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_backup_schedule.foobar", "name"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_backup_schedule.foobar", "retain_weeks"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_backup_schedule.foobar", "retain_months"),
				),
			},
		},
	})
}

func TestAccHerokuxPostgresBackupSchedule_BasicZeroHour(t *testing.T) {
	postgresID := testAccConfig.GetPostgresIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresBackupSchedule_NoTimeZone(postgresID, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_backup_schedule.foobar", "postgres_id", postgresID),
					resource.TestCheckResourceAttr(
						"herokux_postgres_backup_schedule.foobar", "hour", "0"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_backup_schedule.foobar", "timezone", "UTC"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_backup_schedule.foobar", "name"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_backup_schedule.foobar", "retain_weeks"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_backup_schedule.foobar", "retain_months"),
				),
			},
		},
	})
}

func testAccCheckHerokuxPostgresBackupSchedule_basic(postgresID, timezone string, hour int) string {
	return fmt.Sprintf(`
resource "herokux_postgres_backup_schedule" "foobar" {
	postgres_id = "%s"
	hour = %d
	timezone = "%s"
}
`, postgresID, hour, timezone)
}

func testAccCheckHerokuxPostgresBackupSchedule_NoTimeZone(postgresID string, hour int) string {
	return fmt.Sprintf(`
resource "herokux_postgres_backup_schedule" "foobar" {
	postgres_id = "%s"
	hour = %d
}
`, postgresID, hour)
}
