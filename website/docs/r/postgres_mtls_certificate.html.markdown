---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_mtls_certificate"
sidebar_current: "docs-herokux-resource-postgres-mtls-certificate"
description: |-
  Provides a resource to manage a certificate for an existing MTLS enabled postgres database
---

# herokux\_postgres\_mtls\_certificate

This resource manages a certificate for an existing MTLS enabled postgres database. Certificates are valid for one year from the date of creation and cannot be extended beyond the aforementioned duration.

-> **IMPORTANT!**
Please be very careful when deleting this resource as any deleted certificates are NOT recoverable and invalidated immediately.
Furthermore, this resource renders the "private_key" attribute in plain-text in your state file.
Please ensure that your state file is properly secured and encrypted at rest.

### Resource Timeouts
During creation and deletion, this resource checks the status of the MTLS certificate creation or deletion.
Both checks' default timeout is ~10 minutes, which can be customized via
the `timeouts.mtls_certificate_create_timeout` and `timeouts.mtls_certificate_delete_timeout` in your `provider` block.

For example:
```hcl-terraform
provider "herokux" {
  timeouts {
    mtls_certificate_create_timeout = 15
    mtls_certificate_delete_timeout = 15
  }
}
```

### Why have separate resources for enabling MTLS and managing certificates?
Although the certificate API endpoint is a child of the MTLS endpoint, each certificate has its own UUID. Therefore, it is better
to have a certificate managed as a separate resource for optimal lifecycle management with terraform. If you have many certificates
to create, please utilize Terraform's `count` or `for_each` expression to keep your code DRY.

## Example Usage
```hcl-terraform
resource "herokux_postgres_mtls" "foobar" {
	database_name = "SOME_DATABASE_NAME"
}

resource "herokux_postgres_mtls_certificate" "foobar" {
	database_name = herokux_postgres_mtls.foobar.database_name
}
```

## Argument Reference

The following arguments are supported:

* `database_name` - (Required) `<string>` The name of the database. Please note the following:
    * DO NOT use the database UUID.
    * It is **highly recommended** setting this attribute's value to reference an existing `herokux_postgres_mtls` resource.
    This way, Terraform will handle the dependency chain between the two resources as you cannot create a certificate for
    a database that is not MTLS enabled.

## Attributes Reference

The following attributes are exported:

* `cert_id` - The UUID of the certificate. This is a separate attribute as the resource ID is a composite value.

* `name` - The name of certificate. It in the format of a hostname URL.

* `status` - The status of certificate.

* `expiration_date` - When the certificate expires in RFC822Z format.

* `private_key` - The client private key. This attribute value does not get displayed in logs or regular output.

* `certificate_with_chain` - The client certificate with chain. This attribute value does not get displayed in logs or regular output.

## Import

An existing database MTLS certificate can be imported using a composite value
of the database name and certificate ID separated by a colon.

For example:
```shell script
$ terraform import herokux_postgres_mtls_certificate.foobar "<MY_DB_NAME>:<CERT_ID>"
```
