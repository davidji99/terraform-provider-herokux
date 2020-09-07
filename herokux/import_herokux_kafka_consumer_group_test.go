package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxKafkaConsumerGroup_importBasic(t *testing.T) {
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)
	groupName := fmt.Sprintf("tftest-%s", acctest.RandString(15))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxKafkaConsumerGroup_basic(kafkaID, groupName),
			},
			{
				ResourceName:      "herokux_kafka_consumer_group.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
