---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_mtls"
sidebar_current: "docs-herokux-resource-postgres-mtls"
description: |-
  Provides a resource to manage the MTLS configuration for Heroku Private or Shield Postgres.
---

# herokux\_postgres\_mtls

This resource manages the MTLS configuration for Heroku Private or Shield Postgres database (version 10 or above).

### Initial Certificate
Upon successful MTLS provisioning, Heroku provisions a certificate ready for use by clients.
This initial certificate ID is exposed through the `initial_certificate_id` attribute.
Users could then do either of the following:

1. Use the data source `herokux_postgres_mtls_certificate` to retrieve the details of this certificate.
1. Import this certificate using the resource `herokux_postgres_mtls_certificate`. Once resource import is done,
   this certificate can now be managed via Terraform.

### Resource Timeouts
During creation and deletion, this resource checks the status of the MTLS provisioning or deprovisioning.
Both checks' default timeout is 10 minutes, which can be customized
via the `timeouts.mtls_provision_verify_timeout` and `timeouts.mtls_deprovision_verify_timeout` attributes in your `provider` block.

For example:

```hcl-terraform
provider "herokux" {
  timeouts {
    mtls_provision_verify_timeout = 15
    mtls_deprovision_verify_timeout = 15
  }
}
```

## Example Usage

```hcl-terraform
resource "heroku_space" "foobar" {
  name         = "foobar-space"
  organization = "my_org"
  region       = "virginia"
}

resource "heroku_app" "foobar" {
  name   = "my_foobar_app"
  region = "us"
  space  = heroku_space.foobar.name

  organization {
    name = "my_org"
  }
}

resource "heroku_addon" "database" {
  app_id  = heroku_app.foobar.id
  plan = "heroku-postgresql:private-0"
}

resource "herokux_postgres_mtls" "foobar" {
  database_name = heroku_addon.database.name
}
```

## Argument Reference

The following arguments are supported:

* `database_name` - (Required) `<string>` The name of the Postgres database name. DO NOT use the database UUID.
For example, 'postgresql-slippery-18242'.

## Attributes Reference

The following attributes are exported:

* `app_name` - The app which the postgres addon is tied to.
* `status` - The status of MTLS configuration.
* `enabled_by` - The Heroku user that enabled the MTLS configuration.
* `certificate_authority_chain` - the certificate authority chain. This attribute value does not get displayed in
logs or regular output.
* `initial_certificate_id` - The ID of the first certificate automatically created when MTLS is provisioned for a database.
Users will need to use the data source `herokux_postgres_mtls_certificate` to retrieve the certificate and private key.
The provider only sets this attribute on initial resource creation.

## Import

An existing database MTLS configuration can be imported using the database name.

For example:

```shell script
$ terraform import herokux_postgres_mtls.foobar "my_database_name"
```
