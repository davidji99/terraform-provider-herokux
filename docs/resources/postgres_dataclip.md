---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_dataclip"
sidebar_current: "docs-herokux-resource-postgres-dataclip"
description: |-
Provides a resource to manage a dataclip on a Heroku Postgres database.
---

# herokux\_postgres\_dataclip

This resource manages a [Dataclip](https://devcenter.heroku.com/articles/dataclips) on a Heroku Postgres database.
Heroku Dataclips enable you to create SQL queries for your Heroku Postgres databases
and share the results with colleagues, third-party tools, and the public.
Recipients of a dataclip can view the data in their browser and also download it in JSON and CSV formats.

Please be mindful of a Dataclip's [limitations and restrictions](https://devcenter.heroku.com/articles/dataclips#limits-and-restrictions).

-> **IMPORTANT!**
Dataclips cannot connect to Shield databases.

## Example Usage

```hcl-terraform
resource "heroku_app" "default" {
  name   = "my_default_app"
  region = "us"

  organization {
    name = "my_org"
  }
}

resource "heroku_app" "secondary" {
  name   = "my_secondary_app"
  region = "us"

  organization {
    name = "my_org"
  }
}

resource "heroku_addon" "primary-db" {
  app  = heroku_app.default.name
  plan = "heroku-postgresql:premium-0"
}

resource "heroku_addon_attachment" "database" {
  app_id  = heroku_app.secondary.id
  addon_id = heroku_addon.primary-db.id
}

resource "herokux_postgres_dataclip" "primary-db-users" {
  postgres_attachment_id = heroku_addon_attachment.database.id
  title = "list of all primary db users"
  sql = "select * from users"
  enable_shareable_links = true
}
```

## Argument Reference

The following arguments are supported:

* `postgres_attachment_id` - (Required) `<string>` The UUID of the addon attachment.
* `title` - (Required) `<string>` Title of the dataclip.
* `sql` - (Required) `<string>` SQL query.
* `enable_shareable_links` - `<boolean>` Enable shareable links to share the results of this dataclip publicly.
Defaults to `false`.

## Attributes Reference

The following attributes are exported:

* `slug` - Slug value of the dataclip.
* `creator_email` - Email address of the dataclip's creator.
* `attachment_name` - Addon attachment name that the dataclip is using. This usually is `DATABASE`.
* `addon_id` - The UUID of the Postgres database used by the dataclip.
* `addon_name` - The name of the Postgres database used by the dataclip.
* `app_id` - The UUID of the app that owns the Postgres addon.
* `app_name` - The name of the app that own the Postgres addon.

## Import

An existing Postgres dataclip can be imported using the dataclip slug value. The slug value can be found via
the browser URL when viewing a single dataclip

For example:

If the existing dataclip browser URL is `https://data.heroku.com/dataclips/lfcdwnpbqthzyeyiucvgtgnuevhi`,
the slug is `lfcdwnpbqthzyeyiucvgtgnuevhi`.

```shell script
$ terraform import herokux_postgres_dataclip.primary-db-users "lfcdwnpbqthzyeyiucvgtgnuevhi"
```