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
* `settings` (required) is a map of setting names to string values. Quote dotted names in HCL. Terraform checks known names and values during validation and planning. An unknown name produces an error listing the supported settings. A new Heroku setting becomes configurable when a provider release adds it to this table.

| Setting | Accepted configuration values |
| --- | --- |
| `auto_explain.log_analyze` | Boolean |
| `auto_explain.log_buffers` | Boolean. Buffer reporting takes effect when `auto_explain.log_analyze` is enabled. |
| `auto_explain.log_format` | `text`, `xml`, `json`, `yaml` |
| `auto_explain.log_min_duration` | Whole milliseconds from `-1` to `2147483647`. Use `-1` to disable plan logging or `0` to log every plan. |
| `auto_explain.log_nested_statements` | Boolean |
| `auto_explain.log_triggers` | Boolean |
| `auto_explain.log_verbose` | Boolean |
| `idle_in_transaction_session_timeout` | Whole milliseconds from `0` to `2147483647`. Use `0` to disable the timeout. |
| `log_connections` | Boolean |
| `log_lock_waits` | Boolean |
| `log_min_duration_statement` | Whole milliseconds from `-1` to `2147483647`. Use `-1` to disable duration logging or `0` to log every statement. |
| `log_min_error_statement` | `error`, `log`, `fatal`, `panic` |
| `log_statement` | `none`, `ddl`, `mod`, `all` |
| `track_functions` | `none`, `pl`, `all` |

Boolean settings accept `true`/`false`, `on`/`off`, `yes`/`no`, and `1`/`0`. Letter case is ignored for boolean values. Write durations as decimal strings such as `"500"`, with no unit suffix, leading zeros, or plus sign.

Terraform sends changed settings together and waits for Heroku to report their new values. Removing a key from the map stops managing that setting and preserves its value on Heroku. Destroying the resource preserves all database settings.

## Import

Import by add-on UUID to read all current, non-null settings into state. Import and refresh preserve values reported by Heroku, including values such as `log_min_error_statement = "info"` that are outside the configuration choices above.

```shell
terraform import herokux_postgres_advanced_settings.orders '29138cab-acf1-40c6-ac83-56697c8f0221'
```

Keep the keys you want to manage in `settings`. If you configure a subset, the first plan shows the other keys leaving the map. Applying that plan leaves those values on Heroku.
