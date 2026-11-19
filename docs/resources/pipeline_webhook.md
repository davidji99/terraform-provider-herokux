---
layout: "herokux"
page_title: "HerokuX: herokux_pipeline_webhook"
sidebar_current: "docs-herokux-resource-pipeline-webhook"
description: |-
  Manages a webhook subscription on a Heroku pipeline.
---

# herokux_pipeline_webhook

Manages one webhook subscription on a Heroku pipeline. Heroku sends matching events to the configured receiver URL.

## Example usage

```hcl
resource "heroku_pipeline" "deploy" {
  name = "deploy"
}

resource "herokux_pipeline_webhook" "builds" {
  pipeline_id = heroku_pipeline.deploy.id
  url         = "https://hooks.example.com/heroku"
  level       = "notify"
  event_types = ["api:build"]
}
```

## Arguments

* `pipeline_id` (required) is the pipeline UUID. Changing it replaces the subscription.
* `url` (required) is the HTTP or HTTPS URL that receives webhook deliveries.
* `level` (required) controls retries. Use `notify` for one delivery attempt or `sync` for retried delivery.
* `event_types` (required) is a nonempty set of event names. Heroku validates which events the pipeline can subscribe to.
* `secret` (optional, sensitive) is the secret Heroku uses to sign deliveries. Changing it rotates the secret in place. Removing a previously configured secret replaces the subscription so Heroku can generate a new one.
* `authorization` (optional, sensitive) is the complete Authorization header sent to the receiver. Removing a previously configured value clears the remote header.

## Attributes

* `id` combines the pipeline UUID and webhook UUID with a colon.
* `pipeline_name` is the current name of the pipeline.
* `signing_secret` (sensitive) contains the Heroku-generated signing secret when the provider creates a subscription with `secret` omitted. Heroku returns this value once. Terraform stores signing secrets and configured authorization headers in state.

## Import

Use the pipeline name or UUID followed by a colon and the webhook UUID.

```shell
terraform import herokux_pipeline_webhook.builds 'deploy:38b6e411-1361-4adf-ace6-734c9a95513d'
```

Refresh resolves the pipeline name to its UUID. Heroku omits signing secrets and authorization headers from read responses. After import those attributes are empty in state, and edits to the URL, level, or event types preserve the existing credentials. To rotate either credential, set its new value in configuration.
