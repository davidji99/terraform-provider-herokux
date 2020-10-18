---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_maintenance_window"
sidebar_current: "docs-herokux-resource-postgres-maintenance-window"
description: |-
  Provides a resource to manage a maintenance window of a postgres database
---

# herokux\_postgres\_maintenance\_window

This resource manages the [maintenance window](https://devcenter.heroku.com/articles/heroku-postgres-maintenance)
for a Heroku postgres database.

Maintenance windows are 4 hours long. Heroku attempts to begin the maintenance as close to the beginning of the specified window as possible.
The duration of the maintenance can vary, but usually, your database is only offline for 10 – 60 seconds.

Note that the window you specify must end before the required by time indicated by the `heroku pg:maintenance` command.

-> **IMPORTANT!**
It is not possible to delete a maintenance window, so the resource will just be removed from state upon deletion.

## Example Usage

```hcl-terraform
resource "herokux_postgres_maintenance_window" "foobar" {
	postgres_id = "SOME_POSTGRES_ID"
	window = "Mondays 10:30"
}
```

## Argument Reference

The following arguments are supported:

* `postgres_id` - (Required) `<string>` The UUID of a Heroku postgres addon.

* `window` - (Required) `<string>` The day-of-week and time (UTC) at which the window begins.
For example: `Sundays 10:30`. Note: the `s` attached to the word `Sunday` is required.

## Import

An existing maintenance window can be imported using the postgres ID.

For example:
```shell script
$ terraform import herokux_postgres_maintenance_window.foobar "<POSTGRES_ID>"
```
