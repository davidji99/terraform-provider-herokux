package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxShieldPrivateSpace_Basic(t *testing.T) {
	name := fmt.Sprintf("tftest-%s", acctest.RandString(15))
	teamID := testAccConfig.GetTeamIDorSkip(t)
	url := "https://somename:somesecret@loghost.example.com/logpath"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxShieldPrivateSpace_basic(name, teamID, url),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_shield_private_space.foobar", "name", name),
					resource.TestCheckResourceAttr(
						"herokux_shield_private_space.foobar", "team_id", teamID),
					resource.TestCheckResourceAttr(
						"herokux_shield_private_space.foobar", "region", "virginia"),
					resource.TestCheckResourceAttr(
						"herokux_shield_private_space.foobar", "log_drain", url),
					resource.TestCheckResourceAttr(
						"herokux_shield_private_space.foobar", "cidr", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(
						"herokux_shield_private_space.foobar", "data_cidr", "10.1.0.0/16"),
					resource.TestCheckResourceAttr(
						"herokux_shield_private_space.foobar", "outbound_ips.#", "4"),
					resource.TestCheckResourceAttr(
						"herokux_shield_private_space.foobar", "is_shield", "true"),
					resource.TestCheckResourceAttrSet(
						"herokux_shield_private_space.foobar", "team_name"),
				),
			},
			{
				Config: testAccCheckHerokuxShieldPrivateSpace_basic(name+"-edit", teamID, url+"123"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_shield_private_space.foobar", "name", name+"-edit"),
					resource.TestCheckResourceAttr(
						"herokux_shield_private_space.foobar", "log_drain", url+"123"),
				),
			},
		},
	})
}

func testAccCheckHerokuxShieldPrivateSpace_basic(name, teamID, logDrain string) string {
	return fmt.Sprintf(`
resource "herokux_shield_private_space" "foobar" {
	name = "%s"
	team_id = "%s"
	region = "virginia"
	log_drain = "%s"
}
`, name, teamID, logDrain)
}
