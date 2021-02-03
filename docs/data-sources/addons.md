---
layout: "herokux"
page_title: "Herokux: herokux_addons"
sidebar_current: "docs-herokux-datasource-addons-x"
description: |-
  Get information about all add-ons installed on a specific App.
---

# Data Source: herokux_addons

Use this data source to get information about all add-ons installed on a specific app.

## Example Usage

```hcl
data "herokux_addons" "foobar" {
  app_id = "44e263f7-2e06-403b-9f37-44f0f9bcd5e9"
  addon_service_name = "heroku-postgresql"
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) The UUID of the app.

* `addon_service_name` - (Optional) Filter add-ons by service name (or include all add-ons if omitted).
  E.g. `heroku-postgresql`

## Attributes Reference

In addition to all arguments above, the following attribute is exported:

* `addons` - A map containing the add-ons UUID and names as key-value pairs.
