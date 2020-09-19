package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgres_OnlyLeader(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	plan := "private-0"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgres_OnlyLeader(appID, plan),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres.foobar", "description", "High Availablility for Foobar"),
					resource.TestCheckResourceAttr(
						"herokux_postgres.foobar", "database.#", "1"),
					//helper.TestCheckTypeSetElemAttr("herokux_postgres.foobar",
					//	"allowed_accounts.*", "123456789123"),
				),
			},
		},
	})
}

func TestAccHerokuxPostgres_LeaderAndFollower(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	plan := "private-0"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgres_LeaderAndFollower(appID, plan),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres.foobar", "description", "High Availablility for Foobar"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres.foobar", "database_leader_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres.foobar", "database_follower_id"),
					resource.TestCheckResourceAttr(
						"herokux_postgres.foobar", "database_count", "2"),
					resource.TestCheckResourceAttr(
						"herokux_postgres.foobar", "database.#", "2"),
					//helper.TestCheckTypeSetElemAttr("herokux_postgres.foobar",
					//	"allowed_accounts.*", "123456789123"),
				),
			},
		},
	})
}

func testAccCheckHerokuxPostgres_OnlyLeader(appID, plan string) string {
	return fmt.Sprintf(`
resource "herokux_postgres" "foobar" {
	database {
		position = "leader"
		app_id = "%s"
		plan = "%s"
	}
}
`, appID, plan)
}

func testAccCheckHerokuxPostgres_LeaderAndFollower(appID, plan string) string {
	return fmt.Sprintf(`
resource "herokux_postgres" "foobar" {
	database {
		position = "leader"
		app_id = "%[1]s"
		plan = "%[2]s"
	}

	database {
		position = "follower"
		app_id = "%[1]s"
		plan = "%[2]s"
	}
}
`, appID, plan)
}
