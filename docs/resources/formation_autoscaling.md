---
layout: "herokux"
page_title: "HerokuX: herokux_formation_autoscaling"
sidebar_current: "docs-herokux-resource-formation_autoscaling"
description: |-
  Provides a resource to manage the autoscaling settings of an app dyno formation.
---

# herokux\_formation\_autoscaling

This resource manages the autoscaling settings of an app dyno formation.
For more information about Heroku dyno formation autoscaling, please visit this [help article](https://devcenter.heroku.com/articles/scaling#autoscaling).

-> **IMPORTANT!**
When an existing `herokux_formation_autoscaling` is deleted, the provider will disable the autoscaling remotely
and remove the resource from state.

~> **WARNING:**
Please make sure you understand all [common issues](#common-issues) prior to using this resource. Failure to understand
them will result in potentially bricking your app dynos. You have been warned!

## Regarding `heroku_formation`

This resource can replace [`heroku_formation`](https://registry.terraform.io/providers/heroku/heroku/latest/docs/resources/formation)
if you are using dynos that can be autoscaled and wish to do so. Otherwise, continue using `heroku_formation`.
It is recommended NOT TO USE both resources concurrently.

Like `heroku_formation`, users will need to add the
[`depends_on`](https://www.terraform.io/docs/language/meta-arguments/depends_on.html) meta-argument
to `herokux_formation_autoscaling` when `heroku_app_release` is present. See the example [resource configuration](#example-usage) below.

## Common Issues

1. If you receive a `403` error during a `terraform apply`, it is likely you are trying to setup autoscaling
on an unsupported dyno type. Autoscaling is currently available only for Performance-tier dynos and dynos running in Private Spaces.

1. If you are migrating from using `heroku_formation` to `herokux_formation_autoscaling`, you can simply replace the former
with the latter ONLY IF the app dyno has never been autoscaled previously. Otherwise, follow the guidance below.

1. In the event you remove an existing `herokux_formation_autoscaling.foobar` resource after it's been successfully applied to an app,
   you will HAVE to `import` the resource first if the new `herokux_formation_autoscaling.foobar` resource is targeting
   the same app prior to its removal. This is due to two reasons:

    * The resource does not delete the formation autoscaling during resource destruction as it'll render any subsequent
      autoscaling operations an impossibility for the same dyno.

    * Due to the first reason, the underlying API does not allow for a `POST` request when an existing formation autoscaling
      exists in the API. Therefore, the resource must be imported first and then modified afterwards.

1. In the event you fail to comply with the aforementioned issue's guidance, the only solution is to delete the app
and start over.

## Example Usage

```hcl-terraform
resource "heroku_app" "foobar" {
  name   = "my-cool-app"
  region = "us"

  config_vars = {
    FOOBAR = "baz"
  }
}

resource "heroku_slug" "foobar" {
  app      = heroku_app.foobar.id
  file_url = "url_to_slug_artifact"

  process_types = {
    web = "ruby server.rb"
  }
}

resource "heroku_app_release" "foobar" {
  app = heroku_app.foobar.id
  slug_id = heroku_slug.foobar.id
}

resource "herokux_formation_autoscaling" "foobar" {
  app_id = heroku_app.foobar.uuid
  formation_name = "web"
  is_active = true
  min_quantity = 2
  max_quantity = 4
  desired_p95_response_time = 1001
  dyno_type = "performance-l"
  set_notification_channels = ["app"]

  # Tells Terraform that this formation autoscaling resource must be created/updated
  # only after the app release has been successfully.
  depends_on = ["heroku_app_release.foobar"]
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) `<string>` An existing app's UUID. The app name is not valid for this argument.

* `formation_name` - (Required) `<string>` The name of the dyno formation process, such as `web`.

* `is_active` - (Required) `<boolean>` Whether to enable or disable the autoscaling.

* `min_quantity` - (Required) `<integer>` Minimum dyno unit count. Must be at least 1.

* `max_quantity` - (Required) `<integer>` Max dyno unit count. Must be at least one number greater than `min_quantity`.

* `desired_p95_response_time` - (Required) `<integer>` Desired P95 Response Time in milliseconds. Must be at least 1ms.

* `dyno_type` - (Optional) `<string>` The type of dyno. (Example: “standard-1X”). Capitalization does not matter.
    - Use with caution if you already defined the dyno type in a `heroku_formation.size` resource attribute.
    Defining different values can lead to an infinite `plan` delta.

* `notification_channels` - (Optional) `<list(string)>` Channels you want to be notified if autoscaling occurs
for a dyno formation. The only currently valid value is `["app"]` or `[]`, which will turn on email notifications.

* `notification_period` - (Optional) `<integer>` Not sure what this does at the moment. Default value is `0`.

* `period` - (Optional) `<integer>` Not sure what this does at the moment, but the valid options are `1`, `5`, and `10`.
Default value is `1`.

## Attributes Reference

The following attributes are exported:

* `action_type`
* `operation`
* `quantity`

## Import

Existing formation autoscaling settings can be imported using the combination
of the application UUID, a colon, and the formation name.

For example:

```shell script
$ terraform import herokux_formation_autoscaling.foobar "d54b26d4-a6e1-48a3-a71f-8bf833b82c04:5f1091b8-eff5-4670-b1ad-20e980d24fc0"
```