---
layout: "herokux"
page_title: "HerokuX: herokux_formation_autoscaling"
sidebar_current: "docs-herokux-resource-formation_autoscaling"
description: |-
  Provides a resource to manage the autoscaling settings of an app dyno formation.
---

# herokux\_formation\_autoscaling

This resource manages the autoscaling settings of an app dyno formation.
For more information about Heroku dyno formation scaling, please visit this [help article](https://devcenter.heroku.com/articles/scaling#autoscaling).

Autoscaling is currently available only for Performance-tier dynos and dynos running in Private Spaces.
Heroku’s auto-scaling uses response time which relies on your application to have very small variance in response time.
If your application does not, then you may want to consider a third-party add-on such as Rails Auto Scale
which scales based on queuing time instead of overall response time. Scaling limits are also between apps
in Private Spaces and Common Runtime.

This resource can replace  [`heroku_formation`](https://registry.terraform.io/providers/heroku/heroku/latest/docs/resources/formation).
It is recommended NOT TO USE both resources concurrently. Furthermore, like `heroku_formation`, users will need
to add the [`depends_on`](https://www.terraform.io/docs/language/meta-arguments/depends_on.html) meta-argument
to `herokux_formation_autoscaling` when `heroku_app_release` is present.
See the example [resource configuration](#example-usage) below.

-> **IMPORTANT!**
Due to API limitations, the provider will only remove the resource from state if you remove an existing
`herokux_formation_autoscaling` from your terraform configuration. You will then need to visit the Heroku UI for further action.

## Example Usage

```hcl-terraform
resource "heroku_app_release" "foobar-release" {
  app = "SOME_APP"
  slug_id = "01234567-89ab-cdef-0123-456789abcdef"
}

resource "herokux_formation_autoscaling" "foobar" {
  app_id = "d54b26d4-a6e1-48a3-a71f-8bf833b82c04"
  formation_name = "web"
  is_active = true
  min_quantity = 2
  max_quantity = 4
  desired_p95_response_time = 1001
  dyno_type = "performance-l"
  set_notification_channels = ["app"]

  # Tells Terraform that this formation autoscaling resource must be created/updated
  # only after the app release has been created.
  depends_on = ["heroku_app_release.foobar-release"]
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
for a dyno formation. The only currently valid value is `["app"]`, which will turn on email notifications.

* `notification_period` - (Optional) `<integer>` Not sure what this does at the moment, but the default value is `0`.

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