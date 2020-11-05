package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxShieldPrivateSpace_importBasic(t *testing.T) {
	name := fmt.Sprintf("tftest-%s", acctest.RandString(15))
	teamID := testAccConfig.GetTeamIDorSkip(t)
	url := "https://somename:somesecret@loghost.example.com/logpath"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxShieldPrivateSpace_basic(name, teamID, url),
			},
			{
				ResourceName:      "herokux_shield_private_space.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
