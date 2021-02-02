---
layout: "herokux"
page_title: "Herokux: herokux_team"
sidebar_current: "docs-herokux-datasource-addons-x"
description: |-
  Get list of add-ons installed on a specific App.
---

# Data Source: herokux_addons

Use this data source to get a list of add-ons installed on a specific app.

## Example Usage

```hcl
data "herokux_addons" "foobar" {
  app_id = "MY_APP_ID"
  addon_service_name = "MY_ADDON_NAME"
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) The app ID

* `addon_service_name` - (Optional) Filter add-ons by service name (or include all add-ons if omitted). E.g. `heroku-postgresql`

## Attributes Reference

In addition to all arguments above, the following attribute is exported:

* `addons` - A map containing the add-ons UID and names as key-value pairs.
