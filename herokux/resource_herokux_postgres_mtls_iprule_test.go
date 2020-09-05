package herokux

import (
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/helper/test"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresMTLSIPRule_Basic(t *testing.T) {
	dbName := testAccConfig.GetDBNameorSkip(t)
	cidr := test.GenerateRandomCIDR()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresMTLSIPRule_basic(dbName, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_mtls.foobar", "database_name", dbName),
					resource.TestCheckResourceAttr(
						"herokux_postgres_mtls.foobar", "status", "Operational"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_mtls_iprule.foobar", "database_name", dbName),
					resource.TestCheckResourceAttr(
						"herokux_postgres_mtls_iprule.foobar", "cidr", cidr),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_mtls_iprule.foobar", "rule_id"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_mtls_iprule.foobar", "description", "this is a test IP rule"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_mtls_iprule.foobar", "status", "Authorized"),
				),
			},
		},
	})
}

func testAccCheckHerokuxPostgresMTLSIPRule_basic(dbName, cidr string) string {
	return fmt.Sprintf(`
resource "herokux_postgres_mtls" "foobar" {
	database_name = "%s"
}

resource "herokux_postgres_mtls_iprule" "foobar" {
	database_name = herokux_postgres_mtls.foobar.database_name
	cidr = "%s"
	description = "this is a test IP rule"
}
`, dbName, cidr)
}
