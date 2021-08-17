---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_dataclip_team_association"
sidebar_current: "docs-herokux-resource-postgres-dataclip-team-association"
description: |-
Provides a resource to manage the association of a Heroku team and Heroku Postgres dataclip.
---

# herokux_postgres_dataclip_team_association

This resource manages the [association](https://devcenter.heroku.com/articles/dataclips#sharing-with-individuals-and-teams)
of a Heroku team and Heroku Postgres dataclip.

-> **IMPORTANT!**
All members of the selected team can view the dataclip.

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

data "heroku_team" "super-team" {
  name = "super team"
}

resource "herokux_postgres_dataclip_team_association" "team" {
  dataclip_id = herokux_postgres_dataclip.dataclip.id
  dataclip_slug = herokux_postgres_dataclip.dataclip.slug
  team_id = data.heroku_team.super-team.id
}
```

## Argument Reference

The following arguments are supported:

* `dataclip_id` - (Required) `<string>` The UUID of the dataclip. Ideal to source this value from `herokux_postgres_dataclip.id`.
* `dataclip_slug` - (Required) `<string>` The slug of the dataclip. Ideal to source this value from `herokux_postgres_dataclip.slug`.
* `team_id` - (Required) `<string>` The UUID of a Heroku team.

## Attributes Reference

The following attributes are exported:

* `team_name` - Name of team.

## Import

An existing Postgres dataclip team association can be imported using a composite of the dataclip slug
and team name separated by a colon character (`:`).

For example:

If `super team` has shared access to an existing dataclip, whose browser URL is
`https://data.heroku.com/dataclips/lfcdwnpbqthzyeyiucvgtgnuevhi`, the import ID is as follows:

```shell script
$ terraform import herokux_postgres_dataclip_team_association.team "lfcdwnpbqthzyeyiucvgtgnuevhi:super team"
```