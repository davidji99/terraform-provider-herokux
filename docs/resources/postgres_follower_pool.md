---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_follower_pool"
description: |-
  Manages a follower instance pool on a Heroku Postgres Advanced database.
---

# herokux_postgres_follower_pool

Add read capacity to an existing Heroku Postgres Advanced database. Heroku bills follower compute separately. This resource owns the follower pool and waits for provisioning, resizing, and deletion to finish. [Heroku instance pool guide](https://devcenter.heroku.com/articles/postgres-advanced-instance-pools)

## Example usage

```hcl
resource "herokux_postgres_follower_pool" "analytics" {
  postgres_id    = var.postgres_addon_uuid
  name           = "analytics"
  level          = "4G-Performance"
  instance_count = 1
}
```

## Arguments

* `postgres_id` (required) is the Advanced Postgres add-on UUID.
* `name` (required) is a unique pool name. Heroku reserves `leader`. Changing the name or database replaces the follower pool.
* `level` (required) is the Heroku compute level per instance. Changing it resizes the pool in place.
* `instance_count` (optional) is the number of follower instances. It defaults to `1`. Increasing the count adds billable instances.

Heroku applies level and count changes separately. Terraform refreshes the pool between changes so a subsequent apply can resume a partially completed resize. Destroy removes the selected follower pool and its read capacity.

## Import

Use the database UUID followed by the follower pool ID from the database info API response.

```shell
terraform import herokux_postgres_follower_pool.analytics '29138cab-acf1-40c6-ac83-56697c8f0221:POOL_ID'
```
