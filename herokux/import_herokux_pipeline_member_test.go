package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccHerokuxPipelineMember_importBasic(t *testing.T) {
	pipelineID := testAccConfig.GetPipelineIDorSkip(t)
	email := testAccConfig.GetUserEmailorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPipelineMember_basic(pipelineID, email,
					"\"view\", \"deploy\", \"operate\", \"manage\""),
			},
			{
				ResourceName:      "herokux_pipeline_member.foobar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccHerokuxPipelineMemberImportStateIDFunc(
					"herokux_pipeline_member.foobar"),
			},
		},
	})
}

func testAccHerokuxPipelineMemberImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("[ERROR] Not found: %s", resourceName)
		}

		importID := fmt.Sprintf("%s:%s", rs.Primary.Attributes["pipeline_id"], rs.Primary.Attributes["email"])

		return importID, nil
	}
}
