package herokux

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresMTLS_importBasic(t *testing.T) {
	dbName := testAccConfig.GetDBNameorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresMTLS_basic(dbName),
			},
			{
				ResourceName:      "herokux_formation_autoscaling.foobar",
				ImportStateId:     dbName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
