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

This resource **cannot be used** to manage backup schedules for hobby tier Postgres databases.

-> **IMPORTANT!**
Production tier Postgres databases have Continuous protection enabled, so scheduled backups of large databases
are likely to fail.

## Example Usage

```hcl-terraform
resource "herokux_postgres_backup_schedule" "foobar" {
  postgres_id = "2508ebbd-74bb-4e81-a63c-d193d2bd5716"
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
