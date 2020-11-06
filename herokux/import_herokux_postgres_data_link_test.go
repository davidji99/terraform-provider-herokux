package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxDataLink_importBasic(t *testing.T) {
	localID := testAccConfig.GetAddonIDorSkip(t)
	remoteName := testAccConfig.GetDBNameorSkip(t)
	name := fmt.Sprintf("tftest_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresDataLink_WithCustomName(localID, remoteName, name),
			},
			{
				ResourceName:      "herokux_postgres_data_link.foobar",
				ImportStateId:     fmt.Sprintf("%s:%s", localID, name),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
