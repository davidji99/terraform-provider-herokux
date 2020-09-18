package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPrivatelink_importBasic(t *testing.T) {
	addonID := testAccConfig.GetAddonIDorSkip(t)
	allowedAccounts := "\"123456789123\", \"123456789124\""

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPrivatelink_basic(addonID, allowedAccounts),
			},
			{
				ResourceName:      "herokux_privatelink.foobar",
				ImportStateId:     fmt.Sprintf("%s", addonID),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
