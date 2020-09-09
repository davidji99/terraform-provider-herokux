package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxKafkaTopic_Basic(t *testing.T) {
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)
	topicName := fmt.Sprintf("tftest-%s", acctest.RandString(15))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxKafkaTopic_basic(kafkaID, topicName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "kafka_id", kafkaID),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "name", topicName),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "partitions", "8"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "replication_factor", "3"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "retention_time", "2d"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "compaction", "true"),
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_topic.foobar", "status"),
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_topic.foobar", "cleanup_policy"),
				),
			},
		},
	})
}

func TestAccHerokuxKafkaTopic_Simple(t *testing.T) {
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)
	topicName := fmt.Sprintf("tftest-%s", acctest.RandString(15))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxKafkaTopic_NoRetentionReplicationSpecified(kafkaID, topicName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "kafka_id", kafkaID),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "name", topicName),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "partitions", "8"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "replication_factor", "3"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "retention_time", "1d"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "compaction", "false"),
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_topic.foobar", "status"),
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_topic.foobar", "cleanup_policy"),
				),
			},
		},
	})
}

func TestAccHerokuxKafkaTopic_UpdatePlan(t *testing.T) {
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)
	topicName := fmt.Sprintf("tftest-%s", acctest.RandString(15))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxKafkaTopic_basic(kafkaID, topicName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "kafka_id", kafkaID),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "name", topicName),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "partitions", "8"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "replication_factor", "3"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "retention_time", "2d"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "compaction", "true"),
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_topic.foobar", "status"),
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_topic.foobar", "cleanup_policy"),
				),
			},
			{
				Config: testAccCheckHerokuxKafkaTopic_updated(kafkaID, topicName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "kafka_id", kafkaID),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "name", topicName),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "partitions", "8"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "replication_factor", "3"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "retention_time", "95h"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "compaction", "false"),
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_topic.foobar", "status"),
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_topic.foobar", "cleanup_policy"),
				),
			},
		},
	})
}

func TestAccHerokuxKafkaTopic_DisableRetention(t *testing.T) {
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)
	topicName := fmt.Sprintf("tftest-%s", acctest.RandString(15))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxKafkaTopic_basic(kafkaID, topicName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "kafka_id", kafkaID),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "name", topicName),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "partitions", "8"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "replication_factor", "3"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "retention_time", "2d"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "compaction", "true"),
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_topic.foobar", "status"),
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_topic.foobar", "cleanup_policy"),
				),
			},
			{
				Config: testAccCheckHerokuxKafkaTopic_retentionDisabled(kafkaID, topicName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "kafka_id", kafkaID),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "name", topicName),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "partitions", "8"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "replication_factor", "3"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "retention_time", "disable"),
					resource.TestCheckResourceAttr(
						"herokux_kafka_topic.foobar", "compaction", "true"),
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_topic.foobar", "status"),
					resource.TestCheckResourceAttrSet(
						"herokux_kafka_topic.foobar", "cleanup_policy"),
				),
			},
		},
	})
}

func testAccCheckHerokuxKafkaTopic_basic(kafkaID, name string) string {
	return fmt.Sprintf(`
resource "herokux_kafka_topic" "foobar" {
	kafka_id = "%s"
	name = "%s"
	partitions = 8
	replication_factor = 3
	retention_time = "2d"
	compaction = true
}
`, kafkaID, name)
}

func testAccCheckHerokuxKafkaTopic_NoRetentionReplicationSpecified(kafkaID, name string) string {
	return fmt.Sprintf(`
resource "herokux_kafka_topic" "foobar" {
	kafka_id = "%s"
	name = "%s"
	partitions = 8
}
`, kafkaID, name)
}

func testAccCheckHerokuxKafkaTopic_updated(kafkaID, name string) string {
	return fmt.Sprintf(`
resource "herokux_kafka_topic" "foobar" {
	kafka_id = "%s"
	name = "%s"
	partitions = 8
	replication_factor = 3
	retention_time = "95h"
	compaction = false
}
`, kafkaID, name)
}

func testAccCheckHerokuxKafkaTopic_retentionDisabled(kafkaID, name string) string {
	return fmt.Sprintf(`
resource "herokux_kafka_topic" "foobar" {
	kafka_id = "%s"
	name = "%s"
	partitions = 8
	replication_factor = 3
	retention_time = "disable"
	compaction = true
}
`, kafkaID, name)
}
