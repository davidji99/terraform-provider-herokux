---
layout: "herokux"
page_title: "Herokux: herokux_team"
sidebar_current: "docs-herokux-datasource-postgres-mtls-certificate-x"
description: |-
  Get information about a Heroku MTLS certificate
---

# Data Source: herokux_postgres_mtls_certificate

Use this data source to get information about a Heroku MTLS certificate.

-> **IMPORTANT!**
This data source renders the certificate "private_key" attribute in plain-text in your state file. Please ensure that your state file is properly secured and encrypted at rest.

## Example Usage

```hcl-terraform
data "heroku_addon" "database" {
  name = "postgres-fitted-123"
}

data "herokux_postgres_mtls_certificate" "foobar" {
  database_name = heroku_addon.database.name
  cert_id = "1d17bd09-6ad2-4a39-b50a-e02e467f5ee2"
}
```

## Argument Reference

The following arguments are supported:

* `database_name` - (Required) The database name.
* `cert_id` - (Required) The certificate UUID.

## Attributes Reference

The following attributes are exported:

* `name` - The name of certificate. It in the format of a hostname URL.
* `status` - The status of certificate.
* `expiration_date` - When the certificate expires in RFC822Z format.
* `private_key` - The client private key. This attribute value does not get displayed in logs or regular output.
* `certificate_with_chain` - The client certificate with chain. This attribute value does not get displayed in logs or regular output.
