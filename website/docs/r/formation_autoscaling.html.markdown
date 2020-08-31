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

-> **IMPORTANT!**
Autoscaling is currently available only for Performance-tier dynos and dynos running in Private Spaces.
Heroku’s auto-scaling uses response time which relies on your application to have very small variance in response time.
If your application does not, then you may want to consider a third-party add-on such as Rails Auto Scale
which scales based on queuing time instead of overall response time. Scaling limits are also different for apps in Private Spaces
and apps in the Common Runtime.

## Example Usage

```hcl-terraform
resource "herokux_formation_autoscaling" "foobar" {
	app_id = "SOME_APP_ID"
	formation_name = "web"
	is_active = true
	min_quantity = 2
	max_quantity = 4
	desired_p95_response_time = 1001
	dyno_type = "performance-l"
	set_notification_channels = ["app"]
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) `<string>` An existing app's UUID. The app name is not valid for this argument.

* `formation_name` - (Required) `<string>` The name of the dyno formation process. A common name would be `web`.

* `is_active` - (Required) `<boolean>` Whether to enable or disable the autoscaling.

* `min_quantity` - (Required) `<number>` Minimum dyno unit count. Must be at least 1.

* `max_quantity` - (Required) `<number>` Max dyno unit count. Must be at least one number greater than `min_quantity`.

* `desired_p95_response_time` - (Required) `<number>` Desired p95 Response Time in milliseconds. Must be at least 1ms.

* `dyno_type` - (Optional) `<string>` The type of dyno. (Example: “standard-1X”). Capitalization does not matter.
    - Use with caution if you already defined the dyno type in a `heroku_formation.size` resource attribute.
    Defining different values can lead to an infinite `plan` delta.

* `set_notification_channels` - (Optional) `<list(string)>` Channels you want to be notified if autoscaling occurs
for a dyno formation. The only currently valid value is `["app"]`, which will turn on email notifications.

* `period` - (Optional) `<number>` Not sure what this does at the moment but the valid options are `1`, `5`, and `10`.
Default value is `1`.

## Attributes Reference

The following attributes are exported:

* `action_type`
* `operation`
* `quantity`

## Import

Existing formation autoscaling settingss can be imported using the combination
of the application UUID, a colon, and the formation name.

For example:
```shell script
$ terraform import herokux_formation_autoscaling.foobar <APP_ID>:<FORMATION_NAME>
```