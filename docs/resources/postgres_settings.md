---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_settings"
sidebar_current: "docs-herokux-resource-postgres-settings"
description: |-
  Provides a resource to manage the settings of a postgres database
---

# herokux\_postgres\_settings

This resource manages the [settings](https://devcenter.heroku.com/articles/heroku-postgres-settings)
of a Heroku postgres database.

-> **IMPORTANT!**
It is not possible to delete a postgres database's settings, so the resource will just be removed from state upon deletion.

### Resource Delay
This resource will wait a given amount of time after modification. This delay is to allow for Heroku to actually reflect
the changes to the database before the settings can be modified again. Heroku documentation does not provide
a recommended or estimated delay period, so this provider provides the means for users to configure this delay period.
It is likely the size of the database will affect the length of this delay.

The aforementioned delay can be customized via the `delays.postgres_settings_modify_delay` attribute in your `provider` block.
The delay value defaults to 2 minutes with a minimum requirement of 1 minute.

For example:

```hcl-terraform
provider "herokux" {
  delays {
    postgres_settings_modify_delay = 20
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
  app_id  = heroku_app.foobar.id
  plan = "heroku-postgresql:premium-0"
}

resource "herokux_postgres_settings" "foobar" {
  postgres_id = heroku_addon.database.id
  log_lock_waits = true
  log_connections = false
  log_min_duration_statement = 123
  log_statement = "none"
}
```

## Argument Reference

The following arguments are supported:

* `postgres_id` - (Required) `<string>` The UUID or name of a Heroku postgres addon.
* `log_lock_waits` - `<boolean>` Enables logging when a session waits longer than 1 second
  to acquire a lock. This is useful in determining if lock waits are causing poor performance issues.
* `log_connections` - `<boolean>` Enables logging of all attempted connection.
* `log_min_duration_statement` - `<integer>` Causes the duration of each completed statement to be logged
  if the statement ran for at least the specified number of milliseconds. A value of `0` will log everything,
  and a value of `-1` will disable logging.
* `log_statement` - `<string>` Controls which normal SQL statements are logged. This feature is useful
  when hunting a bug that involves complex queries or inspecting queries made by your app or any database user.
  Valid values for log-statement are:
    * `none`: Stops logging normal queries. Other logs will still be generated such as slow query logs, queries waiting in locks, and syntax errors.
    * `ddl`: All data definition statements, such as CREATE, ALTER and DROP will be logged.
    * `mod`: Includes all statements from ddl as well as data-modifying statements such as `INSERT`, `UPDATE`, `DELETE`, `TRUNCATE`, `COPY.`
    * `all`: All statements are logged.

## Import

Existing postgres settings can be imported using the postgres ID.

For example:

```shell script
$ terraform import herokux_postgres_settings.foobar "867f0740-82f9-4b9d-9994-cfbae2011abc"
```
