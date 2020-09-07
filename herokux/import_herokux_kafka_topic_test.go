package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxKafkaTopic_importBasic(t *testing.T) {
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)
	topicName := fmt.Sprintf("tftest-%s", acctest.RandString(15))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxKafkaTopic_basic(kafkaID, topicName),
			},
			{
				ResourceName:      "herokux_kafka_topic.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
