package herokux

import (
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/helper/test"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxKafkaMTLSIPRule_Basic(t *testing.T) {
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)
	cidr := test.GenerateRandomCIDR()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxKafkaMTLSIPRule_basic(kafkaID, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_kafka_mtls_iprule.foobar", "kafka_id", kafkaID),
					resource.TestCheckResourceAttr(
						"herokux_kafka_mtls_iprule.foobar", "cidr", cidr),
					resource.TestCheckResourceAttr(
						"herokux_kafka_mtls_iprule.foobar", "description", "this is a test IP rule created for terraform provider testing"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_mtls_iprule.foobar", "status", "Authorized"),
				),
			},
		},
	})
}

func testAccCheckHerokuxKafkaMTLSIPRule_basic(kafkaID, cidr string) string {
	return fmt.Sprintf(`
resource "herokux_kafka_mtls_iprule" "foobar" {
	kafka_id = "%s"
	cidr = "%s"
	description = "this is a test IP rule created for terraform provider testing"
}
`, kafkaID, cidr)
}
