---
layout: "herokux"
page_title: "HerokuX: herokux_pipeline_member"
sidebar_current: "docs-herokux-resource-pipeline-member"
description: |-
Provides a resource to manage membership for a Pipeline.
---

# herokux_pipeline_member

This resource manages a Heroku user's membership to a pipeline. A user's pipeline membership only controls access to the
pipeline's ephemeral apps such as review and CI apps. Please visit [this article](https://devcenter.heroku.com/articles/pipelines#permissions-and-capabilities)
for more information regarding each permission's capabilities.

If you find yourself needing to constantly add users with “member” role in the Enterprise Teams and Heroku Teams,
consider using the `herokux_pipeline_ephemeral_apps_permission` resource.

## Example Usage

```hcl-terraform
// First, create a pipeline.
resource "heroku_pipeline" "foobar" {
  name = "foobar-pipeline"
}

resource "herokux_pipeline_member" "admin" {
  pipeline_id = heroku_pipeline.foobar.id
  email = "admin@mycompany.com"
  permissions = ["view", "operate", "manage", "operate"]
}
```

## Argument Reference

The following arguments are supported:

* `pipeline_id` - (Required) `<string>` The UUID for a Heroku pipeline.
* `email` - (Required) `<string>` The email address of a Heroku user.
* `permissions` - (Required) `<list(string)>` What permissions to grant to the Heroku user. Acceptable permissions are
  `view`, `operate`, `deploy`, and `manage`. At least one permission is required.
    * Please note that the `view` is always set even if not explicitly defined for this attribute. Therefore, it is
      recommended to define the `view` permission in your configuration to match what permissions will be set remotely.

## Attributes Reference

The following attributes are exported:

N/A

## Import

An existing pipeline member can be imported using a composite value of the pipeline ID and email address separated
by a colon.

For example:

```shell script
$ terraform import herokux_pipeline_member.admin "2508ebbd-74bb-4e81-a63c-d193d2bd5716:admin@mycompany.com"
```
