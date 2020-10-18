---
layout: "herokux"
page_title: "HerokuX: herokux_oauth_authorization"
sidebar_current: "docs-herokux-resource-oauth-authorization"
description: |-
  Provides a resource to manage an OAuth Authorization for a Heroku user account.
---

# herokux\_oauth\_authorization

This resource manages an [oauth authorization](https://devcenter.heroku.com/articles/oauth#direct-authorization) in Heroku.
You can use these access tokens obtained with OAuth authorization to grant access for your own scripts on your machine
or to other applications. Generated access tokens can be non-expiring or short-lived with varied scopes.

-> **IMPORTANT!**
Please be very careful when deleting this resource as any deleted authorizations are NOT recoverable and invalidated immediately.
Furthermore, this resource renders the `access_token` attribute in plain-text in your state file.
Please ensure that your state file is properly secured and encrypted at rest.

### Rotating an existing authorization access token
If you wish to rotate an existing `access_token` created by this resource, the recommended way is to `taint` the resource
and then execute `terraform apply`. This will generate a new authorization and access token.

**DO NOT USE** `heroku authorizations:rotate <ID>` or its underlying [API](https://devcenter.heroku.com/articles/platform-api-reference#oauth-authorization-regenerate)
as the new access token's time to live is set to 28880 seconds (~8 hours) regardless of the original TTL. This out-of-band
TTL change does not reflect an existing resource configuration TTL and will likely lead to configuration drift.

### Expired authorizations
Heroku deletes an expired authorization completely from their systems, making it unavailable via the Platform API.
In this scenario, this resource will cause the `terraform apply` command to exit with an error
indicating an expired authorization. Users will then need to execute `terraform taint [options] <address>` on the resource
and `apply` again.

## Example Usage

```hcl-terraform
resource "herokux_oauth_authorization" "foobar" {
	scope = ["read"]
	auth_api_key_name = "MYBOTUSER"
	time_to_live = 100000
	description = "This is an oauth authorization test from Terraform"
}
```

## Argument Reference

The following arguments are supported:

* `scope` - (Required) `<list(string)>` Set custom OAuth scopes. Valid scopes are:
    * `global` - Read and write access to all of your account, apps and resources.
    [Equivalent to the default authorization obtained when using the CLI](https://devcenter.heroku.com/articles/authentication).
    * `identity` - Read-only access to your [account information](https://devcenter.heroku.com/articles/platform-api-reference#account).
    * `read` & `write` - Read and write access to all of your apps and resources, excluding account information and configuration variables.
    This scope lets you request access to an account without necessarily getting access to runtime secrets such as database connection strings.
    * `read-protected` & `write-protected` - Read and write access to all of your apps and resources, excluding account information.
    This scope lets you request access to an account including access to runtime secrets such as database connection strings.

* `auth_api_key_name` - `<string>` A name representing an existing API key for a Heroku user account.
Setting this attribute allows an oauth authorization to be created in a user account that's different from the account used
initially to authenticate with the provider. This attribute's value will then replace the `%s` in `HEROKUX_%s_API_KEY`.
For example, if the attribute value is `myBotUser_X`, you will need to have `HEROKUX_MYBOTUSER_X_API_KEY` defined in the environment.
Please also note the following:
    * If this attribute is not set, the resource will create the oauth authorization in the same account
    used to authenticate with the provider.
    * A value may only include words, letters, or underscore with a max length of 32 characters. Case-insensitive.
    * Each `herokux_oauth_authorization` resource can define a unique value for this attribute. However, this translates
    to an equal number of equivalent variables define in the environment.

* `time_to_live` - `<integer>` Set expiration in seconds. No expiration if attribute is not set in the configuration.

* `description` - `<string>` Set a custom authorization description.

## Attributes Reference

The following attributes are exported:

* `access_token` - The access token. This attribute value does not get displayed in logs or regular output.

* `expires_in` - How long (in seconds) before the access token will be expired.
If there is no expiration date, this attribute value will be `0`.

* `token_id` - The ID of the token. This differs from the resource ID, which is the authorization ID.

## Import

An existing oauth authorization can be imported in two different ways.

If the resource specifies an `auth_api_key_name` attribute value, the import ID is a composite value
of the authorization ID, time to live and `auth_api_key_name` value separated by a colon:

```shell script
$ terraform import herokux_data_connector.foobar "09071f30-0e82-11eb-adc1-0242ac120002:8600000:HBOTUSER"
```

If the resource does not specify an `auth_api_key_name` attribute value, the import ID is a composite value
of the authorization ID and time to live value separated by a colon:

```shell script
$ terraform import herokux_data_connector.foobar "09071f30-0e82-11eb-adc1-0242ac120002:8600000"
```
