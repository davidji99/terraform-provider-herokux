package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccHerokuxPostgresDataclip_importBasic(t *testing.T) {
	attachmentID := testAccConfig.GetAttachmentIDorSkip(t)
	title := fmt.Sprintf("tftest_%s", acctest.RandString(10))
	sql := "select * from pg_catalog"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresDataclip_basic(attachmentID, title, sql, "true"),
			},
			{
				ResourceName:      "herokux_postgres_dataclip.foobar",
				ImportStateIdFunc: testAccHerokuxPostgresDataclipImportStateIDFunc("herokux_postgres_dataclip.foobar"),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccHerokuxPostgresDataclipImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("[ERROR] Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["slug"], nil
	}
}
