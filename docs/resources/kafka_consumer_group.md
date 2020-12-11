---
layout: "herokux"
page_title: "HerokuX: herokux_kafka_consumer_group"
sidebar_current: "docs-herokux-resource-kafka-consumer-group"
description: |-
  Provides a resource to manage a Kafka consumer group
---

# herokux\_kafka\_consumer\_group

This resource manages consumer groups in an existing Heroku Kafka instance.

### Resource Timeouts
This resource checks the status of the creation or deletion action.
Both checks' default timeout is 10 minutes, which can be customized via the
`timeouts.kafka_cg_create_timeout` and `timeouts.kafka_cg_delete_timeout` attributes in your `provider` block.

For example:

```hcl-terraform
provider "herokux" {
  timeouts {
    kafka_cg_create_timeout = 15
    kafka_cg_delete_timeout = 15
  }
}
```

## Example Usage

```hcl-terraform
resource "herokux_kafka_consumer_group" "foobar" {
	kafka_id = "2bccd770-e7aa-4865-98d2-6e222f2d2582"
	name = "my new group"
}
```

## Argument Reference

The following arguments are supported:

* `kafka_id` - (Required) `<string>` The UUID of an existing Kafka instance.

* `name` - (Required) `<string>` The name of the consumer group.

## Attributes Reference

The following attributes are exported:

N/A

## Import

An existing consumer group can be imported using a composite value of the Kafka ID and consumer group name
separated by a colon.

For example:

```shell script
$ terraform import herokux_kafka_consumer_group.foobar "2bccd770-e7aa-4865-98d2-6e222f2d2582:ac42b355-074e-493a-8fa4-d7e4364623a3"
```