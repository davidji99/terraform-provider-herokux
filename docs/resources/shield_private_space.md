---
layout: "herokux"
page_title: "HerokuX: herokux_shield_private_space"
sidebar_current: "docs-herokux-resource-shield-private-space"
description: |-
  Provides a resource to manage a Shield Private Space.
---

# herokux\_shield\_private\_space

This resource manages a [Shield Private Space](https://devcenter.heroku.com/articles/shield-private-space) in Heroku.

Please note the following differences between `herokux_shield_private_space` & [`heroku_space`](https://registry.terraform.io/providers/heroku/heroku/latest/docs/resources/space):

* `herokux_shield_private_space` cannot create non-shield Private Spaces.
* `herokux_shield_private_space` cannot set ingress/inbound IP ranges. Please use [`heroku_space_inbound_ruleset`](https://registry.terraform.io/providers/heroku/heroku/latest/docs/resources/space_inbound_ruleset).
This resource does not replicate `heroku_space.trusted_ip_ranges` as it is deprecated.
* `herokux_shield_private_space` can set a [log drain URL](https://devcenter.heroku.com/articles/private-space-logging#enable-private-space-logging)
on initial resource creation.

### Regarding Log Drain
Although the Heroku UI allows users to create a Shield Private Space without defining a `log_drain`, this resource
enforces the `log_drain` as a required attribute. There are a few reasons for this:

1. You cannot turn on Private Space Logging after a space has been created.
You can, however, change the `log_drain` at a later point if the space was created with Private Space Logging enabled.
This particular design makes it difficult to replicate a log drain enabled shield Private Space as a terraform resource.

1. It is not possible to set `log_drain` to an empty string on resource creation or modification.

1. If you do not supply a log drain when creating a Shield Space then that Shield Private Space will not have
Private Space Logging enabled and will not benefit from the compliance and data-residency related enhancements
associated with this feature.

### Resource Timeouts
During creation, this resource checks the status of the shield private space provisioning status.
The aforementioned timeout can be customized via the `timeouts.shield_private_space_create_verify_timeout`
attribute in your `provider` block.

For example:

```hcl-terraform
provider "herokux" {
  timeouts {
    shield_private_space_create_verify_timeout = 30
  }
}
```

## Example Usage

```hcl-terraform
resource "herokux_shield_private_space" "foobar" {
  name = "my secret shield space"
  team_id = "0f17fba1-269a-45fb-b62d-ebcc6e749987"
  region = "virginia"
  log_drain = "https://somename:somesecret@loghost.example.com/logpath"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) `<string>` Name of shield private space to create. Minimum is 3 characters.
  Name may only contain numbers (0-9), lowercase letters (a-z) and non-consecutive hyphens ('-').
  It must begin and end with either a number or letter.
* `region` - (Required) `<string>` [Heroku region name](https://devcenter.heroku.com/articles/regions#viewing-available-regions).
  Valid options: `dublin`, `frankfurt`, `oregon`, `sydney`, `tokyo`, `virginia`.
* `log_drain` - (Required) `<string>` Direct log drain url. Must be a HTTPS url.
* `team_id` - (Required) `<string>` The UUID of the Heroku Team which will own the Shield Private Space.
* `cidr` - (Optional) `<string>` The RFC-1918 CIDR the Private Space will use.
  It must be a /16 in `10.0.0.0/8`, `172.16.0.0/12` or `192.168.0.0/16`.
* `data_cidr` - (Optional) `<string>` The RFC-1918 CIDR that the Private Space will use for the Heroku-managed peering connection
that’s automatically created when using Heroku Data add-ons. It must be between a `/16` and a `/20`.

## Attributes Reference

The following attributes are exported:

* `outbound_ips` - The space's stable outbound [NAT IPs](https://devcenter.heroku.com/articles/platform-api-reference#space-network-address-translation).
* `is_shield` - This is just a "make-sure" attribute that the created space is actually a Shield private space.
* `team_name` - The name of the Heroku Team (aka organization) that owns the Private Space.

## Import

An existing shield private space can be imported using its UUID. It is not possible to use this resource to import
a non-shield private space.

For example:

```shell script
$ terraform import herokux_shield_private_space.foobar "19ef8940-6ad2-4ea8-825e-cee883f679e2"
```
