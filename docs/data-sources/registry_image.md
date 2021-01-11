---
layout: "herokux"
page_title: "Herokux: herokux_registry_image"
sidebar_current: "docs-herokux-datasource-registry-image"
description: |-
    Get information about a Heroku registry image.
---

# Data Source: herokux_registry_image

Use this data source to get information about a Heroku registry image.
The image must be [pushed](https://devcenter.heroku.com/articles/container-registry-and-runtime#building-and-pushing-image-s)
before it can be retrieved by this data source.

## Example Usage

```hcl-terraform
resource "heroku_app" "foobar" {
  name   = "my-cool-app"
  region = "us"
}

# Push image to Heroku Registry

data "herokux_registry_image" "foobar" {
  app_id = heroku_app.foobar.uuid
  process_type = "web"
  docker_tag = "latest"
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) `<string>` The UUID of the app to subscribe to.

* `process_type` - (Required) `<string>` Type of process such as "web".

* `docker_tag` - (Optional) `<string>` Docker tag. Defaults to `latest`, which should be valid most of the cases.

## Attributes Reference

The following attributes are exported:

* `size` - Total size of the Docker image.

* `schema_version` - Docker registry version.

* `digest` - A [hash of a Docker image](https://www.mikenewswanger.com/posts/2020/docker-image-digests/) supported
  by the Docker v2 registry format. This hash is generated as sha256 and is deterministic based on the image build.
  This means that so long as the Dockerfile and all build components (base image selected in FROM, any files downloaded
  or copied into the image, etc) are unchanged between builds, the built image will always resolve to the same digest.
  This is important as a change in digest indicates that something changed in the image.

  - This computed attribute is handy for the `herokux_app_container_release.image_id` attribute.

* `number_of_layers` - Number of layers that exist for the docker image.