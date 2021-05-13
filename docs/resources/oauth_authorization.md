---
layout: "herokux"
page_title: "HerokuX: herokux_oauth_authorization"
sidebar_current: "docs-herokux-resource-oauth-authorization"
description: |-
  Provides a resource to manage an OAuth Authorization for a Heroku user account.
---

# herokux\_oauth\_authorization

This resource manages an [OAuth authorization](https://devcenter.heroku.com/articles/oauth#direct-authorization) in Heroku.
Currently, this resource only supports managing an OAuth direction authorization.

You can use the access token created from this resource to grant access for your own scripts on your machine
or to other applications. These access tokens can be varied in scope and non-expiring or short-lived.

-> **IMPORTANT!**
Please be very careful when deleting this resource as the deleted authorization is NOT recoverable and invalidated immediately.
Furthermore, this resource renders the `access_token` attribute in plain-text in your state file.
Please ensure that your state file is properly secured and encrypted at rest.

### Rotating an existing authorization access token
If you wish to rotate an existing `access_token` created by this resource, the recommended way is to `taint` the resource
and then execute `terraform apply` **ONLY** if the authorization is still valid and has not expired.
This will generate a new authorization and access token.

~> **WARNING:**
Do not use`heroku authorizations:rotate <ID>` or its underlying [API](https://devcenter.heroku.com/articles/platform-api-reference#oauth-authorization-regenerate)
as the new authorization access token's time to live is set to 28880 seconds (~8 hours) regardless of the original TTL.
This out-of-band time to live change will not be reflected in an existing resource configuration and will lead to configuration drift.

### Expired authorizations
Heroku automatically deletes an expired authorization completely from their systems, making it unavailable during state refresh.
In this scenario, the resource will remove itself from state and be created again on the next `terraform apply`.

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
  Setting this attribute allows an OAuth authorization to be created in a user account that's different from the account
  used to authenticate with the provider. To define the equivalent environment variable name, replace the `%s` in
  `HEROKUX_%s_API_KEY` with the attribute's value. For example if the attribute value is `myBotUser_X`,
  the environment variable should be `HEROKUX_MYBOTUSER_X_API_KEY` in all upper case. Please also note the following:
    * If this attribute is not set, the resource **will create** the OAuth authorization in the same account
      used to authenticate with the provider.
    * The resource will surface an error indicating the missing environment variable if this attribute is set and the
      equivalent variable is not set in the environment.
    * A value may only include words, letters, or underscore with a max length of 32 characters. Case-insensitive.
    * Each `herokux_oauth_authorization` resource can define a unique value for this attribute. However, this translates
      to an equal number of equivalent environment variables.
* `time_to_live` - `<integer>` Set expiration in seconds. No expiration if attribute is not set in the configuration.
* `description` - `<string>` Set a custom authorization description.

-> **IMPORTANT!**
Modifying any of the attributes above sans `description` will result in a resource recreation.

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
