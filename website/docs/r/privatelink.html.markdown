---
layout: "herokux"
page_title: "HerokuX: herokux_privatelink"
sidebar_current: "docs-herokux-resource-privatelink"
description: |-
  Provides a resource to manage a privatelink for a Heroku Redis, Postgres, or Kafka addon.
---

# herokux\_privatelink

This resource manages the privatelink configuration for a Heroku Redis, Postgres, or Kafka addon.
For more information about each addon & privatelink, please refer to the following documentations:
- [Connecting to Heroku Redis in a Private or Shield Space via PrivateLink](https://devcenter.heroku.com/articles/heroku-redis-via-privatelink)
- [Connecting to a Private or Shield Heroku Postgres Database via PrivateLink](https://devcenter.heroku.com/articles/heroku-postgres-via-privatelink)
- [Connecting to Apache Kafka on Heroku in a Private or Shield Space via PrivateLink](https://devcenter.heroku.com/articles/heroku-kafka-via-privatelink)

-> **IMPORTANT!**
To use this resource, the Amazon VPC Endpoint you create must be provisioned in a subnet
that is in the same region as your Apache Kafka on Heroku add-on. New PrivateLink endpoints typically take
between 5 and 10 minutes to become available.

### Resource Timeouts
During creation and deletion, this resource checks the status of the privatelink provisioning/deprovisioning
as well as allowlisting AWS account IDs. All the aforementioned timeouts can be customized
via the `timeouts.privatelink_create_timeout`, `timeouts.privatelink_delete_timeout`
and `timeouts.privatelink_allowed_acccounts_add_timeout` in your `provider` block.

For example:
```hcl-terraform
provider "herokux" {
  timeouts {
    privatelink_create_timeout = 20
    privatelink_delete_timeout = 20
    privatelink_allowed_acccounts_add_timeout = 20
  }
}
```

## Example Usage

```hcl-terraform
resource "herokux_privatelink" "foobar" {
	addon_id = "SOME_ADDON_ID"
	allowed_accounts = ["123456789123", "123456789124"]
}
```

## Argument Reference

The following arguments are supported:

* `addon_id` - (Required) `<string>` The UUID of a Heroku postgres, redis or kafka addon.

* `allowed_accounts` - (Required) `<list(string)>` Unordered list of AWS account IDs.

## Attributes Reference

The following attributes are exported:

* `status` - The status of privatelink configuration.

* `service_name` - The privatelink endpoint service name.

## Import

An existing privatelink can be imported using the addon UUID.

For example:
```shell script
$ terraform import herokux_privatelink.foobar <ADDON_UUID>
```
