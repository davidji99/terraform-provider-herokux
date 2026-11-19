---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_publication"
description: |-
  Manages a logical replication publication on Heroku Postgres Advanced.
---

# herokux_postgres_publication

Manage which tables a Heroku Postgres Advanced database publishes through logical replication. Enable publishing on the source database before applying. [Heroku logical replication guide](https://devcenter.heroku.com/articles/logical-replication-on-heroku-postgres-advanced)

## Example usage

```hcl
resource "herokux_postgres_publication" "orders" {
  postgres_id = var.postgres_addon_uuid
  name        = "orders"
  tables      = ["public.orders", "public.order_items"]
}
```

## Arguments

* `postgres_id` (required) is the Advanced Postgres add-on UUID.
* `name` (required) is the publication name. Changing the database or name replaces the publication.
* Set exactly one of `tables` or `schemas`. Both accept a nonempty set of names. Use fully qualified names for tables. A named schema includes new tables created in that schema.

Target changes update the publication in place. Destroy deletes the publication and stops subscribers from receiving its changes.

## Import

```shell
terraform import herokux_postgres_publication.orders '29138cab-acf1-40c6-ac83-56697c8f0221:orders'
```

Match the declared tables or schemas returned by Heroku. For a publication originally created with the CLI’s `--all-schemas` snapshot, Heroku returns the resolved target.
