package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxImportFormationAlertLatency_Basic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	processType := "web"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxFormationAlert_basic(appID, processType, "LATENCY",
					"1202", 10, 1440),
			},
			{
				ResourceName:      "herokux_formation_alert.foobar",
				ImportStateId:     fmt.Sprintf("%s:%s:%s", appID, processType, "LATENCY"),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccHerokuxImportFormationAlertErrorRate_Basic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	processType := "web"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxFormationAlert_basic(appID, processType, "ERROR_RATE",
					"0.042", 5, 60),
			},
			{
				ResourceName:      "herokux_formation_alert.foobar",
				ImportStateId:     fmt.Sprintf("%s:%s:%s", appID, processType, "ERROR_RATE"),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
