---
layout: "herokux"
page_title: "Herokux: herokux_kafka_mtls_iprules"
sidebar_current: "docs-herokux-datasource-kafka-mtls-iprules-x"
description: |-
  Get information about all MTLS IP rules for Heroku Private or Shield Kafka.
---

# Data Source: herokux_kafka_mtls_iprules

Use this data source to get information about all MTLS IP rules for a Heroku Private or Shield Kafka cluster.

## Example Usage

```hcl-terraform
data "heroku_addon" "kafka" {
  name = "kafka-fitted-123"
}

data "herokux_kafka_mtls_iprules" "rules" {
  kafka_id = data.heroku_addon.kafka.id
}
```

## Argument Reference

The following arguments are supported:

* `kafka_id` - (Required) The UUID of the Kafka.

## Attributes Reference

The following attributes are exported:

* `rules` - A list of maps containing the following IP rule information:
    * `id` - The UUID of the IP rule.
    * `cidr` - Range of IP addresses.
    * `description` - The description of the IP rule.
    * `status` - The status of the IP rule.