---
layout: "herokux"
page_title: "Herokux: herokux_team"
sidebar_current: "docs-herokux-datasource-postgres-mtls-certificate-x"
description: |-
  Get information about a Heroku MTLS certificate
---

# Data Source: herokux_postgres_mtls_certificate

Use this data source to get information about a Heroku MTLS certificate.

## Example Usage

```hcl
data "herokux_postgres_mtls_certificate" "foobar" {
  database_name = "MY_DB_NAME"
  cert_id = "MY_CERT_ID"
}
```

## Argument Reference

The following arguments are supported:

* `database_name` - (Required) The database name

* `cert_id` - (Required) The certificate ID

## Attributes Reference

The following attributes are exported:

* `name` - The name of certificate. It in the format of a hostname URL.

* `status` - The status of certificate.

* `expiration_date` - When the certificate expires in RFC822Z format.

* `private_key` - The client private key. This attribute value does not get displayed in logs or regular output.

* `certificate_with_chain` - The client certificate with chain. This attribute value does not get displayed in logs or regular output.
