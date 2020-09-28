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

func testAccCheckHerokuxPostgresCredential_basic(postgresID, name string) string {
	return fmt.Sprintf(`
resource "herokux_postgres_credential" "foobar" {
	postgres_id = "%s"
	name = "%s"
}
`, postgresID, name)
}
