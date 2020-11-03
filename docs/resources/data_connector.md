---
layout: "herokux"
page_title: "HerokuX: herokux_data_connector"
sidebar_current: "docs-herokux-resource-data-connector"
description: |-
  Provides a resource to manage a Data Connector for a Heroku Redis, Postgres, or Kafka addon.
---

# herokux\_data\_connector

This resource manages a data connector in Heroku. A data connector configures Change Data Capture (CDC)
for Heroku Postgres events and stream them to your Apache Kafka on Heroku add-on provisioned in a Private Space
or a Shield Private Space. This resource requires Postgres and Kafka addons installed in a Private or Private Shield space.

Additional Heroku documentation:

- [Best Practices for Heroku's Streaming Data Connectors](https://devcenter.heroku.com/articles/best-practices-for-heroku-data-connectors)
- [Heroku's streaming data connectors](https://devcenter.heroku.com/articles/heroku-data-connectors)

-> **IMPORTANT!**
Per this [Heroku article](https://devcenter.heroku.com/articles/heroku-data-connectors), data connectors are a Beta feature.
Any use of Beta Services is subject to the terms in your Master Subscription Agreement and the Beta Services terms.
These terms include provisions that the following types of sensitive Personal Data (including images, sounds or other
information containing or revealing such sensitive data) may not be submitted to Data Science Programs, Non-GA Service
and Non-GA Software: government-issued identification numbers; financial information (such as credit or debit card numbers,
any related security codes or passwords, and bank account numbers); racial or ethnic origin, political opinions, religious
or philosophical beliefs, trade-union membership, information concerning health or sex life; information related to an individual’s
physical or mental health; and information related to the provision or payment of health care.

### Resource Timeouts
During creation, modification and deletion, this resource checks the status of the data connector.
All the aforementioned timeouts can be customized via the `timeouts.data_connector_create_timeout`, `timeouts.data_connector_delete_timeout`
and `timeouts.data_connector_update_timeout` attributes in your `provider` block.

For example:

```hcl-terraform
provider "herokux" {
  timeouts {
    data_connector_create_timeout = 20
    data_connector_delete_timeout = 20
    data_connector_update_timeout = 20
  }
}
```

## Example Usage

```hcl-terraform
resource "herokux_data_connector" "foobar" {
	source_id = "<SOME_POSTGRES_ID>"
	store_id = "<SOME_KAFKA_ID"
	name = "my-custom-connector-name"
	tables = ["public.users"]
}
```

## Argument Reference

The following arguments are supported:

* `source_id` - (Required) `<string>` The UUID of the database instance whose change data you want to store.
This is an existing Heroku Postgres addon.

* `store_id` - (Required) `<string>` The UUID of the database instance that will store the change data.
This is an existing Heroku Kafka addon.

* `tables` - (Required) `<list(string)>` Tables to connect.

* `name` - `<string>` Name of the connector. If no name is supplied, Heroku generates one automatically for you.

* `state` - `<string>` Controls whether to pause or resume the data connector. Valid options are: `available` or `paused`.
By default, the data connector is `available` on initial creation. Please also note the following:

    * No action is taken if `state` is set to `available` on initial resource creation.
    * No action is taken if this attribute is removed from a configuration.
    * This attribute value will generally mirror the exported `status` attribute.

* `excluded_columns` - `<list(string)>` List of columns to exclude.

* `settings` - `<map>` Properties of the connector. Please [visit](https://devcenter.heroku.com/articles/heroku-data-connectors#update-configuration)
this article for more information.

-> **IMPORTANT!**
The only updatable attributes are `settings` and `state`. All other attribute modifications will result
in destruction and recreation.

## Attributes Reference

The following attributes are exported:

* `status` - The status of data connector.

## Import

An existing data connector can be imported using the data connector UUID.

For example:

```shell script
$ terraform import herokux_data_connector.foobar <UUID>
```