package herokux

import (
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/helper/test"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"strings"
	"testing"
)

func TestAccHerokuxPostgresConnectionPooling_Basic(t *testing.T) {
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	name := strings.ToUpper(fmt.Sprintf("tftest_%s", acctest.RandString(10)))
	orgName := testAccConfig.GetAnyOrganizationOrSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		Providers:         testAccProviders,
		ExternalProviders: externalProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresConnectionPooling_basic(orgName, appName, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_connection_pooling.foobar", "name", name),
					resource.TestCheckResourceAttr(
						"herokux_postgres_connection_pooling.foobar", "config_var", fmt.Sprintf("%s_URL", name)),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_connection_pooling.foobar", "app_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_connection_pooling.foobar", "postgres_id"),
				),
			},
		},
	})
}

func TestAccHerokuxPostgresConnectionPooling_NoName(t *testing.T) {
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	orgName := testAccConfig.GetAnyOrganizationOrSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		Providers:         testAccProviders,
		ExternalProviders: externalProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresConnectionPooling_NoName(orgName, appName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_connection_pooling.foobar", "name", "DATABASE_CONNECTION_POOL"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_connection_pooling.foobar", "config_var", fmt.Sprintf("%s_URL", "DATABASE_CONNECTION_POOL")),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_connection_pooling.foobar", "app_id"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_connection_pooling.foobar", "postgres_id"),
				),
			},
		},
	})
}

func TestAccHerokuxPostgresConnectionPooling_Invalid(t *testing.T) {
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	name := "0test-name"
	orgName := testAccConfig.GetAnyOrganizationOrSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		Providers:         testAccProviders,
		ExternalProviders: externalProviders(),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckHerokuxPostgresConnectionPooling_basic(orgName, appName, name),
				ExpectError: regexp.MustCompile(`must start with a letter and can only contain uppercase letters, numbers, and underscores`),
			},
		},
	})
}

func testAccCheckHerokuxPostgresConnectionPooling_basic(orgName, appName, name string) string {
	return fmt.Sprintf(`
%s

resource "herokux_postgres_connection_pooling" "foobar" {
	postgres_id = heroku_addon.foobar.id
	app_id = heroku_app.foobar.uuid
	name = "%s"
}
`, test.HerokuAppAddonBlock(appName, orgName, "heroku-postgresql:standard-0"), name)
}

func testAccCheckHerokuxPostgresConnectionPooling_NoName(orgName, appName string) string {
	return fmt.Sprintf(`
%s

resource "herokux_postgres_connection_pooling" "foobar" {
	postgres_id = heroku_addon.foobar.id
	app_id = heroku_app.foobar.uuid
}
`, test.HerokuAppAddonBlock(appName, orgName, "heroku-postgresql:standard-0"))
}
