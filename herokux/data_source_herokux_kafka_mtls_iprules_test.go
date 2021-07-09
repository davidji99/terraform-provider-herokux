package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDatasourceHerokuKafkaMTLSIPRules_BasicWithRules(t *testing.T) {
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxKafkaMTLSIPRules_Basic(kafkaID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.herokux_kafka_mtls_iprules.foobar", "kafka_id", kafkaID),
					resource.TestCheckResourceAttrSet(
						"data.herokux_kafka_mtls_iprules.foobar", "rules.0.id"),
					// The following resource attribute checks can be changed depending on the testing scenario.
					// I'm hardcoding values to avoid the need to create multiple rules simultaneously.
					resource.TestCheckResourceAttr(
						"data.herokux_kafka_mtls_iprules.foobar", "rules.#", "3"),
					resource.TestCheckResourceAttr(
						"data.herokux_kafka_mtls_iprules.foobar", "rules.0.status", "Authorized"),
					resource.TestCheckResourceAttr(
						"data.herokux_kafka_mtls_iprules.foobar", "rules.0.cidr", "1.5.5.2/32"),
					resource.TestCheckResourceAttr(
						"data.herokux_kafka_mtls_iprules.foobar", "rules.0.description", "this is a test IP rule created for terraform provider testing"),
				),
			},
		},
	})
}

func testAccCheckHerokuxKafkaMTLSIPRules_Basic(kafkaID string) string {
	return fmt.Sprintf(`
data "herokux_kafka_mtls_iprules" "foobar" {
  kafka_id = "%s"
}
`, kafkaID)
}
