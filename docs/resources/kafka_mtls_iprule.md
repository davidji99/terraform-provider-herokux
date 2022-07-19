---
layout: "herokux"
page_title: "HerokuX: herokux_kafka_mtls_iprule"
sidebar_current: "docs-herokux-resource-kafka-mtls-iprule"
description: |-
  Provides a resource to manage MTLS IP rules for Heroku Private or Shield Kafka.
---

# herokux\_kafka\_mtls\_iprule

This resource manages MTLS IP rules for a Heroku Private or Shield Kafka cluster. There is a hard limit of 60 IP blocks
that can be allowlisted per cluster.

-> **IMPORTANT!**
Deleting and re-adding a MTLS IP rule using the same CIDR range in succession may cause an unknown server error
in Heroku. Please wait a bit after destruction before attempting to recreate the IP rule.
The actual wait time is unknown at the moment.

### Resource Timeouts
During creation, this resource verifies if the MTLS IP rule status successfully changes from 'Authorizing' to 'Authorized'.
This check's default timeout is ~10 minutes, which can be customized via the `timeouts.mtls_iprule_create_verify_timeout`
attribute in your `provider` block.

For example:

```hcl-terraform
provider "herokux" {
  timeouts {
    mtls_iprule_create_verify_timeout = 15
  }
}
```

## Example Usage

```hcl-terraform
resource "heroku_space" "foobar" {
  name         = "foobar-space"
  organization = "my_org"
  region       = "virginia"
}

resource "heroku_app" "foobar" {
  name   = "my_foobar_app"
  region = "us"
  space  = heroku_space.foobar.name

  organization {
    name = "my_org"
  }
}

resource "heroku_addon" "kafka" {
  app_id  = heroku_app.foobar.id
  plan = "heroku-kafka:private-0"
}

resource "herokux_kafka_mtls_iprule" "foobar" {
  kafka_id    = heroku_addon.kafka.id
  cidr        = "1.2.3.4/32"
  description = "CI/CD outbound IPs"
}
```

## Argument Reference

The following arguments are supported:

* `kafka_id` - (Required) `<string>` The UUID of a Kafka cluster/addon.
* `cidr` - (Required) `<string>` Valid IPv4 CIDR value. Example: `1.2.3.4/32`.
* `description` - (Optional) `<string>` A description of the MTLS IP rule.

## Attributes Reference

The following attributes are exported:

* `status` - The status of the MTLS IP rule.

## Import

An existing Kafka MTLS IP rule can be imported using a composite value of the Kafka UUID and CIDR separated
by a colon.

For example:

```shell script
$ terraform import herokux_kafka_mtls_iprule.foobar "1d17bd09-6ad2-4a39-b50a-e02e467f5ee2:1.2.3.4/32"
```
