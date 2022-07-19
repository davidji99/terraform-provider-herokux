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

* `postgres_id` - (Required) `<string>` The UUID for a Postgres database.
* `hour` - (Required) `<integer>` Hour at which to run the backup. Acceptable values are between `0` & `23`.
* `timezone` - (Optional) `<string>` The timezone in the [full TZ format](http://en.wikipedia.org/wiki/List_of_tz_database_time_zones) (America/Los_Angeles).
Except for `UTC`, this resource's underlying API requires the timezone to be in full TZ format. Defaults to `UTC` if not set.

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
