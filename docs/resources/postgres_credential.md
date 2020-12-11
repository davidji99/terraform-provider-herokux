---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_credential"
sidebar_current: "docs-herokux-resource-postgres-credential"
description: |-
  Provides a resource to manage a credential of a postgres database
---

# herokux\_postgres\_credential

This resource manages credentials for a Heroku postgres database. [Credentials](https://devcenter.heroku.com/articles/heroku-postgresql-credentials)
represent Heroku's management layer around postgres roles. Each credential corresponds to a different
Postgres role and its specific set of database privileges.

Credentials created by this resource do not have any permissions by default. To grant permissions to a credential,
please use the Heroku CLI (`heroku pg:psql`) or visit the UI for further action.

-> **IMPORTANT!**
Please be very careful when deleting this resource as any deleted credentials are NOT recoverable and invalidated immediately.
Furthermore, this resource renders the `secrets.username` & `secrets.password` attributes in plain-text in your state file.
Please ensure that your state file is properly secured and encrypted at rest.

### Resource Timeouts
During creation and deletion, this resource checks the status of the credential. All the aforementioned timeouts
can be customized via the `timeouts.postgres_credential_create_timeout` and
`timeouts.postgres_credential_delete_timeout` attributes in your `provider` block.

For example:

```hcl-terraform
provider "herokux" {
  timeouts {
    postgres_credential_create_timeout = 20
    postgres_credential_delete_timeout = 20
  }
}
```

## Example Usage

```hcl-terraform
resource "herokux_postgres_credential" "foobar" {
	postgres_id = "2508ebbd-74bb-4e81-a63c-d193d2bd5716"
	name = "read-only-credential"
}
```

## Argument Reference

The following arguments are supported:

* `postgres_id` - (Required) `<string>` The UUID of a Heroku postgres addon.

* `name` - (Required) `<string>` Name of the credential. Credential names are restricted to alphanumeric characters
(`-` and `_` are supported) and cannot be longer than 50 characters. Names are not an updatable attribute and will
force and destroy and create flow if changed.

## Attributes Reference

The following attributes are exported:

* `state` - The state of credential.

* `database` - The name of the database that the credential belongs to.

* `host` - The database host URL. This attribute value does not get displayed in logs or regular output.

* `port` - The database port number. This attribute value does not get displayed in logs or regular output.

* `secrets` - List of maps of usernames and passwords for the credential. By default, there will be always be at least
one set of a username and password. This attribute value does not get displayed in logs or regular output.

    * `username` - The username. This attribute value does not get displayed in logs or regular output.
    * `password` - The password. This attribute value does not get displayed in logs or regular output.
    * `state` - The state of the secret.

* `uuid` - The UUID for the credential.

## Import

An existing credential can be imported using a composite value
of the postgres ID and credential name separated by a colon.

For example:

```shell script
$ terraform import herokux_postgres_credential.foobar "2508ebbd-74bb-4e81-a63c-d193d2bd5716:read-only-credential"
```

**Please Note:** DO NOT import the 'default' credential provisioned with every new Heroku postgres database.
Heroku does not allow you to destroy this credential, so it will not be possible manage it via Terraform.
