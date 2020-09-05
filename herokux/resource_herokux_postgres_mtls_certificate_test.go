package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresMTLSCertificate_Basic(t *testing.T) {
	dbName := testAccConfig.GetDBNameorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresMTLSCertificate_basic(dbName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_postgres_mtls.foobar", "database_name", dbName),
					resource.TestCheckResourceAttr(
						"herokux_postgres_mtls.foobar", "status", "Operational"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_mtls_certificate.foobar", "database_name", dbName),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_mtls_certificate.foobar", "name"),
					resource.TestCheckResourceAttr(
						"herokux_postgres_mtls_certificate.foobar", "status", "ready"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_mtls_certificate.foobar", "expiration_date"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_mtls_certificate.foobar", "private_key"),
					resource.TestCheckResourceAttrSet(
						"herokux_postgres_mtls_certificate.foobar", "certificate_with_chain"),
				),
			},
		},
	})
}

func testAccCheckHerokuxPostgresMTLSCertificate_basic(dbName string) string {
	return fmt.Sprintf(`
resource "herokux_postgres_mtls" "foobar" {
	database_name = "%[1]s"
}

resource "herokux_postgres_mtls_certificate" "foobar" {
	database_name = herokux_postgres_mtls.foobar.database_name
}
`, dbName)
}
