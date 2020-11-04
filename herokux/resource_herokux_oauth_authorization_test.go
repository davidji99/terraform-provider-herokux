package herokux

import (
	"fmt"
	helper "github.com/davidji99/terraform-provider-herokux/helper/test"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxOauthAuthorization_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxOauthAuthorization_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"herokux_oauth_authorization.foobar", "access_token"),
					resource.TestCheckResourceAttrSet(
						"herokux_oauth_authorization.foobar", "expires_in"),
					resource.TestCheckResourceAttrSet(
						"herokux_oauth_authorization.foobar", "token_id"),
					resource.TestCheckResourceAttr(
						"herokux_oauth_authorization.foobar", "description", "This is an oauth authorization test from Terraform"),
					helper.TestCheckTypeSetElemAttr("herokux_oauth_authorization.foobar",
						"scope.*", "read"),
				),
			},
		},
	})
}

func TestAccHerokuxOauthAuthorization_CustomAPIKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxOauthAuthorization_CustomAPIKey(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"herokux_oauth_authorization.foobar", "access_token"),
					resource.TestCheckResourceAttrSet(
						"herokux_oauth_authorization.foobar", "expires_in"),
					resource.TestCheckResourceAttrSet(
						"herokux_oauth_authorization.foobar", "token_id"),
					resource.TestCheckResourceAttr(
						"herokux_oauth_authorization.foobar", "description", "This is an oauth authorization test from Terraform"),
					helper.TestCheckTypeSetElemAttr("herokux_oauth_authorization.foobar",
						"scope.*", "read"),
				),
			},
		},
	})
}

func TestAccHerokuxOauthAuthorization_NoTTL(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxOauthAuthorization_NoTTL(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"herokux_oauth_authorization.foobar", "access_token"),
					resource.TestCheckResourceAttr(
						"herokux_oauth_authorization.foobar", "expires_in", "0"),
					resource.TestCheckResourceAttrSet(
						"herokux_oauth_authorization.foobar", "token_id"),
					resource.TestCheckResourceAttr(
						"herokux_oauth_authorization.foobar", "description", "This is an oauth authorization test from Terraform"),
					helper.TestCheckTypeSetElemAttr("herokux_oauth_authorization.foobar",
						"scope.*", "read"),
				),
			},
		},
	})
}

func testAccCheckHerokuxOauthAuthorization_basic() string {
	return `
resource "herokux_oauth_authorization" "foobar" {
	scope = ["read"]
	time_to_live = 100000
	description = "This is an oauth authorization test from Terraform"
}
`
}

func testAccCheckHerokuxOauthAuthorization_CustomAPIKey() string {
	return `
resource "herokux_oauth_authorization" "foobar" {
	scope = ["read"]
	auth_api_key_name = "TESTACC"
	time_to_live = 100000
	description = "This is an oauth authorization test from Terraform"
}
`
}

func testAccCheckHerokuxOauthAuthorization_NoTTL() string {
	return fmt.Sprintf(`
resource "herokux_oauth_authorization" "foobar" {
	scope = ["read"]
	auth_api_key_name = "TESTACC"
	description = "This is an oauth authorization test from Terraform"
}
`)
}
