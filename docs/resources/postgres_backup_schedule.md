---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_backup_schedule"
sidebar_current: "docs-herokux-resource-postgres-backup-schedule"
description: |-
Provides a resource to manage a backup schedule for a Postgres database.
---

# herokux\_postgres\_backup\_schedule

This resource manages a [backup schedule](https://devcenter.heroku.com/articles/heroku-postgres-backups#scheduling-backups)
for a Postgres database.

When a schedule is removed outside Terraform, refresh clears its stored ID. If the resource is still configured, the next plan schedules a backup again.

By default, this resource is configured to work with Heroku Postgres professional plans. If you wish to set backup
schedules for Heroku Postgres starter plans, please set the
[`postgres_api_url` attribute](https://registry.terraform.io/providers/davidji99/herokux/latest/docs#postgres_api_url)
to ` https://postgres-starter-api.heroku.com`. If you have a mix of starter and professional Postgres databases within one
terraform configuration, please consider leveraging [provider aliases](https://www.terraform.io/docs/language/providers/configuration.html).

-> **IMPORTANT!**
Production tier Postgres databases have Continuous protection enabled, so scheduled backups of large databases
are likely to fail.

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

resource "herokux_postgres_backup_schedule" "foobar" {
  postgres_id = heroku_addon.database.id
  hour        = 3
  timezone    = "Australia/Perth"
}
```

## Argument Reference

The following arguments are supported:

* `postgres_id` - (Required) `<string>` The UUID or name for a Postgres database.
* `hour` - (Required) `<integer>` Hour at which to run the backup. Acceptable values are between `0` & `23`.
* `timezone` - (Optional) Time zone for the local backup hour. Defaults to `UTC`. Supply a full name such as `America/Los_Angeles` or a case-insensitive Heroku CLI alias. Terraform sends and stores the expanded zone name. For example, both `PST` and `PDT` follow Los Angeles local time throughout the year. [Heroku CLI alias table](https://github.com/heroku/cli/blob/v11.14.0/src/commands/pg/backups/schedule.ts)

| CLI aliases | Time zone |
| --- | --- |
| `BST`, `GMT` | `Europe/London` |
| `CDT`, `CST` | `America/Chicago` |
| `CEST`, `CET` | `Europe/Paris` |
| `EDT`, `EST` | `America/New_York` |
| `MDT`, `MST` | `America/Boise` |
| `PDT`, `PST` | `America/Los_Angeles` |
| `Z` | `UTC` |

## Attributes Reference

The following attributes are exported:

* `name` - The environment variable of the postgres database (`DATABASE_URL`).
* `retain_weeks` - The [number of weeks](https://devcenter.heroku.com/articles/heroku-postgres-backups#scheduled-backups-retention-limits)
Heroku will retain a scheduled backup.
* `retain_months` - The [number of months](https://devcenter.heroku.com/articles/heroku-postgres-backups#scheduled-backups-retention-limits)
Heroku will retain a scheduled backup.

## Import

An existing database backup schedule can be imported using the postgres database UUID.

For example:

```shell script
$ terraform import herokux_postgres_backup_schedule.foobar "2508ebbd-74bb-4e81-a63c-d193d2bd5716"
```
