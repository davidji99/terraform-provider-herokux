package herokux

import (
	"fmt"
	helper "github.com/davidji99/terraform-provider-herokux/helper/test"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPrivatelink_Basic(t *testing.T) {
	addonID := testAccConfig.GetAddonIDorSkip(t)
	allowedAccounts := "\"123456789123\", \"123456789124\""
	allowedAccountsUpdatedd := "\"123456789125\", \"123456789124\""

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPrivatelink_basic(addonID, allowedAccounts),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_privatelink.foobar", "addon_id", addonID),
					resource.TestCheckResourceAttr(
						"herokux_privatelink.foobar", "allowed_accounts.#", "2"),
					helper.TestCheckTypeSetElemAttr("herokux_privatelink.foobar",
						"allowed_accounts.*", "123456789123"),
					helper.TestCheckTypeSetElemAttr("herokux_privatelink.foobar",
						"allowed_accounts.*", "123456789124"),
					resource.TestCheckResourceAttr(
						"herokux_privatelink.foobar", "status", "Operational"),
					resource.TestCheckResourceAttrSet(
						"herokux_privatelink.foobar", "service_name"),
				),
			},
			{
				Config: testAccCheckHerokuxPrivatelink_basic(addonID, allowedAccountsUpdatedd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_privatelink.foobar", "addon_id", addonID),
					resource.TestCheckResourceAttr(
						"herokux_privatelink.foobar", "allowed_accounts.#", "2"),
					helper.TestCheckTypeSetElemAttr("herokux_privatelink.foobar",
						"allowed_accounts.*", "123456789125"),
					helper.TestCheckTypeSetElemAttr("herokux_privatelink.foobar",
						"allowed_accounts.*", "123456789124"),
					resource.TestCheckResourceAttr(
						"herokux_privatelink.foobar", "status", "Operational"),
					resource.TestCheckResourceAttrSet(
						"herokux_privatelink.foobar", "service_name"),
				),
			},
		},
	})
}

func testAccCheckHerokuxPrivatelink_basic(addonID, allowedAccounts string) string {
	return fmt.Sprintf(`
resource "herokux_privatelink" "foobar" {
	addon_id = "%s"
	allowed_accounts = [%s]
}
`, addonID, allowedAccounts)
}
