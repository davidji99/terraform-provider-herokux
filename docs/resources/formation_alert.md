---
layout: "herokux"
page_title: "HerokuX: herokux_formation_alert"
sidebar_current: "docs-herokux-resource-formation-alert"
description: |-
Provides a resource to manage an alert for a given formation.
---

# herokux\_formation\_alert

This resource manages an alert for a Heroku formation. In Heroku documentation, this resource is known as a
['Threshold Alert'](https://devcenter.heroku.com/articles/metrics#threshold-alerting). This resource is available to apps
running on Professional, Private, and Shield dynos. For those using Heroku Enterprise Teams, the `operate` permission
is required to use this resource.

It is recommended to add the [`depends_on`](https://www.terraform.io/docs/language/meta-arguments/depends_on.html) meta-argument
to `herokux_formation_alert` when `heroku_app_release` is present. This will make Terraform wait for a running dyno
before creating/modifying a formation alert as an alert cannot be enabled with no running dyno/formation.

See the example [resource configuration](#example-usage) below on how to use `heroku_app_release` with `herokux_formation_alert`.

-> **IMPORTANT!**
When an existing `herokux_formation_alert` is deleted, the provider will disable the alert remotely
and remove the resource from state.

~> **WARNING:**
Please make sure you understand all [common issues](#common-issues) prior to using this resource. Failure to understand
them will result in potentially bricking your alert(s).

## Common Issues

1. In the event you remove an existing `herokux_formation_alert.foobar` resource after it's been successfully applied
   to an app, you **MUST** `import` the resource first if the new `herokux_formation_alert.foobar` resource is targeting
   the same app+alert prior to its removal. Otherwise, the resource will error with a message indicating
   a resource `import` prerequisite. This is due to following reasons:

    * The resource does not delete the formation alert during resource destruction as it'll render any subsequent
      alert management impossible for the same app/dyno process type.

    * Due to the above reason, the underlying API does not allow for a `POST` request when an existing formation alert
      exists for the app. Therefore, the resource must be imported first and then modified afterwards.

1. In the event you fail to adhere with the aforementioned guidance, the only solution is to delete the app and start over.

## Example Usage

```hcl-terraform
variable "process_type" {
  value = "web"
}

variable "dyno_size" {
  value = "Performance-L"
}

resource "heroku_app" "foobar" {
  name   = "my-cool-app"
  region = "us"

  config_vars = {
    FOOBAR = "baz"
  }
}

resource "heroku_slug" "foobar" {
  app_id   = heroku_app.foobar.id
  file_url = "url_to_slug_artifact"

  process_types = {
    web = "ruby server.rb"
  }
}

resource "heroku_app_release" "foobar" {
  app_id  = heroku_app.foobar.id
  slug_id = heroku_slug.foobar.id
}

resource "heroku_formation" "foobar" {
  app_id = heroku_app.foobar.id
  type = var.process_type
  quantity = 8
  size = var.dyno_size

  # Tells Terraform that this formation must be created/updated only after the app release has been created
  depends_on = [heroku_app_release.foobar-release]
}

resource "herokux_formation_alert" "foobar" {
  app_id = heroku_app.foobar.uuid
  process_type = heroku_formation.foobar.type
  name = "LATENCY"
  threshold = "1202"
  sensitivity = 10
  is_active = true
  notification_frequency = 1440
  notification_channels = ["app"]

  # Tells Terraform that this formation alert resource must be created/updated
  # only after the app release has been successfully applied.
  depends_on = [heroku_app_release.foobar]
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) `<string>` An existing app's UUID. The app name is not valid for this argument.
* `process_type` - (Required) `<string>` The type of the dyno formation process, such as `web`.
* `name` - (Required) `<string>` The name of the alert aka the alert 'type'.
  Valid options are `LATENCY` & `ERROR_RATE` (case-sensitive):
    * `LATENCY` corresponds to the 'Monitor Response Times' alert in the UI.
    * `ERROR_RATE` corresponds to the 'Monitor Failed Requests' alert in the UI.
* `threshold` - (Required) `<string>` When to trigger an alert. This attribute value is dependent on
  the `name` attribute's value:
    * If `name` is `LATENCY`, `threshold` represents the 95th percentile response time in milliseconds. Minimum `"50"`.
    * If `name` is `ERROR_RATE`, `threshold` represents the percentage (%) of failed web requests exceeding the threshold.
    An acceptable value would be `"0.42"`(42%). Minimum `"0.005"` (0.5%). Maximum `"1.0"` (100%).
* `sensitivity` - (Required) `<integer>` How many minutes the underlying formation alert metric must be at or above
  the threshold to trigger the alert. Acceptable values are as follows:
    * `1` - 'High'
    * `5` - 'Medium'
    * `10` - 'Low'
* `is_active` - (Required) `<boolean>` Whether to enable or disable the alert. Defaults to `true`.
* `notification_channels` - (Optional) `<list(string)>` Set to `['app']` if you wish to send email notifications
to all app members. Defaults to `[]`, which means no email notifications. Please also note the following:
    * It is strongly recommended not to define email addresses here due to Heroku requiring email verification.
      The API does not return unconfirmed email addresses when fetching for the formation alert. This means users
      may be faced with an infinite `plan` delta ('dirty plan') until the inputted email addresses are confirmed.
      Furthermore, this infinite `plan` delta may also occur if users manually enter email addresses via the UI.
    * Should users wish to use email addresses as notification channels, please do so via the UI and not set
    this attribute at all in their configurations.
* `notification_frequency` - (Optional) `<integer>` The frequency (in minutes) of email reminders for the formation alert
  that remains above the `threshold`. Acceptable values are as follows:
    * `5` - 'At most every 5 minutes'
    * `60` - 'At most every hour'
    * `1440` - 'At most every day'

## Attributes Reference

The following attributes are exported:

* `state` - The state of the formation alert.

## Import

An existing formation alert can be imported using the combination of the application UUID, process type,
and name each separated by a colon (':'). The name must be in all caps and acceptable values are either
`LATENCY` or `ERROR_RATE`.

For example:

```shell script
$ terraform import herokux_formation_alert.foobar1 "d54b26d4-a6e1-48a3-a71f-8bf833b82c04:web:LATENCY"
$ terraform import herokux_formation_alert.foobar2 "d54b26d4-a6e1-48a3-a71f-8bf833b82c04:web:ERROR_RATE"
```