package herokux

import (
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/helper/test"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresMTLSIPrule_importBasic(t *testing.T) {
	dbName := testAccConfig.GetDBNameorSkip(t)
	cidr := test.GenerateRandomCIDR()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresMTLSIPRule_basic(dbName, cidr),
			},
			{
				ResourceName:      "herokux_postgres_mtls_iprule.foobar",
				ImportStateId:     fmt.Sprintf("%s:%s", dbName, cidr),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

//func TestAccHerokuxPostgresMTLSIPrule_importUnknown(t *testing.T) {
//	dbName := testAccConfig.GetDBNameorSkip(t)
//	cidr := test.GenerateRandomCIDR()
//
//	resource.Test(t, resource.TestCase{
//		PreCheck: func() {
//			testAccPreCheck(t)
//		},
//		Providers: testAccProviders,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccCheckHerokuxPostgresMTLSIPRule_basic2(dbName, cidr),
//			},
//			{
//				ResourceName:  "herokux_postgres_mtls_iprule.foobar",
//				ImportStateId: fmt.Sprintf("%s:%s", dbName, "0.0.0.0/32"),
//				ImportState:   true,
//				ExpectError:   regexp.MustCompile(`.*no existing IP rule found with CIDR.*`),
//			},
//		},
//	})
//}
