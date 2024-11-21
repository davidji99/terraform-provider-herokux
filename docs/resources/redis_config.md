---
layout: "herokux"
page_title: "HerokuX: herokux_redis_config"
sidebar_current: "docs-herokux-resource-redis-config"
description: |-
  Provides a resource to manage the configurations for a Redis instance
---

# herokux\_redis\_config

This resource manages the [configurations](https://devcenter.heroku.com/articles/heroku-redis#configuring-your-instance)
for a Heroku Redis instance.

-> **IMPORTANT!**
Due to API limitations, deleting this resource will result only in state removal. All user defined configuration values
will remain the same until they are changed by `terraform` or `heroku redis:***` commands.

## Example Usage

```hcl-terraform
resource "heroku_app" "foobar" {
  name   = "my_foobar_app"
  region = "us"

  organization {
    name = "my_org"
  }
}

resource "heroku_addon" "redis" {
  app_id  = heroku_app.foobar.id
  plan = "heroku-redis:premium-0"
}

resource "herokux_redis_config" "foobar" {
  redis_id = heroku_addon.redis.id
  maxmemory_policy = "allkeys-lfu"
  notify_keyspace_events = "K"
  timeout = 500
}
```

## Argument Reference

The following arguments are supported:

* `redis_id` - (Required) `<string>` The UUID or name of a Heroku Redis instance.
* `maxmemory_policy` - (Optional) `<string>` Set the key eviction policy used when an instance reaches its storage limit.
  Heroku, by default, sets this to `noeviction`. Valid options are as follows:
    * `noeviction` will return errors when the memory limit is reached.
    * `allkeys-lru` will remove less recently used keys first.
    * `volatile-lru` will remove less recently used keys first that have an expiry set.
    * `allkeys-random` will evict random keys.
    * `volatile-random` will evict random keys but only those that have an expiry set.
    * `volatile-ttl` will only evict keys with an expiry set and a short TTL.
    * `volatile-lfu` will evict using approximated LFU among the keys with an expire set.
    * `allkeys-lfu` will evict any key using approximated LFU.
* `notify_keyspace_events` - (Optional) `<string>` Enable specific [keyspace notifications](https://redis.io/topics/notifications)
configuration. Heroku, by default, sets this to `disabled`. Valid options are as follows:
    * `K` Keyspace events, published with `__keyspace@<db>__` prefix.
    * `E` Keyevent events, published with `__keyevent@<db>__` prefix.
    * `g` Generic commands (non-type specific) like DEL, EXPIRE, RENAME, ....
    * `$` String commands.
    * `l` List commands.
    * `s` Set commands.
    * `h` Hash commands.
    * `z` Sorted set commands.
    * `t` Stream commands.
    * `x` Expired events (events generated every time a key expires).
    * `e` Evicted events (events generated when a key is evicted for maxmemory).
    * `m` Key miss events (events generated when a key that doesn't exist is accessed).
    * `A` Alias for `g$lshztxe`, so that the "AKE" string means all the events except "m".
    * `disabled` Disable keyspace notifications
* `timeout` - (Optional) `<integer>` Number of seconds Redis waits before killing idle connections.
A value of zero means that connections will not be closed. Heroku, by default, sets the value to 300 seconds (5 minutes).
Minimum required value is `0`.

## Attributes Reference

N/A

## Import

Existing redis configurations can be imported using the Heroku Redis instance UUID.

For example:

```shell script
$ terraform import herokux_redis_config.foobar "57d660e0-3d20-40b7-8d20-e77b95189e5a"
```
