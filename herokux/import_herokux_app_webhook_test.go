package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccHerokuxAppWebhook_importBasic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	name := fmt.Sprintf("tftest-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAppWebhook_CustomNameNoSecret(appID, name),
			},
			{
				ResourceName:            "herokux_app_webhook.foobar",
				ImportStateIdFunc:       testAccHerokuxAppWebhookImportStateIDFunc("herokux_app_webhook.foobar"),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"signing_secret", "secret"},
			},
		},
	})
}

func testAccHerokuxAppWebhookImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("[ERROR] Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["id"], nil
	}
}
