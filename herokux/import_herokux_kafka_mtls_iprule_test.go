package herokux

import (
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/helper/test"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxKafkaMTLSIPrule_importBasic(t *testing.T) {
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)
	cidr := test.GenerateRandomCIDR()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxKafkaMTLSIPRule_basic(kafkaID, cidr),
			},
			{
				ResourceName:      "herokux_kafka_mtls_iprule.foobar",
				ImportStateId:     fmt.Sprintf("%s:%s", kafkaID, cidr),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
