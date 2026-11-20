---
layout: "herokux"
page_title: "HerokuX: herokux_connect_mappings"
sidebar_current: "docs-herokux-resource-connect-mappings"
description: |-
    Provides a resource to manage a Heroku Connect's mappings.
---

# herokux\_connect\_mappings

This resource manages the [mappings](https://devcenter.heroku.com/articles/heroku-connect#mapping-objects)
for a [Heroku Connect](https://devcenter.heroku.com/articles/heroku-connect) instance/addon.

This resource manages multiple mappings at once with the reasoning for this peculiar design explained in the
[Recommended Workflow](#recommended-workflow) section below. This is why this resource name is plural,
not singular.

### Prerequisites

Due to Connect API limitations, the following MUST be completed either manually or via Terraform prior to using this resource:

1. [Create a Heroku app and Heroku PostgreSQL database](https://devcenter.heroku.com/articles/heroku-connect-api#step-1-create-a-heroku-app-and-heroku-postgresql-database).
1. [Create the Heroku Connect add-on](https://devcenter.heroku.com/articles/heroku-connect-api#step-2-create-the-heroku-connect-add-on)
1. [Link the new add-on to your Heroku user account](https://devcenter.heroku.com/articles/heroku-connect-api#step-3-link-the-new-add-on-to-your-heroku-user-account)
1. [Configure the database key and schema for the connection](https://devcenter.heroku.com/articles/heroku-connect-api#step-4-configure-the-database-key-and-schema-for-the-connection)
1. [Authenticate the connection to your Salesforce Org](https://devcenter.heroku.com/articles/heroku-connect-api#step-5-authenticate-the-connection-to-your-salesforce-org)

Steps #1 and #2 can be achieved using the Heroku provider's `heroku_app` & `heroku_addon` resource.
Step #3 is done automatically as part of this resource's lifecycle but can be done externally should
users run into any authentication issues.

### Recommended Workflow
Mirroring [Heroku's documentation](https://devcenter.heroku.com/articles/heroku-connect-api#step-6-import-a-mapping-configuration),
the easiest way to have this resource manage mappings is to first export them from an existing Connect instance.
Then, look at the [Example Usage](#example-usage) section below for a few ways to reference the exported JSON file data.

Once you export existing mapping(s), please delete the `connection` field from the exported JSON prior to using the file
for this resource's `mappings` attribute. For example, please remove the following:

```json
{
  "connection": {
    "app_name": "SOME_APP_NAME",
    "organization_id": "SOME_ORG_ID",
    "exported_at": "2020-12-11T14:40:54.084290+00:00",
    "features": {
      "disable_bulk_writes": false,
      "poll_db_no_merge": true,
      "poll_external_ids": false,
      "rest_count_only": false
    },
    "api_version": "50.0",
    "name": "this is the name",
    "logplex_log_enabled": false
  }
}
```

Please note that while the Connect API has an [endpoint](https://devcenter.heroku.com/articles/heroku-connect-api#create-a-new-mapping)
to create a single mapping, this provider's author designed the workflow to fit Heroku's documented 'easiest' workflow
to manage mappings as described in this section's first paragraph.

### Resource Behavior
Apply manages the Salesforce objects named in the configured JSON. Terraform records each object's mapping ID. Removing an object from HCL or destroying the resource deletes only a matching recorded ID. Other mappings on the connection remain available. If someone recreates a managed mapping under the same name, verify the new mapping before adopting it through import and a configured apply.

After import, `ownership_recorded` is false and the `mappings` state value contains `{"mappings":[]}`. Review the object names in your configured JSON and run a normal apply to establish ownership. With `lifecycle { ignore_changes = [mappings] }`, the resource stays unowned. Remove the ignore rule and apply when you want Terraform to manage the declared objects.

Upgrading from the old resource schema, including provider v1.8.0 and v1.9.0, requires this same one-time apply even if state previously showed `ownership_recorded = true`. An earlier adoption could have recorded objects from the entire connection. Previously recorded IDs serve only to detect a same-name replacement during reconfirmation. Destroying the resource before reconfirmation preserves every remote mapping, including when using `-refresh=false`.

If a bulk import is interrupted after submission, `pending_mapping_ids` records the objects whose ownership still needs verification. This applies to new mappings added to an existing resource as well as the first import. Previously confirmed IDs remain in `mapping_data`. An empty ID means Terraform never verified that mapping. Destroy retains the state entry while ownership is unconfirmed. To recover, inspect the mappings and IDs in Heroku, back up Terraform state, remove this resource's state entry with `terraform state rm`, and use the existing composite import below. Apply the verified list of objects before destroying them.

### Resource Delay
This resource will wait a given amount of time after creating or updating mappings. This delay is to address a likely
race condition when retrieving mappings soon after creating them in Heroku. Heroku documentation does not provide
a recommended or estimated delay period, so this provider provides the means for users to configure this wait period.
It is likely the number of mappings will affect the length of this delay.

The aforementioned delay can be customized via the `delays.connect_mapping_modify_delay` attribute in your `provider` block.
The delay value defaults to 15 seconds with a minimum requirement of 5 seconds.

For example:

```hcl-terraform
provider "herokux" {
  delays {
    connect_mapping_modify_delay = 60
  }
}
```

## Example Usage

Using shell-style "here doc" syntax:

```hcl-terraform
resource "heroku_app" "foobar" {
  name   = "my_foobar_app"
  region = "us"

  organization {
    name = "my_org"
  }
}

resource "heroku_addon" "connect" {
  app_id  = heroku_app.foobar.id
  plan = "heroku-shield:free"
}

resource "herokux_connect_mappings" "foobar" {
  app_id = heroku_app.foobar.uuid
  connect_id = heroku_addon.connect.id
  mappings = <<-EOF
{
    "mappings": [
        {
            "object_name": "AcceptedEventRelation",
            "config": {
                "access": "read_only",
                "sf_notify_enabled": false,
                "sf_polling_seconds": 600,
                "sf_max_daily_api_calls": 30000,
                "fields": {
                    "CreatedDate": {},
                    "Id": {},
                    "IsDeleted": {},
                    "SystemModstamp": {}
                },
                "indexes": {
                    "Id": {
                        "unique": true
                    },
                    "SystemModstamp": {
                        "unique": false
                    }
                }
            }
        }
    ],
    "version": 1
}
EOF
}
```

When reading mappings from a file, use the configured object names as `for_each` keys for downstream resources. Terraform knows those names at plan time. A lookup such as `mapping_data["Account"]` resolves to the verified Heroku ID during the same adoption apply.

```hcl-terraform
resource "heroku_app" "foobar" {
  name   = "my_foobar_app"
  region = "us"

  organization {
    name = "my_org"
  }
}

resource "heroku_addon" "connect" {
  app_id  = heroku_app.foobar.id
  plan = "heroku-shield:free"
}

locals {
  mapping_json = file("${path.module}/mappings.json")
  object_names = toset([for mapping in jsondecode(local.mapping_json).mappings : mapping.object_name])
}

resource "herokux_connect_mappings" "foobar" {
  app_id     = heroku_app.foobar.uuid
  connect_id = heroku_addon.connect.id
  mappings   = local.mapping_json
}

resource "terraform_data" "mapping" {
  for_each = local.object_names
  input    = herokux_connect_mappings.foobar.mapping_data[each.key]
}
```

Using Terraform's [`local`](https://registry.terraform.io/providers/hashicorp/local/latest/docs/data-sources/file)
provider:

```hcl-terraform
data "local_file" "mapping_json_file" {
  filename = "${path.module}/foo.bar"
}

resource "heroku_app" "foobar" {
  name   = "my_foobar_app"
  region = "us"

  organization {
    name = "my_org"
  }
}

resource "heroku_addon" "connect" {
  app_id  = heroku_app.foobar.id
  plan = "heroku-shield:free"
}

resource "herokux_connect_mappings" "foobar" {
  app_id = heroku_app.foobar.uuid
  connect_id = heroku_addon.connect.id
  mappings = data.local_file.mapping_json_file.content
}
```

### Resource Dependency
It is recommended to make use of Terraform's [`depends_on`](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html)
meta-argument to establish a resource dependency on a `heroku_addon` resource representing the Heroku postgres instance.
This resource does not provide a `postgres_id` attribute as this ID is not used by the underlying APIs.

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) `<string>` The UUID of the app. It is recommended to reference the UUID via
`heroku_app.foobar.uuid` if the app was created via Terraform. This allows the configuration to create
resource dependencies.
* `connect_id` - (Required) `<string>` The UUID of the Heroku Connect instance/addon.
* `mappings` - (Required) `<string>` Properly formatted JSON string representing Connect mappings
between Heroku and Salesforce.

## Attributes Reference

The following attributes are exported:

* `ownership_recorded` - True after a configured apply has established ownership of the declared mappings.
* `mapping_ids` - List of all Connect mapping IDs currently managed by this resource.
* `mapping_object_names` - List of all Connect mapping object names currently managed by this resource.
* `mapping_data` - Map of the mappings where the key is the mapping object name, and the value is the map UUID.

## Import

An existing Connect's mapping can be imported using a composite value of the app UUID and Connect UUID
separated by a colon.

For example:

```shell script
$ terraform import herokux_connect_mappings.foobar "33d4631b-2c77-4b99-b657-752ad8f68322:7f1f2784-2c35-4efa-b0cd-544c9784fe9b"
```
