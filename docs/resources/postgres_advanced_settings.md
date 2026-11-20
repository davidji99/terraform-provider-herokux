---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_advanced_settings"
description: |-
  Manages selected settings on a Heroku Postgres Advanced database.
---

# herokux_postgres_advanced_settings

Manage selected database settings with the same API as `heroku data:pg:settings`. For classic plans, use `herokux_postgres_settings`. [Heroku Postgres settings](https://devcenter.heroku.com/articles/heroku-postgres-settings)

## Example usage

```hcl
resource "herokux_postgres_advanced_settings" "orders" {
  postgres_id = var.postgres_addon_uuid
  settings = {
    log_lock_waits             = "off"
    log_min_duration_statement = "500"
    "auto_explain.log_format"  = "json"
  }
}
```

## Arguments

* `postgres_id` (required) is the Advanced Postgres add-on UUID.
* `settings` (required) is a map of setting names to string values. Use the names and value formats reported by `heroku data:pg:settings` and quote dotted names in HCL. For boolean settings, you can also use `on` and `off`, or `1` and `0`. The API uses commas and colons as delimiters.

Terraform sends changed settings together and waits for Heroku to report their new values. Removing a key from the map stops managing that setting and preserves its value on Heroku. Destroying the resource preserves all database settings.

## Import

Import by add-on UUID to read all current, non-null settings into state.

```shell
terraform import herokux_postgres_advanced_settings.orders '29138cab-acf1-40c6-ac83-56697c8f0221'
```

Keep the keys you want to manage in `settings`. If you configure a subset, the first plan shows the other keys leaving the map. Applying that plan leaves those values on Heroku.
