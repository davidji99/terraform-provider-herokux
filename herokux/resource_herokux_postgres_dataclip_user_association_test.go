package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresDataclipUserAssociation_Basic(t *testing.T) {
	attachmentID := testAccConfig.GetAttachmentIDorSkip(t)
	email := testAccConfig.GetUserEmailorSkip(t)
	title := fmt.Sprintf("tftest_%s", acctest.RandString(10))
	sql := "select * from pg_catalog"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresDataclipUserAssociation_basic(attachmentID, title, sql, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_dataclip_user_association.user", "email", email),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip_user_association.user", "shared_by_email"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip_user_association.user", "dataclip_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip_user_association.user", "dataclip_slug"),
				),
			},
		},
	})
}

func testAccCheckHerokuxPostgresDataclipUserAssociation_basic(attachmentID, title, sql, email string) string {
	return fmt.Sprintf(`
resource "herokux_postgres_dataclip" "dataclip" {
	postgres_attachment_id = "%s"
	title = "%s"
	sql = "%s"
	enable_shareable_links = true
}

resource "herokux_postgres_dataclip_user_association" "user" {
	dataclip_id = herokux_postgres_dataclip.dataclip.id
	dataclip_slug = herokux_postgres_dataclip.dataclip.slug
	email = "%s"
}
`, attachmentID, title, sql, email)
}
