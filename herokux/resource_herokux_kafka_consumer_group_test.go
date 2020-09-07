package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxKafkaConsumerGroup_Basic(t *testing.T) {
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)
	groupName := fmt.Sprintf("tftest-%s", acctest.RandString(15))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxKafkaConsumerGroup_basic(kafkaID, groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_kafka_consumer_group.foobar", "kafka_id", kafkaID),
					resource.TestCheckResourceAttr(
						"herokux_kafka_consumer_group.foobar", "name", groupName),
				),
			},
		},
	})
}

func testAccCheckHerokuxKafkaConsumerGroup_basic(kafkaID, name string) string {
	return fmt.Sprintf(`
resource "herokux_kafka_consumer_group" "foobar" {
	kafka_id = "%s"
	name = "%s"
}
`, kafkaID, name)
}
