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
them will result in potentially bricking your app dynos.

## Regarding `heroku_formation`

This resource can be used with [`heroku_formation`](https://registry.terraform.io/providers/heroku/heroku/latest/docs/resources/formation)
if you are using dyno sizes that can be autoscaled and wish to do so. Otherwise, continue using just `heroku_formation`.
If you end up using `heroku_formation` in conjunction with `herokux_formation_autoscaling`, do not make any changes to
`heroku_formation` as those changes will not be reflected on the app dyno if autoscaling is enabled.

Users will need to add the [`depends_on`](https://www.terraform.io/docs/language/meta-arguments/depends_on.html) meta-argument
to `herokux_formation_autoscaling` when `heroku_app_release` and/or `heroku_formation` are present. `heroku_formation`
does not need to be in `herokux_formation_autoscaling.depends_on` if `herokux_formation_autoscaling.process_type` is set
to `heroku_formation.foobar.type`.

See the example [resource configuration](#example-usage) below on how to use `heroku_formation` with `herokux_formation_autoscaling`.

## Common Issues

1. If you receive a `403` error during a `terraform apply`, it is likely you are trying to setup autoscaling
on an unsupported dyno type. Autoscaling is currently available only for Performance-tier dynos and dynos running in Private Spaces.

1. In the event you remove an existing `herokux_formation_autoscaling.foobar` resource after it's been successfully applied to an app,
   you **MUST** `import` the resource first if the new `herokux_formation_autoscaling.foobar` resource is targeting
   the same app prior to its removal. Otherwise, the resource will error with a message indicating a resource `import` prerequisite.
   This is due to following reasons:

    * The resource does not delete the formation autoscaling during resource destruction as it'll render any subsequent
      autoscaling management impossible for the same app/dyno process type.

    * Due to the above reason, the underlying API does not allow for a `POST` request when an existing formation autoscaling
      exists for the app. Therefore, the resource must be imported first and then modified afterwards. If you do not first `import`
      the resource and are using a version of the provider less than `v0.20.3`, the provider will surface a `409 Conflict` error
      on resource creation.

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

resource "heroku_formation" "foobar" {
  app = heroku_app.foobar.id
  type = var.process_type
  quantity = 8
  size = var.dyno_size

  # Tells Terraform that this formation must be created/updated only after the app release has been created
  depends_on = [heroku_app_release.foobar-release]
}

resource "herokux_formation_autoscaling" "foobar" {
  app_id = heroku_app.foobar.uuid
  process_type = heroku_formation.foobar.type
  is_active = true
  min_quantity = 7
  max_quantity = 9
  desired_p95_response_time = 1001
  dyno_size = var.dyno_size
  notification_channels = ["app"]

  # Tells Terraform that this formation autoscaling resource must be created/updated
  # only after the app release has been successfully applied.
  depends_on = [heroku_app_release.foobar]
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) `<string>` An existing app's UUID. The app name is not valid for this argument.
* `process_type` - (Required) `<string>` The type of the dyno formation process, such as `web`.
* `is_active` - (Required) `<boolean>` Whether to enable or disable the autoscaling.
* `min_quantity` - (Required) `<integer>` Minimum dyno unit count. Must be at least 1.
* `max_quantity` - (Required) `<integer>` Max dyno unit count. Must be at least one number greater than `min_quantity`.
* `desired_p95_response_time` - (Required) `<integer>` Desired P95 Response Time in milliseconds. Must be at least 1ms.
* `dyno_size` - (Required) `<string>` The size of dyno. (Example: “performance-l”). Capitalization does not matter.
Only specify dyno sizes that can be autoscaled. You can only modify dynos of the same type.
* `notification_channels` - (Optional) `<list(string)>` Channels you want to be notified if autoscaling occurs
for a dyno formation. The only currently valid value is `["app"]` or `[]`, which will turn on email notifications.
* `notification_period` - (Optional) `<integer>` Not sure what this does at the moment. Default value is `0`.

## Attributes Reference

The following attributes are exported:

* `action_type` - Type of formation autoscaling.
* `operation` - Operation of the formation autoscaling.
* `quantity` - Number of dynos

## Import

Existing formation autoscaling settings can be imported using the combination
of the application UUID, a colon, and the process type.

For example:

```shell script
$ terraform import herokux_formation_autoscaling.foobar "d54b26d4-a6e1-48a3-a71f-8bf833b82c04:web"
```
