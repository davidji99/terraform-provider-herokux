---
layout: "herokux"
page_title: "HerokuX: herokux_pipeline_ephemeral_apps_permission"
sidebar_current: "docs-herokux-resource-pipeline-ephemeral-apps-permission"
description: |-
Provides a resource to manage the Ephemeral Apps permissions for Heroku pipeline.
---

# herokux_pipeline_ephemeral_apps_permission

This resource manages the [Ephemeral Apps permissions](https://devcenter.heroku.com/articles/pipelines#ephemeral-app-permissions),
specifically the auto-join functionality, for a Heroku Pipeline.

Deleting this resource from an existing configuration will turn off the auto-join functionality.

-> **IMPORTANT!**
Any changes to permissions will only be applied to new apps from the time you make those changes.
The changes do not apply existing apps.

## Example Usage

```hcl-terraform
// First, create a pipeline.
resource "heroku_pipeline" "foobar" {
  name = "foobar-pipeline"
}

resource "herokux_pipeline_ephemeral_apps_permission" "foobar" {
  pipeline_id = heroku_pipeline.foobar.id
  permissions = ["view", "operate", "manage"]
}
```

## Argument Reference

The following arguments are supported:

* `pipeline_id` - (Required) `<string>` The UUID for a Heroku pipeline.
* `permissions` - (Required) `<list(string)>` What permissions all users with “member” role in the Enterprise Teams and Heroku Teams
  should be automatically granted for the pipeline's ephemeral apps. Acceptable permissions are `view`, `operate`,
  `deploy`, and `manage`. At least one permission is required.
  * Please note that the `view` is always set even if not explicitly defined for this attribute. Therefore, it is
    recommended to define the `view` permission in your configuration to match what permissions will be set remotely.

## Attributes Reference

The following attributes are exported:

* `id` - The pipeline ID.
* `pipeline_name` - Name of the pipeline.
* `owner_id` - The Heroku user ID that owns the pipeline.

## Import

An existing pipeline Ephemeral Apps permission can be imported using the pipeline UUID.

For example:

```shell script
$ terraform import herokux_pipeline_ephemeral_apps_permission.foobar "2508ebbd-74bb-4e81-a63c-d193d2bd5716"
```
