---
layout: "herokux"
page_title: "HerokuX: herokux_kafka_topic"
sidebar_current: "docs-herokux-resource-kafka-topic"
description: |-
  Provides a resource to manage a Kafka topic
---

# herokux\_kafka\_topic

This resource manages topics in an existing Heroku Kafka instance.

-> **IMPORTANT!**
Design Kafka topics carefully. Parameters like retention or compaction can be changed relatively easily,
and replication can be changed with some additional care, but partitions CANNOT currently be changed after creation.
Compaction and time-based retention are mutually exclusive configurations for a given topic,
though different topics within a cluster may have a mix of these configurations.

### Resources
Configuring Kafka topics is very dependent on your Kafka addon plan.
Please refer to the following documentation when deciding how to configure a topic within the plan's limitations:

* [Apache Kafka on Heroku](https://devcenter.heroku.com/articles/kafka-on-heroku)
* [Apache Kafka on Heroku Plan Details](https://elements.heroku.com/addons/heroku-kafka)
* [Multi-Tenant Apache Kafka on Heroku](https://devcenter.heroku.com/articles/multi-tenant-kafka-on-heroku#basic-plans)
* [Apache Kafka on Heroku Add-on Migration](https://devcenter.heroku.com/articles/kafka-addon-migration)

### Resource Timeouts
This resource checks the status of a creation or update action.
Both checks' default timeout is 10 minutes, which can be customized via the
`timeouts.kafka_topic_create_verify_timeout` and `timeouts.kafka_topic_update_verify_timeout` attributes in your `provider` block.

For example:

```hcl-terraform
provider "herokux" {
  timeouts {
    kafka_topic_create_verify_timeout = 15
    kafka_topic_update_verify_timeout = 15
  }
}
```

## Example Usage

```hcl-terraform
resource "heroku_app" "foobar" {
  name   = "my_foobar_app"
  region = "us"

  organization {
    name = "my_org"
  }
}

resource "heroku_addon" "kafka" {
  app  = heroku_app.foobar.name
  plan = "heroku-kafka:standard-0"
}

resource "herokux_kafka_topic" "foobar" {
  kafka_id = heroku_addon.kafka.id
  name = "my-cool-topic"
  partitions = 8
  replication_factor = 3
  retention_time = "2d"
  compaction = true
}
```

## Argument Reference

The following arguments are supported:

* `kafka_id` - (Required) `<string>` The UUID of an existing Kafka instance.
* `name` - (Required) `<string>` The name of the topic. Alphanumeric characters, periods, underscores, and hyphens.
Immutable after creation.
* `partitions` - (Required) `<integer>` Number of partitions. Partitions are discrete subsets of a topic used to
balance the concerns of parallelism and ordering. Increased numbers of partitions can increase the number
of producers and consumers that can work on a given topic, increasing parallelism and throughput.
* `replication_factor` - (Optional) `<integer>` The replication factor for the topic. The default & minimum value is 3.
The upper limit is the number of brokers available for your Kafka plan.
* `retention_time` - (Optional) `<string>` How long to keep messages before they are cleaned up and removed.
Please note the following:
    * Default and minimum value is "1d" or equivalent in other units of duration. Each Heroku Kafka plan has different maximum retention times.
    * Acceptable values follow this format: `<NUMERICAL_DIGITS><ms|s|m|h|d|w>`. For example:
        * "6w" is six weeks.
        * "13d" is thirteen days.
        * "12000m" is twelve thousand minutes.
    * If using a retention time that can be expressed in two different units of duration, please use the larger unit of duration.
    For example, you must use "2w" over "14d".
    * Depending on the Kafka plan, to disable retention time, specify "disable' as this attribute's value.
    * `retention_time` is required when `compaction` is disabled. Retention time must be set for multi-tenanted plans.
* `compaction` - (Optional) `<boolean>` Enable log compaction. This configuration changes the semantics of a topic such
that it keeps only the most recent message for a given key, tombstoning any predecessor.
This allows for the creation of a value-stream, or table-like view of data,
and is a very powerful construct in modeling your data and systems. Defaults is `false`.

## Attributes Reference

The following attributes are exported:

* `status` - (Optional) `<string>` Status of the topic.
* `cleanup_policy` - (Optional) `<string>` The current cleanup policy for the topic.

## Import

An existing topic can be imported using a composite value of the Kafka ID and topic name
separated by a colon.

For example:

```shell script
$ terraform import herokux_kafka_topic.foobar "11db7126-0cb7-4b42-a64a-d4ae70110216:my-cool-topic"
```
