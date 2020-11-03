---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_mtls_iprule"
sidebar_current: "docs-herokux-resource-postgres-mtls-iprule"
description: |-
  Provides a resource to manage IP rules for an existing MTLS enabled postgres database
---

# herokux\_postgres\_mtls\_iprule

This resource manages IP rules for an existing MTLS enabled postgres database. There is a hard limit of 60 IP blocks that can be allowlisted per Postgres database.

-> **IMPORTANT!**
Deleting and re-adding the same CIDR range to the same MTLS enabled database may cause an unknown server error in Heroku.
Please wait a bit before attempting the action. The actual wait time is unknown at the moment.

### Resource Timeouts
During creation, this resource checks if the newly IP rule's status changes from 'Authorizing' to 'Authorized'.
This check's default timeout is ~10 minutes, which can be customized via the `timeouts.mtls_iprule_create_timeout` attribute
in your `provider` block.

For example:

```hcl-terraform
provider "herokux" {
  timeouts {
    mtls_iprule_create_timeout = 15
  }
}
```

### Reason for separate resources to manage MTLS and MTLS IP rules
Although the IP rule API endpoint is a child of the MTLS endpoint, each IP rule has its own UUID. Therefore, it is better
to have an IP rule managed as a separate resource for optimal lifecycle management with terraform. If you have a lot of IP rules
to add, please utilize Terraform's `count` or `for_each` expression to keep your code DRY.

## Example Usage

```hcl-terraform
resource "herokux_postgres_mtls" "foobar" {
	database_name = "SOME_DATABASE_NAME"
}

resource "herokux_postgres_mtls_iprule" "foobar" {
	database_name = herokux_postgres_mtls.foobar.database_name
	cidr = "1.2.3.4/32"
	description = "this is a test IP rule"
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

An existing database MTLS IP rule can be imported using a composite value of the database name and IP CIDR separated by a colon.

For example:

```shell script
$ terraform import herokux_postgres_mtls_iprule.foobar "<MY_DB_NAME>:1.2.3.4/32"
```
