package herokux

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresMTLSCertificate_importBasic(t *testing.T) {
	dbName := testAccConfig.GetDBNameorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresMTLSCertificate_basic(dbName),
			},
			{
				ResourceName:      "herokux_postgres_mtls_certificate.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
