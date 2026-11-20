---
layout: "herokux"
page_title: "HerokuX: herokux_team_oauth_authorization"
description: |-
  Manages a direct OAuth authorization owned by a Heroku team.
---

# herokux_team_oauth_authorization

Create, import, rename, or revoke a direct authorization for a Heroku team. Configure the provider with a team administrator's credentials. [Heroku team authorizations](https://devcenter.heroku.com/articles/oauth#team-authorization)

## Example usage

```hcl
resource "herokux_team_oauth_authorization" "reports" {
  team        = "example"
  scope       = ["read"]
  description = "CI reports"
}
```

## Arguments

* `team` (required) is the team name or UUID.
* `scope` (required) is a nonempty set of OAuth scopes. Supported names are `global`, `identity`, `read`, `write`, `read-protected`, and `write-protected`. Heroku limits team tokens to team resources.
* `description` (optional) can be updated in place.
* `time_to_live` (optional) requests a lifetime at creation, in seconds from `1` to `31536000`. Heroku applies its team default of one year when omitted.

Changing the team, scope, or requested lifetime replaces the authorization. Destroy revokes it immediately and integrations using the token lose access.

## Attributes

* `id` is the authorization UUID.
* `token_id` identifies the access token when Heroku supplies it.
* `expires_in` is the last reported remaining lifetime in seconds.
* `access_token` contains the bearer token when Terraform creates the authorization. It is marked sensitive and stored in Terraform state. Restrict access to the state backend. Refresh preserves the saved value when Heroku omits the secret and clears a stale value if Heroku reports a different token ID.

## Import

Use `heroku authorizations --team example` to find the authorization UUID. Import with the team name or UUID followed by a slash and the **authorization UUID**.

```shell
terraform import herokux_team_oauth_authorization.reports 'example/ac19340e-e020-417c-8ab6-4646796eb156'
```

Import reads the team's authorization metadata. The existing credential keeps working and `access_token` is `null` in Terraform state. Consumers that need the token must use their previously stored copy.

For an empty plan after import, match the team and scopes, omit or match the description, and omit `time_to_live`. Heroku reports the remaining lifetime. Setting a new requested lifetime after import plans a replacement.
