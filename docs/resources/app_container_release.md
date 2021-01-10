---
layout: "herokux"
page_title: "HerokuX: herokux_app_container_release"
sidebar_current: "docs-herokux-resource-app-container-release"
description: |-
    Provides the ability to deploy a docker image to a Heroku application
---

# herokux\_app\_container\_release

This resource provides the ability to deploy a docker image to a Heroku application.

The specified docker image must be build and pushed to the Heroku Registry prior to using this resource.
For more information regarding building and pushing image(s),
please visit the [Container Registry & Runtime](https://devcenter.heroku.com/articles/container-registry-and-runtime#build-an-image-and-push) article.

-> **IMPORTANT!**
Please be advised that this resource will destroy the container (or dyno) on the application
if it is removed from a configuration.

## Example Usage

```hcl-terraform
resource "heroku_app" "foobar" {
  name   = "my-cool-app"
  region = "us"
}

resource "herokux_app_container_release" "foobar" {
	app_id = heroku_app.foobar.uuid
	image_id = "sha256:4d2647aab0e8fbe92cb0fc88c500eb51661c5907f4f14e79efe8bfbda1f7d159"
	process_type = "web"
}

# Update the web formation for the foobar application's web process type
resource "heroku_formation" "foobar-web" {
  app = heroku_app.foobar.name
  type = "web"
  quantity = 2
  size = "standard-2x"

  # Tells Terraform that the formation can only be updated after the container release is applied.
  depends_on = ["herokux_app_container_release.foobar"]
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) `<string>` The UUID of the app to subscribe to.

* `image_id` - (Required) `<string>` The `algorithm:hex` value of an already pushed docker image.
For more information regarding how to retrieve this value, visit [this article](https://devcenter.heroku.com/articles/container-registry-and-runtime#getting-a-docker-image-id).

* `process_type` - (Required) `<string>` Type of process such as "web".

## Attributes Reference

The following attributes are exported:

N/A

## Import

An existing docker release can be imported using a composite of the app UUID, image ID, and process type separated
by a pipe character (`|`).

For example:

```shell script
$ terraform import herokux_app_container_release.foobar "4d264cb9-d996-44f6-ba6d-e8e33a48a630|sha256:4d2647aab0e8fbe92cb0fc88c500eb51661c5907f4f14e79efe8bfbda1f7d159|web"
```
