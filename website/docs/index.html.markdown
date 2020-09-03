---
layout: "herokux"
page_title: "Provider: HerokuX"
sidebar_current: "docs-herokux-index"
description: |-
  Use the HerokuX provider to interact with the resources backed by undocumented Heroku APIs.
---

# HerokuX Provider

The HerokuX provider interacts with undocumented Heroku APIs to provide additional resources not available
in the official [Heroku Terraform provider](https://github.com/heroku/terraform-provider-heroku).

-> **IMPORTANT!**
This provider should be treated as experimental and to be used with caution when terraforming resources in environments
that receive customer traffic. Additionally, the resources may change in behavior at any given time to match any API changes.

## Contributing

Development happens in the [GitHub repo](https://github.com/davidji99/terraform-provider-herokux):

* [Releases](https://github.com/davidji99/terraform-provider-herokux/releases)
* [Issues](https://github.com/davidji99/terraform-provider-herokux/issues)

## Example Usage

```hcl
# Configure the HerokuX provider
provider "herokux" {
  # ...
}

# Create a new project
resource "herokux_formation_autoscaling" "service-x" {
  # ...
}
```

## Authentication

The HerokuX provider offers a flexible means of providing credentials for authentication.
The following methods are supported, listed in order of precedence, and explained below:

- Static credentials
- Environment variables

### Static credentials

Credentials can be provided statically by adding an `api_key` arguments to the HerokuX provider block:

```hcl
provider "herokux" {
  api_key = var.heroku_api_key
}
```

### Environment variables

When the HerokuX provider block does not contain an `api_key` argument, the missing credentials will be sourced
from the environment via the `HEROKU_API_KEY` environment variables respectively:

```hcl
provider "herokux" {}
```

```shell
$ export HEROKU_API_KEY="SOME_KEY"
$ terraform plan
Refreshing Terraform state in-memory prior to plan...
```

In order to prevent duplicate environment variables, the HerokuX provider uses the same environment variable name
as the Heroku provider to retrieve the API key. This will be the only common variable name between the two providers.

## Argument Reference

The following arguments are supported:

* `api_key` - (Required) Heroku API token. It must be provided, but it can also
  be sourced from [other locations](#Authentication).

* `metrics_api_url` - (Optional) Custom Metrics API url

* `postgres_api_url` - (Optional) Custom Postgres API url

* `headers` - (Optional) Additional API headers.

* `timeouts` - (Optional) Timeouts help certain resources to be properly created or deleted before proceeding with further actions.
Only a single `timeouts` block may be specified and it supports the following arguments:

  * `mtls_provision_timeout` - (Optional) The number of minutes to wait for an MTLS configuration to be provisioned on a database.

  * `mtls_deprovision_timeout` - (Optional) The number of minutes to wait for an MTLS configuration to be deprovisioned from a database.

  * `mtls_iprule_create_timeout` - (Optional) The number of minutes to wait for an MTLS IP rule to be created/authorized for a database.