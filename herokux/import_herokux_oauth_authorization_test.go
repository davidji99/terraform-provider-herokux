package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccHerokuxOauthAuthorization_importCustomAPIKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxOauthAuthorization_CustomAPIKey(),
			},
			{
				ResourceName: "herokux_oauth_authorization.foobar",
				ImportStateIdFunc: testAccHerokuxOauthAuthorizationImportStateIDFunc(
					"herokux_oauth_authorization.foobar", "100000", "TESTACC"),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"expires_in"},
			},
		},
	})
}

func TestAccHerokuxOauthAuthorization_importNoKeyName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxOauthAuthorization_basic(),
			},
			{
				ResourceName: "herokux_oauth_authorization.foobar",
				ImportStateIdFunc: testAccHerokuxOauthAuthorizationImportStateIDFunc(
					"herokux_oauth_authorization.foobar", "100000", ""),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"expires_in"},
			},
		},
	})
}

func testAccHerokuxOauthAuthorizationImportStateIDFunc(resourceName string, ttl, keyName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("[ERROR] Not found: %s", resourceName)
		}

		importID := fmt.Sprintf("%s:%s:%s", rs.Primary.Attributes["id"], ttl, keyName)
		if keyName == "" {
			importID = fmt.Sprintf("%s:%s", rs.Primary.Attributes["id"], ttl)
		}

		return importID, nil
	}
}
