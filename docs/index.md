---
layout: "herokux"
page_title: "Provider: HerokuX"
sidebar_current: "docs-herokux-index"
description: |-
  Use the HerokuX provider to interact with the resources backed by undocumented Heroku APIs.
---

# HerokuX Provider

The HerokuX provider interacts with non-Platform APIs and Platform API variants
to provide additional resources not available
in the official [Heroku Terraform provider](https://github.com/heroku/terraform-provider-heroku).
All resources in this provider use the same API endpoints as the Heroku CLI and/or the web UI.

Overall, this provider's purpose is to supplement, not supplant, the official Heroku provider.

-> **IMPORTANT!**
This provider should be used with caution when Terraforming resources in environments that receive customer traffic.
Additionally, resources may change in behavior or be removed at any given time to match any API changes.

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

# Create a new Kafka topic
resource "herokux_kafka_topic" "topic_foobar" {
  # ...
}
```

## Authentication

The HerokuX provider offers a flexible means of providing credentials for authentication.
The following methods are supported, listed in order of precedence, and explained below:

* Static credentials
* Environment variables
* Netrc

### Static credentials

Credentials can be provided statically by adding an `api_key` arguments to the HerokuX provider block:

```hcl
provider "herokux" {
  api_key = var.heroku_api_key
}
```

Please note that if you go with the static credentials route, the `api_key` must be defined
or passed in as an input variable even if you did the same for the `heroku` provider.
The two providers do not implicitly share authentication when using static credentials.

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

### Netrc

Credentials can also be sourced from the [`.netrc`](https://everything.curl.dev/usingcurl/netrc)
file in your home directory:

```hcl
provider "herokux" {}
```

```shell
$ cat ~/.netrc
...
machine api.heroku.com
  login <your_heroku_email>
  password <your_heroku_api_key>
...
```

## Argument Reference

The following arguments are supported:

* `api_key` - (Required) Heroku API token. It must be provided, but it can also
  be sourced from [other locations](#Authentication).

* `metrics_api_url` - (Optional) Custom Metrics API url.
  Can also be sourced from the `HEROKUX_METRICS_API_URL` environment variable.

* `postgres_api_url` - (Optional) Custom Postgres API url.
  Can also be sourced from the `HEROKUX_POSTGRES_API_URL` environment variable.

* `data_api_url` - (Optional) Custom Data API url.
  Can also be sourced from the `HEROKUX_DATA_API_URL` environment variable.

* `platform_api_url` - (Optional) Custom Platform API url.
  Can also be sourced from the `HEROKU_API_URL` environment variable.

* `redis_api_url` - (Optional) Custom Redis API url.
  Can also be sourced from the `HEROKUX_REDIS_API_URL` environment variable.

* `connect_central_api_url` - (Optional) Custom Connect Central API url.
  Can also be sourced from the `HEROKUX_CONNECT_CENTRAL_API_URL` environment variable.

* `registry_api_url` - (Optional) Custom Registry API url.
  Can also be sourced from the `HEROKUX_REGISTRY_API_URL` environment variable.

* `kolkrabbi_api_url` - (Optional) Custom Kolkrabbi API url.
  Can also be sourced from the `HEROKUX_KOLKRABBI_API_URL` environment variable.

* `headers` - (Optional) Additional API headers.

* `delays` - (Optional) Delays define a given amount of time to wait before or after a resource takes an action.
This is to address scenarios where an underlying resource API does not report the status of a change
and subsequent changes require the previous one to be completed first.
Only a single `delays` block may be specified, and it supports the following arguments:

    * `postgres_settings_modify_delay` - (Optional) The number of minutes to wait for a postgres settings modification to be
      properly reflected in Heroku. Defaults to 2 minutes. Minimum required is 1 minute.

    * `connect_mapping_modify_delay` - (Optional) The number of seconds to wait Heroku Connect mapping to be
      properly reflected in Heroku. Defaults to 15 seconds. Minimum required is 5 seconds.

* `timeouts` - (Optional) Timeouts define a max duration the provider will wait for certain resources
  to be properly modified before proceeding with further action(s). Each timeout's polling intervals is set to 20 seconds.
  Only a single `timeouts` block may be specified, and it supports the following arguments:

    * `mtls_provision_timeout` - (Optional) The number of minutes to wait for a MTLS configuration
      to be provisioned on a database. Defaults to 10 minutes. Minimum required (based off of Heroku documentation) is 5 minutes.

    * `mtls_deprovision_timeout` - (Optional) The number of minutes to wait for a MTLS configuration
      to be deprovisioned from a database. Defaults to 10 minutes. Minimum required (based off of Heroku documentation) is 5 minutes.

    * `mtls_iprule_create_timeout` - (Optional) The number of minutes to wait for a MTLS IP rule
      to be created/authorized for a database. Defaults to 10 minutes.

    * `mtls_certificate_create_timeout` - (Optional) The number of minutes to wait for a MTLS certificate
      to be create and ready for use. Defaults to 10 minutes.

    * `mtls_certificate_delete_timeout` - (Optional) The number of minutes to wait for a MTLS certificate
      to be deleted. Defaults to 10 minutes.

    * `kafka_cg_create_timeout` - (Optional) The number of minutes to wait for a Kafka consumer group to be created.
      Defaults to 10 minutes.

    * `kafka_cg_delete_timeout` - (Optional) The number of minutes to wait for a Kafka consumer group to be deleted.
      Defaults to 10 minutes.

    * `kafka_topic_create_timeout` - (Optional) The number of minutes to wait for a Kafka topic to ready. Ready state
      is achieved when the topic itself is provisioned with the specified number of partitions.
      Defaults to 10 minutes. Minimum required is 3 minutes.

    * `kafka_topic_update_timeout` - (Optional) The number of minutes to wait for a Kafka topic to updated remotely.
      Defaults to 10 minutes. Minimum required is 3 minutes.

    * `privatelink_create_timeout` - (Optional) The number of minutes to wait for a privatelink to be provisioned.
      Defaults to 15 minutes. Minimum required is 5 minutes.

    * `privatelink_delete_timeout` - (Optional) The number of minutes to wait for a privatelink to be deprovisioned.
      Defaults to 15 minutes. Minimum required is 5 minutes.

    * `privatelink_allowed_acccounts_add_timeout` - (Optional) The number of minutes to wait for allowed accounts
      to become active for a privatelink. Defaults to 10 minutes. Minimum required is 2 minutes.

    * `data_connector_create_timeout` - (Optional) The number of minutes to wait for a data connector to be provisioned.
      Defaults to 20 minutes. Minimum required is 10 minutes.

    * `data_connector_delete_timeout` - (Optional) The number of minutes to wait for a data connector to be deleted.
      Defaults to 10 minutes. Minimum required is 3 minutes.

    * `data_connector_status_update_timeout` - (Optional) The number of minutes to wait for a data connector status to be updated.
      Defaults to 10 minutes. Minimum required is 5 minutes.

    * `data_connector_settings_update_timeout` - (Optional) The number of minutes to wait for a data connector settings to be updated.
      Defaults to 10 minutes. Minimum required is 5 minutes.

    * `postgres_credential_pre_create_verify_timeout` - (Optional) The number of minutes to wait for a postgres database
      to be available for credential creation. This is to address an edge scenario where one cannot create credentials immediately
      after a Premium, Private, or Shield postgres database is provisioned.
      Defaults to 45 minutes. Minimum required is 20 minutes.

    * `postgres_credential_create_timeout` - (Optional) The number of minutes to wait for a postgres credential to be created.
      Defaults to 10 minutes. Minimum required is 5 minutes.

    * `postgres_credential_delete_timeout` - (Optional) The number of minutes to wait for a postgres credential to be deleted.
      Defaults to 10 minutes. Minimum required is 5 minutes.

    * `shield_private_space_create_timeout` - (Optional) The number of minutes to wait for a shield private space
      to be provisioned. Defaults to 20 minutes. Minimum required is 10 minutes.
