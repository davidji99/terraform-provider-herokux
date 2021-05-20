package herokux

import (
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/helper/test"
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

func TestAccE2EHerokuxKafkaConsumerGroup(t *testing.T) {
	testAccConfig.GetRunE2ETestsOrSkip(t)

	orgName := testAccConfig.GetAnyOrganizationOrSkip(t)
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	plan := "heroku-kafka:standard-0"
	groupName := fmt.Sprintf("tftest-%s", acctest.RandString(15))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		Providers:         testAccProviders,
		ExternalProviders: externalProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxKafkaConsumerGroup_basicWithHerokuResource(appName, orgName, plan, groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_consumer_group.foobar", "kafka_id"),
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

func testAccCheckHerokuxKafkaConsumerGroup_basicWithHerokuResource(appName, orgName, plan, groupName string) string {
	return fmt.Sprintf(`
%s

resource "herokux_kafka_consumer_group" "foobar" {
	kafka_id = heroku_addon.foobar.id
	name = "%s"
}
`, test.HerokuAppAddonBlock(appName, orgName, plan), groupName)
}
