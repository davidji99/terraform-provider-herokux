package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxFormationAutoscaling_importBasic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	formationName := "web"
	minQuantity := acctest.RandIntRange(1, 8)
	maxQuantity := minQuantity + 2
	p95ResponseTime := acctest.RandIntRange(500, 1000)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxFormationAutoscaling_basic(appID, formationName, minQuantity, maxQuantity, p95ResponseTime),
			},
			{
				ResourceName:      "herokux_formation_autoscaling.foobar",
				ImportStateId:     fmt.Sprintf("%s:%s", appID, formationName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
