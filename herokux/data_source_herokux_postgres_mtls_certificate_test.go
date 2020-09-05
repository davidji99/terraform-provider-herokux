package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDatasourceHerokuPostgresMTLSCertificate_Basic(t *testing.T) {
	dbName := testAccConfig.GetDBNameorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuTeamWithDataSource_Basic(dbName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.herokux_postgres_mtls_certificate.foobar", "database_name", dbName),
					resource.TestCheckResourceAttrSet(
						"data.herokux_postgres_mtls_certificate.foobar", "name"),
					resource.TestCheckResourceAttr(
						"data.herokux_postgres_mtls_certificate.foobar", "status", "ready"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_postgres_mtls_certificate.foobar", "expiration_date"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_postgres_mtls_certificate.foobar", "private_key"),
					resource.TestCheckResourceAttrSet(
						"data.herokux_postgres_mtls_certificate.foobar", "certificate_with_chain"),
				),
			},
		},
	})
}

func testAccCheckHerokuTeamWithDataSource_Basic(dbName string) string {
	return fmt.Sprintf(`
resource "herokux_postgres_mtls" "foobar" {
	database_name = "%[1]s"
}

resource "herokux_postgres_mtls_certificate" "foobar" {
	database_name = herokux_postgres_mtls.foobar.database_name
}

data "herokux_postgres_mtls_certificate" "foobar" {
  database_name = herokux_postgres_mtls.foobar.database_name
  cert_id = herokux_postgres_mtls_certificate.foobar.cert_id
}
`, dbName)
}
