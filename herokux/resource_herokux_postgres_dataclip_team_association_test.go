package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresDataclipTeamAssociation_Basic(t *testing.T) {
	attachmentID := testAccConfig.GetAttachmentIDorSkip(t)
	teamID := testAccConfig.GetTeamIDorSkip(t)
	title := fmt.Sprintf("tftest_%s", acctest.RandString(10))
	sql := "select * from pg_catalog"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresDataclipTeamAssociation_basic(attachmentID, title, sql, teamID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_dataclip_team_association.team", "team_id", teamID),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip_team_association.team", "dataclip_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip_team_association.team", "dataclip_slug"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip_team_association.team", "team_name"),
				),
			},
		},
	})
}

func testAccCheckHerokuxPostgresDataclipTeamAssociation_basic(attachmentID, title, sql, teamID string) string {
	return fmt.Sprintf(`
resource "herokux_postgres_dataclip" "dataclip" {
	postgres_attachment_id = "%s"
	title = "%s"
	sql = "%s"
	enable_shareable_links = true
}

resource "herokux_postgres_dataclip_team_association" "team" {
	dataclip_id = herokux_postgres_dataclip.dataclip.id
	dataclip_slug = herokux_postgres_dataclip.dataclip.slug
	team_id = "%s"
}
`, attachmentID, title, sql, teamID)
}
