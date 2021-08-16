package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccHerokuxPostgresDataclipUserAssociation_importBasic(t *testing.T) {
	attachmentID := testAccConfig.GetAttachmentIDorSkip(t)
	title := fmt.Sprintf("tftest_%s", acctest.RandString(10))
	email := testAccConfig.GetUserEmailorSkip(t)
	sql := "select * from pg_catalog"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresDataclipUserAssociation_basic(attachmentID, title, sql, email),
			},
			{
				ResourceName: "herokux_postgres_dataclip_user_association.user",
				ImportStateIdFunc: testAccHerokuxPostgresDataclipUserAssociationImportStateIDFunc(
					"herokux_postgres_dataclip_user_association.user", email),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccHerokuxPostgresDataclipUserAssociationImportStateIDFunc(resourceName, email string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("[ERROR] Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["dataclip_slug"], email), nil
	}
}
