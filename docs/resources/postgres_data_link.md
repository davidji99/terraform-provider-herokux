---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_data_link"
sidebar_current: "docs-herokux-resource-postgres-data-link"
description: |-
  Provides a resource to manage a Data link between two Heroku Postgres databases.
---

# herokux\_postgres\_data\_link

This resource manages a postgres data link in Heroku.

[Heroku Data Links](https://devcenter.heroku.com/articles/heroku-data-links) allows you to connect disparate
data sources, such as Heroku Postgres, to a Heroku Postgres database.
This allows you to use SQL semantics to read and write data, regardless of where the data lives.

Heroku Data Links is only available on production tier Heroku Postgres databases (standard, premium, and enterprise)
that are on Postgres version 9.4 and above. Data Links is not available on hobby tier databases.

When you delete the remote data store, or the remote date store moves to another host (e.g. Postgres failover),
the data link will not be updated. This means you will need to `taint` the existing `herokux_postgres_data_link` resource
and `apply` to recreate the resource.

When you reset the local database with the command like `heroku pg:reset`, the data link information will also be reset.
You will need to recreate the data link for these cases. This means the subsequent `plan` or `apply` will show terraform
attempting to recreate the data link again.

-> **IMPORTANT!**
Only the tables that exist at the time of the link creation will be available. If you create tables after you created
the link, they will not show up in the local database. If you want to include newly created tables, you will need
to `taint` the existing `herokux_postgres_data_link` resource and `apply`.

-> **IMPORTANT!**
Due to [Heroku documentation](https://devcenter.heroku.com/articles/heroku-data-links#linking-heroku-redis-to-heroku-postgres)
stating creation of Heroku Data Links from Heroku Redis to Heroku Postgres are deprecated, this resource will not support
this data link setup.

## Example Usage

```hcl-terraform
resource "heroku_app" "foobar" {
  name   = "my_foobar_app"
  region = "us"

  organization {
    name = "my_org"
  }
}

resource "heroku_addon" "local_db" {
  app  = heroku_app.foobar.name
  plan = "heroku-postgresql:premium-0"
}

resource "heroku_addon" "remote_db" {
  app  = heroku_app.foobar.name
  plan = "heroku-postgresql:premium-0"
}

resource "herokux_postgres_data_link" "foobar" {
  local_db_id = heroku_addon.local_db.id
  remote_db_name = heroku_addon.remote_db.name
  name = "my_custom_data_l1nk_name"
}
```

## Argument Reference

The following arguments are supported:

* `local_db_id` - (Required) `<string>` The UUID of a Heroku Postgres database that’s accepting the data link connection.
* `remote_db_name` - (Required) `<string>` The Postgres database name that is being connected to a Heroku Postgres database.
* `name` - (Optional) `<string>` The name of connection between the remote and local databases. If a custom `name`
is not defined, it will be the same value as the `remote_db_name` with underscores in place of hyphens.
A custom name must be respect the following restrictions:
    * Between 3-63 alphanumeric characters
    * Start with a letter & end with an alphanumeric character
    * No symbols/spaces besides an underscore

-> **IMPORTANT!**
Any changes to the attributes listed above will result in a resource destruction and recreation.

## Attributes Reference

The following attributes are exported:

* `link_id` - The UUID of the data link.
* `remote_attachment_name` - The remote database attachment name. For example, `HEROKU_POSTGRESQL_COPPER_URL`.

## Import

An existing data link can be imported using a composite value of the local database UUID
and data link name. The quickest way to figure out the link name is to execute `heroku pg:links --app "NAME_OF_APP"`.
The data link name is the string with an asterisk on the left-hand side in the command output.

For example:

```shell script
$ terraform import herokux_postgres_data_link.foobar "6fae1ee0-c034-4775-a798-890bc64f98eb:my_custom_data_l1nk_name"
```
