---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_connection_pooling"
sidebar_current: "docs-herokux-resource-postgres-connection-pooling"
description: |-
Provides a resource to manage a Heroku Postgres connection pooling.
---

# herokux_postgres_connection_pooling

This resource manages a [connection pooling](https://devcenter.heroku.com/articles/postgres-connection-pooling)
for a Heroku Postgres database.

When you add create a connection pooling, this adds a new config var called `DATABASE_CONNECTION_POOL_URL`
(or whatever you specify for the `name` attribute), which your app can connect to just like any other Postgres URL.
Essentially, Heroku adds an addon attachment for Heroku postgres database.

-> **IMPORTANT!**
Heroku recommends using the default name `DATABASE_CONNECTION_POOL_URL`. Connection poolers under this name will
automatically get reattached to the new leader during an upgrade of a Postgres version or when changing plans.

## Example Usage

```hcl-terraform
resource "heroku_app" "foobar" {
  name   = "my_foobar_app"
  region = "us"

  organization {
    name = "my_org"
  }
}

resource "heroku_addon" "foobar" {
  app  = heroku_app.foobar.name
  plan = "heroku-postgresql:standard-0"
}

resource "herokux_postgres_connection_pooling" "foobar" {
  postgres_id = heroku_addon.foobar.id
  app_id = heroku_app.foobar.uuid
  name = "CUSTOM_DB_POOL"
}
```

## Argument Reference

The following arguments are supported:

* `postgres_id` - (Required) `<string>` The UUID of the Postgres database.
* `app_id` - (Required) `<string>` The UUID of the app.
* `name` - (Optional) `<string>` Base name of the new config var. Must start with a letter and can only contain
  uppercase letters, numbers, and underscores. Default value is `DATABASE_CONNECTION_POOL`.
    * Any modifications to `name` will result in resource recreation as it is not possible to modify an existing
      connection pooling.

## Attributes Reference

The following attributes are exported:

* `config_var` - The connection pooling config var. This value is essentially `_URL` appended to the value of `name` argument.

## Import

An existing Postgres connection pooling can be imported using the connection pooling ID, which is also
the addon attachment ID.

For example:

```shell script
$ terraform import herokux_postgres_connection_pooling.foobar "78b6e411-1361-4adf-ace6-734c9a95513d"
```
