package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresCredential_Basic(t *testing.T) {
	postgresID := testAccConfig.GetAddonIDorSkip(t)
	name := fmt.Sprintf("tftest-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresCredential_basic(postgresID, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_credential.foobar", "postgres_id", postgresID),
					resource.TestCheckResourceAttr(
						"herokux_postgres_credential.foobar", "name", name),
					resource.TestCheckResourceAttr(
						"herokux_postgres_credential.foobar", "state", "active"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_credential.foobar", "database"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_credential.foobar", "host"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_credential.foobar", "port"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_credential.foobar", "uuid"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_credential.foobar", "secrets.#", "1"),
				),
			},
		},
	})
}

func TestAccHerokuxPostgresCredential_E2E_PremiumPG(t *testing.T) {
	testAccConfig.GetRunE2ETestsOrSkip(t)

	orgName := testAccConfig.GetAnyOrganizationOrSkip(t)
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	plan := "premium-0"
	credName := fmt.Sprintf("pgcredtest-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		Providers:         testAccProviders,
		ExternalProviders: externalProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresCredential_basicWithHerokuResource(appName, orgName, plan, credName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_credential.foobar", "postgres_id"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_credential.foobar", "name", credName),
					resource.TestCheckResourceAttr(
						"herokux_postgres_credential.foobar", "state", "active"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_credential.foobar", "database"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_credential.foobar", "host"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_credential.foobar", "port"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_credential.foobar", "uuid"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_credential.foobar", "secrets.#", "1"),
				),
			},
		},
	})
}

func testAccCheckHerokuxPostgresCredential_basic(postgresID, name string) string {
	return fmt.Sprintf(`
resource "herokux_postgres_credential" "foobar" {
	postgres_id = "%s"
	name = "%s"
}
`, postgresID, name)
}

func testAccCheckHerokuxPostgresCredential_basicWithHerokuResource(appName, orgName, addonPlan, name string) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    heroku = {
      source = "heroku/heroku"
      version = ">= 4.0"
    }
  }
}

provider "heroku" {
}

resource "heroku_app" "foobar" {
  name   = "%s"
  region = "us"

  organization {
    name = "%s"
  }
}

resource "heroku_addon" "database" {
  app  = heroku_app.foobar.name
  plan = "heroku-postgresql:%s"
}

resource "herokux_postgres_credential" "foobar" {
  postgres_id = heroku_addon.database.id
  name = "%s"
}
`, appName, orgName, addonPlan, name)
}
