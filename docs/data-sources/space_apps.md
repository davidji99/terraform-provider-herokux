---
layout: "herokux"
page_title: "Herokux: herokux_space_apps"
sidebar_current: "docs-herokux-datasource-space-apps-x"
description: |-
  Get information about all apps in a Private space.
---

# Data Source: herokux_space_apps

Use this data source to get information about all apps in a Private space.

## Example Usage

```hcl-terraform
data "herokux_space_apps" "apps" {
  space_regex = "prod-.*"
}
```

## Argument Reference

The following arguments are supported:

* `space_regex` - (Required) Valid regex used to filter apps by space name or UUID.

## Attributes Reference

In addition to all arguments above, the following attribute is exported:

* `apps` - List of maps containing app information. Each element in this attribute is defined as follows:
    * `id` - App UUID
    * `name` - App name
    * `web_url` - App web URL
    * `stack` - App stack
    * `region` - App region
