---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_mtls"
sidebar_current: "docs-herokux-resource-postgres-mtls"
description: |-
  Provides a resource to manage the MTLS configuration of a postgres database
---

# herokux\_postgres\_mtls

This resource manages the MTLS configuration of a postgres database in Heroku.

## Example Usage

```hcl-terraform
resource "herokux_postgres_mtls" "foobar" {
	database_name = "my_database_name"
}
```

## Argument Reference

The following arguments are supported:

* `database_name` - (Required) `<string>` The name of the database. DO NOT use the database UUID.

## Attributes Reference

The following attributes are exported:

* `app_name` - The app which the postgres addon is tied to.
* `status` - The status of MTLS configuration.
* `enabled_by` - The Heroku user that enabled the MTLS configuration.
* `certificate_authority_chain` - the certificate authority chain

## Import

Existing database MTLS configurations can be imported using the database name

For example:
```shell script
$ terraform import herokux_postgres_mtls.foobar <MY_APP_NAME>
```