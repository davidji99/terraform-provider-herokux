---
layout: "herokux"
page_title: "HerokuX: herokux_applink_connection"
description: |-
  Connects an existing Heroku AppLink add-on to a Salesforce org using JWT.
---

# herokux_applink_connection

Manage one Salesforce org connection on an existing Heroku AppLink add-on. Creating a connection with the write-only private key requires Terraform 1.11 or later. Configure the Salesforce client for JWT authentication, grant the Salesforce user the **Manage Heroku AppLink** permission, and use a Heroku identity with access to the app’s add-on, config vars, and SSO. [Heroku AppLink setup and permissions](https://devcenter.heroku.com/articles/working-with-heroku-applink)

## Example usage

```hcl
variable "salesforce_jwt_key" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "herokux_applink_connection" "production" {
  app_id             = heroku_app.production.id
  addon_id           = heroku_addon.applink.id
  name               = "production_org"
  client_id          = var.salesforce_client_id
  username           = var.salesforce_username
  jwt_private_key_wo = var.salesforce_jwt_key
}
```

Have your CI secret store set `TF_VAR_salesforce_jwt_key` to the multiline PEM on the runner for plan and apply. Terraform omits the ephemeral variable and write-only argument from saved plans and state. Heroku AppLink receives the PEM over HTTPS when creating the connection. [Terraform sensitive data](https://developer.hashicorp.com/terraform/language/manage-sensitive-data)

## Arguments

* `app_id` (required) is the UUID of the app through which to access the add-on. Switching to another app verifies its attachment and access to the same connection.
* `addon_id` (required) is the AppLink add-on UUID. Changing it replaces the connection.
* `name` (required) is the connection name in AppLink. Changing it replaces the connection.
* `client_id` and `username` are required for creation. Supply the Salesforce consumer key and Salesforce username.
* `jwt_private_key_wo` is required for creation. Supply the unencrypted RSA private key in PEM format (PKCS#1 or PKCS#8).
* `login_url` is the Salesforce HTTPS login URL for creation. The default is `https://login.salesforce.com`. For a sandbox, use `https://test.salesforce.com`.

The provider discovers the attachment alias and uses the AppLink service at `https://applink-api.heroku.com`. Credential changes take effect the next time Terraform creates the connection. To rotate credentials, supply the new values and explicitly replace it:

```shell
terraform apply -replace=herokux_applink_connection.production
```

Replacement disconnects the existing org before connecting again and can interrupt publishing. Destroy disconnects the named connection. An unavailable connection prevents publishing to that org. [Heroku connection lifecycle](https://devcenter.heroku.com/articles/working-with-heroku-applink#remove-your-connections)

## Attributes

* `connection_id` is the remote connection ID used to verify identity during refresh and destroy.
* `org_id`, `instance_url`, and `status` are the Salesforce org ID, instance URL, and status reported by AppLink.

## Import

Import a Salesforce connection by app UUID, add-on UUID, and connection name. Only these three arguments are needed in HCL for an existing connection. Supply the creation credentials before recreating it.

```shell
terraform import herokux_applink_connection.production '6588f886-4e0a-463d-adbc-efeedd0e0001:97d4b501-03da-4eec-955c-efeedd0e0002:production_org'
```

If another connection takes the same name, Terraform reports the remote ID mismatch. Verify the new connection in Heroku before removing the old state entry and importing its replacement.
