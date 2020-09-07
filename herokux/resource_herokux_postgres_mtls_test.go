package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresMTLS_Basic(t *testing.T) {
	dbName := testAccConfig.GetDBNameorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresMTLS_basic(dbName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_mtls.foobar", "database_name", dbName),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_mtls.foobar", "app_name"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_mtls.foobar", "status", "Operational"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_mtls.foobar", "enabled_by"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_mtls.foobar", "certificate_authority_chain"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_mtls.foobar", "initial_certificate_id"),
				),
			},
		},
	})
}

func testAccCheckHerokuxPostgresMTLS_basic(dbName string) string {
	return fmt.Sprintf(`
resource "herokux_postgres_mtls" "foobar" {
	database_name = "%s"
}
`, dbName)
}
