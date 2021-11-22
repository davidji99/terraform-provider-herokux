---
layout: "herokux"
page_title: "HerokuX: herokux_scheduler_job"
sidebar_current: "docs-herokux-resource-scheduler-job"
description: |-
  Provides a resource to manage a Heroku Scheduler job.
---

# herokux_scheduler_job

This resource manages Heroku [Scheduler jobs](https://devcenter.heroku.com/articles/scheduler).

Please make sure you are aware of [Known Issues and Alternatives](https://devcenter.heroku.com/articles/scheduler#known-issues-and-alternatives).

## Example Usage

```hcl-terraform
resource "heroku_app" "foobar" {
  name   = "my_foobar_app"
  region = "us"

  organization {
    name = "my_org"
  }
}

resource "heroku_addon" "scheduler" {
  app  = heroku_app.foobar.name
  plan = "scheduler:standard"
}

resource "herokux_scheduler_job" "foobar" {
  app_id = heroku_app.foobar.uuid
  command = "rake update_feed"
  dyno_size = "Standard-1X"
  frequency = "every_hour_at_30"

  # required in order for Terraform to wait for scheduler addon creation before creating jobs.
  depends_on = [heroku_addon.scheduler]
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) `<string>` The UUID of a Heroku app.
* `command` - (Required) `<string>` The command to run.
* `dyno_size` - (Required) `<string>` Size of [dyno](https://devcenter.heroku.com/articles/dyno-types).
  Valid options are (case-sensitive):
    * `Free`, `Hobby`, `Standard-1X`, `Standard-2X`, `Performance-(M|L)`, `Private-(S|M|L)`, `Shield-(S|M|L)`
* `frequency`  - (Required) `<string>` The interval by which the job will run on schedule in UTC.
See [Frequency specification](#frequency) below for more details.

### `frequency`

Choose one of the following (case-sensitive):

* `every_ten_minutes`
* `every_hour_at_##` - Valid values for `##` are `0`, `10`, `20`, `30`, `40`, `50`.
* `every_day_at_HH:MM` - Valid values for `HH` and `MM` are (using 24hour time format):
* `HH` - `0` through `23`. For example, `5`, `16`, etc. No leading `0` for hours 0-9.
* `MM` - either `30` or `00`.


## Attributes Reference

The following attributes are exported:

* `last_run_at` - When the job last run at

## Import

Existing schedule jobs can be imported using a composite value of the Heroku app UUID and job ID separated
by a colon. The simplest way to locate the job ID is to visit the scheduler webpage and click on the pencil
icon next to an existing job. The job ID will be provided in the URL: `https://dashboard.heroku.com/apps/57d660e0-3d20-40b7-8d20-e77b95189e5a/scheduler?job=881234`

For example:

```shell script
$ terraform import herokux_scheduler_job.foobar "57d660e0-3d20-40b7-8d20-e77b95189e5a:881234"
```
