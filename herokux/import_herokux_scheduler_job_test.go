package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccHerokuxSchedulerJob_importBasic(t *testing.T) {
	orgName := testAccConfig.GetAnyOrganizationOrSkip(t)
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	plan := "scheduler:standard"
	frequency := "every_day_at_17:30"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ExternalProviders: externalProviders(),
		Providers:         testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxSchedulerJob_basicWithHerokuResource(appName, orgName, plan, "test", "Standard-1X", frequency),
			},
			{
				ResourceName:      "herokux_scheduler_job.foobar",
				ImportStateIdFunc: testAccHerokuxSchedulerJobImportStateIDFunc("herokux_scheduler_job.foobar"),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccHerokuxSchedulerJobImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("[ERROR] Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["app_id"], rs.Primary.Attributes["id"]), nil
	}
}
