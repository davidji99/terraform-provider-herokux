---
layout: "herokux"
page_title: "HerokuX: herokux_redis_maintenance_window"
sidebar_current: "docs-herokux-resource-redis-maintenance-window"
description: |-
  Provides a resource to manage a maintenance window of a redis instance
---

# herokux\_redis\_maintenance\_window

This resource manages the [maintenance window](https://devcenter.heroku.com/articles/heroku-redis-maintenance)
for a Heroku redis instance.

Maintenance windows are 4 hours long starting at the time you specify. The actual time required for maintenance
depends on exactly what’s taking place, but it will usually require your Redis instance to be offline for only a
few minutes. If you don’t specify a window, one will be selected randomly.

-> **IMPORTANT!**
It is not possible to delete a maintenance window, so the resource will just be removed from state upon deletion.

## Example Usage

```hcl-terraform
resource "heroku_app" "foobar" {
  name   = "my_foobar_app"
  region = "us"

  organization {
    name = "my_org"
  }
}

resource "heroku_addon" "redis" {
  app_id  = heroku_app.foobar.id
  plan = "heroku-redis:premium-0"
}

resource "herokux_redis_maintenance_window" "foobar" {
  redis_id = heroku_addon.redis.id
  window = "Mondays 10:30"
}
```

## Argument Reference

The following arguments are supported:

* `redis_id` - (Required) `<string>` The UUID or name of a Heroku redis instance.
* `window` - (Required) `<string>` The day-of-week and time (UTC) at which the window begins.
For example: `Sundays 10:30`. Note: the `s` attached to the word `Sunday` is required.

The provider resolves the add-on name or UUID through the Platform API and manages the UTC window through the [Data Maintenance API](https://devcenter.heroku.com/articles/data-maintenance-cli-commands). Existing resource and import IDs retain the identifier you supplied. If you use a custom Data REST endpoint, set `postgres_advanced_api_url` on the provider. Its default is `https://api.data.heroku.com`.

## Import

An existing maintenance window can be imported using the redis ID.

For example:

```shell script
$ terraform import herokux_redis_maintenance_window.foobar "717e9e8f-c4ad-4f45-9ac3-069ecb0fcd60"
```
