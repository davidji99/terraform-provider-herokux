---
layout: "herokux"
page_title: "HerokuX: herokux_postgres_storage_quota"
description: |-
  Manages the storage quota policy on a Heroku Postgres Advanced database.
---

# herokux_postgres_storage_quota

Manage one storage policy per Advanced Postgres add-on. This resource owns both thresholds and the enforcement action. **The first apply clears any existing threshold omitted from configuration.** Import the quota before applying to preview those changes. [Heroku quota behavior](https://devcenter.heroku.com/articles/postgres-advanced-quotas)

## Example usage

```hcl
resource "herokux_postgres_storage_quota" "orders" {
  postgres_id        = var.postgres_addon_uuid
  warning_gb         = 80
  critical_gb        = 95
  enforcement_action = "notify"
}
```

## Arguments

* `postgres_id` (required) is the Advanced Postgres add-on UUID and resource ID.
* `warning_gb` and `critical_gb` (optional) are positive whole gigabytes. Omit a threshold to clear it.
* `enforcement_action` (required) accepts `notify`, `restrict`, or `none`. The `restrict` action can limit writes and connections after Heroku’s grace period.

Destroy removes the resource from Terraform state and preserves the active policy on Heroku. To lift a threshold, remove its argument and apply while Terraform still manages the resource.

## Import

```shell
terraform import herokux_postgres_storage_quota.orders '29138cab-acf1-40c6-ac83-56697c8f0221'
```
