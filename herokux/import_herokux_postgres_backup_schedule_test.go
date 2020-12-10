package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccHerokuxPostgresBackupSchedule_importBasic(t *testing.T) {
	postgresID := testAccConfig.GetPostgresIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresBackupSchedule_basic(postgresID, "UTC", 0),
			},
			{
				ResourceName: "herokux_postgres_backup_schedule.foobar",
				ImportStateIdFunc: testAccHerokuxPostgresBackupScheduleImportStateIDFunc(
					"herokux_postgres_backup_schedule.foobar"),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccHerokuxPostgresBackupScheduleImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("[ERROR] Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["postgres_id"], rs.Primary.Attributes["id"]), nil
	}
}
