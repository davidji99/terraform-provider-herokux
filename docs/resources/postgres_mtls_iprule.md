---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_mtls_iprule"
sidebar_current: "docs-herokux-resource-postgres-mtls-iprule"
description: |-
  Provides a resource to manage MTLS IP rules for Heroku Private or Shield Postgres.
---

# herokux\_postgres\_mtls\_iprule

This resource manages MTLS IP rules for a Heroku Private or Shield Postgres database. There is a hard limit of 60 IP blocks
that can be allowlisted per database.

-> **IMPORTANT!**
Deleting and re-adding a MTLS IP rule using the same CIDR range in succession may cause an unknown server error
in Heroku. Please wait a bit after destruction before attempting to recreate the MTLS IP rule.
The actual wait time is unknown at the moment.

### Resource Timeouts
During creation, this resource verifies if the MTLS IP rule status successfully changes from 'Authorizing' to 'Authorized'.
This check's default timeout is ~10 minutes, which can be customized via the `timeouts.mtls_iprule_create_verify_timeout`
attribute in your `provider` block.

For example:

```hcl-terraform
provider "herokux" {
  timeouts {
    mtls_iprule_create_verify_timeout = 15
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

resource "herokux_postgres_mtls_iprule" "foobar" {
  database_name = herokux_postgres_mtls.foobar.database_name
  cidr          = "1.2.3.4/32"
  description   = "CI/CD outbound IPs"
}
```

## Argument Reference

The following arguments are supported:

* `database_name` - (Required) `<string>` The name of the database. Please note the following:
    * DO NOT use the database UUID.
    * It is **highly recommended** setting this attribute's value to reference an existing `herokux_postgres_mtls` resource.
      This way, Terraform will handle the dependency chain between the two resources as you cannot create an IP rule for
      a database that is not MTLS enabled.
* `cidr` - (Required) `<string>` Valid IPv4 CIDR value. Example: `1.2.3.4/32`.
* `description` - (Optional) `<string>` A description of the IP rule.

## Attributes Reference

The following attributes are exported:

* `rule_id` - The UUID of the rule. This is a separate attribute as the resource ID is a composite value.
* `status` - The status of IP rule configuration.

## Import

An existing Postgres MTLS IP rule can be imported using a composite value of the database name and IP CIDR separated
by a colon.

For example:

```shell script
$ terraform import herokux_postgres_mtls_iprule.foobar "my_database_name:1.2.3.4/32"
```
