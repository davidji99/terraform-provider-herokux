---
layout: "herokux"
page_title: "HerokuX: herokux_app_webhook"
sidebar_current: "docs-herokux-resource-app-webhook"
description: |-
  Provides a resource to manage an App webhook.
---

# herokux\_app\_webhook

This resource manages an app webhook. [App webhooks](https://devcenter.heroku.com/articles/app-webhooks) enable you to
receive notifications whenever particular changes are made to your Heroku app. You can subscribe to notifications
for a wide variety of events.

## Example Usage

```hcl-terraform
resource "herokux_app_webhook" "foobar" {
	app_id = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	level = "notify"
	url = "https://example.com/hooks"
	event_types = ["api:addon-attachment"]
	name = "my-custom-webhook-name"
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) `<string>` The UUID of the app to subscribe to.

* `level` - (Required) `<string>` The type of webhook [push mechanism](https://devcenter.heroku.com/articles/app-webhooks#retries-and-limits.
There are two possible options:

    * `notify` - Does not retry unsuccessful webhook notifications.

    * `sync` - Retries failed requests until they succeed or until a determined limit is reached.

* `url` - (Required) `<string>` The URL of your server endpoint that will receive all webhook notifications.

* `event_types` - (Required) `<list(string)>` [List of the entities](https://devcenter.heroku.com/articles/app-webhooks#step-2-determine-which-events-to-subscribe-to)
you want to subscribe to notifications for. Valid options are:

    * `api:addon-attachment`, `api:addon`, `api:app`, `api:build`, `api:collaborator`, `api:domain`,
    `api:dyno`, `api:formation`, `api:release`, `api:sni-endpoint`, `api:ssl-endpoint`.

* `name` - (Optional) `<string>` Name of the webhook. Name must only include lowercase letters, numbers, and dashes.
If not set, Heroku will randomly generate a webhook name.

* `secret` - (Optional) `<string>` A value that Heroku will use to sign all webhook notification requests
(the signature is included in the request’s `Heroku-Webhook-Hmac-SHA256` header). If you omit this attribute,
this resource will generate a secret which will never be show again. This attribute is marked as `Sensitive`
and will not appear in any console output.

## Attributes Reference

The following attributes are exported:

* `signing_secret` - A randomly generated string that is set if the `secret` attribute is omitted from your configuration.
This attribute is marked as `Sensitive` and will not appear in any console output.

* `app_name` - The name of the app the webhook is subscribed to.

## Import

An existing app webhook can be imported using a composite of the app UUID and webhook UUID separated by a colon.
For convenience, this resource will accept both the app & webhook name to make up the composite import ID.
However, the resource only stores the UUIDs in state.

For example:

```shell script
$ terraform import herokux_app_webhook.foobar "6fae1ee0-c034-4775-a798-890bc64f98eb:38b6e411-1361-4adf-ace6-734c9a95513d"
```

-> **IMPORTANT!**
Due to API limitations, it is not possible to retrieve a webhook's manually defined `secret` or the auto-generated
`signing_secret`. Therefore, this resource recommends users only import webhooks that have a `secret` that can be
manually updated post import.