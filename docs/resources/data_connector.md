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
During creation, modification and deletion, this resource checks the progress of the action you took for the `apply`.
All the aforementioned timeouts can be customized via the following attributes in your `provider` block:

* `data_connector_create_verify_timeout`
* `data_connector_delete_verify_timeout`
* `data_connector_status_update_verify_timeout`
* `data_connector_settings_update_verify_timeout`

For example:

```hcl-terraform
provider "herokux" {
  timeouts {
    data_connector_create_verify_timeout = 20
    data_connector_delete_verify_timeout = 20
    data_connector_status_update_verify_timeout = 20
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

resource "heroku_addon" "database" {
  app  = heroku_app.foobar.name
  plan = "heroku-postgresql:premium-0"
}

resource "heroku_addon" "kafka" {
  app  = heroku_app.foobar.name
  plan = "heroku-kafka:standard-0"
}

resource "herokux_data_connector" "foobar" {
  source_id = heroku_addon.database.id
  store_id = heroku_addon.kafka.id
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
* `settings` - `<map>` Properties of the connector. Please visit [this article](https://devcenter.heroku.com/articles/heroku-data-connectors#update-configuration)
for a list of valid properties that can be set for this attribute. Please also note the following:
    * Due to API limitations, settings that are removed do not get unset remotely. Therefore, if you wish to remove a settings,
      please set its value to the default value as mentioned in the article above and keep the setting in your configuration.

-> **IMPORTANT!**
The only updatable attributes are `settings` and `state`. All other attribute modifications will result
in destruction and recreation.

## Attributes Reference

The following attributes are exported:

* `status` - The status of data connector.
* `lag` - The lag of the data connector.
* `source_app_name` - The source's app name.
* `store_app_name` - The store's app name.

## Import

An existing data connector can be imported using a composite value of the source app name and data connector name
separated by a colon.

The simplest way to get the data connector name is via the `heroku data:connectors:list --app APP` command.

For example:

```shell script
$ terraform import herokux_data_connector.foobar "SOURCE_APP_NAME:DATA_CONNECTOR_NAME"
```