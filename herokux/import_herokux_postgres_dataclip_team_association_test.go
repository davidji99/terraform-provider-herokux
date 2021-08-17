package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccHerokuxPostgresDataclipTeamAssociation_importBasic(t *testing.T) {
	attachmentID := testAccConfig.GetAttachmentIDorSkip(t)
	title := fmt.Sprintf("tftest_%s", acctest.RandString(10))
	teamID := testAccConfig.GetTeamIDorSkip(t)
	sql := "select * from pg_catalog"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresDataclipTeamAssociation_basic(attachmentID, title, sql, teamID),
			},
			{
				ResourceName: "herokux_postgres_dataclip_team_association.team",
				ImportStateIdFunc: testAccHerokuxPostgresDataclipTeamAssociationImportStateIDFunc(
					"herokux_postgres_dataclip_team_association.team"),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccHerokuxPostgresDataclipTeamAssociationImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("[ERROR] Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["dataclip_slug"], rs.Primary.Attributes["team_name"]), nil
	}
}
