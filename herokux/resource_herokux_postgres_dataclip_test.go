package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresDataclip_Basic(t *testing.T) {
	attachmentID := testAccConfig.GetAttachmentIDorSkip(t)
	title := fmt.Sprintf("tftest_%s", acctest.RandString(10))
	sql := "select * from pg_catalog"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresDataclip_basic(attachmentID, title, sql, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_dataclip.foobar", "postgres_attachment_id", attachmentID),
					resource.TestCheckResourceAttr(
						"herokux_postgres_dataclip.foobar", "title", title),
					resource.TestCheckResourceAttr(
						"herokux_postgres_dataclip.foobar", "sql", sql),
					resource.TestCheckResourceAttr(
						"herokux_postgres_dataclip.foobar", "enable_shareable_links", "true"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "slug"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "creator_email"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "attachment_name"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "addon_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "addon_name"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "app_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "app_name"),
				),
			},
			{
				Config: testAccCheckHerokuxPostgresDataclip_basic(attachmentID, title+" edited", sql+" edited", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_dataclip.foobar", "postgres_attachment_id", attachmentID),
					resource.TestCheckResourceAttr(
						"herokux_postgres_dataclip.foobar", "title", title+" edited"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_dataclip.foobar", "sql", sql+" edited"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_dataclip.foobar", "enable_shareable_links", "false"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "slug"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "creator_email"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "attachment_name"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "addon_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "addon_name"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "app_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_dataclip.foobar", "app_name"),
				),
			},
		},
	})
}

func testAccCheckHerokuxPostgresDataclip_basic(attachmentID, title, sql, sharing string) string {
	return fmt.Sprintf(`
resource "herokux_postgres_dataclip" "foobar" {
	postgres_attachment_id = "%s"
	title = "%s"
	sql = "%s"
	enable_shareable_links = %s
}
`, attachmentID, title, sql, sharing)
}
