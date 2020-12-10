package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxConnectMapping_importBasic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	connectID := testAccConfig.GetConnectIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxConnectMapping_basic(appID, connectID, TestConnectMappingBasic),
			},
			{
				ResourceName:      "herokux_connect_mapping.foobar",
				ImportStateId:     fmt.Sprintf("%s:%s", appID, connectID),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
