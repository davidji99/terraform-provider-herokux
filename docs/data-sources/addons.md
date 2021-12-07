---
layout: "herokux"
page_title: "Herokux: herokux_addons"
sidebar_current: "docs-herokux-datasource-addons-x"
description: |-
  Get information about all add-ons.
---

# Data Source: herokux_addons

Use this data source to get information about all add-ons that the provider's authenticated
user has access to in Heroku.

## Example Usage

```hcl-terraform
// Retrieve all addons
data "herokux_addons" "all" {}

// Filter by app name
data "herokux_addons" "by_app_name" {
  app_name_regex = ".*api-stg.*"
}

// Filter by addon name
data "herokux_addons" "by_addon_name" {
  addon_name_regex = "scheduler-.*"
}
```

## Argument Reference

The following arguments are supported:

* `app_name_regex` - (Optional) Valid regex used to filter by app name.
  Cannot be set in conjunction with `addon_name_regex`.
* `addon_name_regex` - (Optional) Valid regex used to filter by addon name.
  Cannot be set in conjunction with `app_name_regex`.

## Attributes Reference

In addition to all arguments above, the following attribute is exported:

* `addons` - List of maps containing addon information. Each element in this attribute is defined as follows:
    * `app_id` - App UUID
    * `app_name` - App name
    * `name` - Addon name
    * `state` - State of the addon
