[![tests](https://github.com/davidji99/terraform-provider-herokux/actions/workflows/tests.yml/badge.svg)](https://github.com/davidji99/terraform-provider-herokux/actions/workflows/tests.yml)
<a href="https://goreportcard.com/report/github.com/davidji99/terraform-provider-herokux"><img class="badge" tag="github.com/davidji99/terraform-provider-herokux" src="https://goreportcard.com/badge/github.com/davidji99/terraform-provider-herokux"></a>

# Terraform Provider HerokuX

The HerokuX provider interacts with Heroku's undocumented APIs and Platform API variants to provide additional resources
not available in the official [Heroku Terraform provider](https://github.com/heroku/terraform-provider-heroku).

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) `v0.13+`
- [Go](https://golang.org/doc/install) `v1.16+` (to build the provider plugin)

## Usage

```hcl
provider "herokux" {
  version = "~> 0.1.0"
}
```

This provider is not compatible with terraform `v0.11.x`.

## Contributing

When contributing new resources, please make sure the new resource adheres to one of the following criteria:

* Uses an undocumented API.
* Uses a Platform API variant.
* Significantly alters the design and logic of an existing `heroku` provider resource.

Regardless of the aforementioned guideline, please feel free to submit contributions. The provider's maintainer(s)
will initiate a discussion regarding resource placement if deemed necessary.

Please also view the `CONTRIBUTING.md` file for the general contribution policy.

## Development

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.14+ is *required*).

### Build the Provider

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```shell script
$ make build
...
$ $GOPATH/bin/terraform-provider-herokux
...
```

### Using the Provider

To use the dev provider with local Terraform, copy the freshly built plugin into Terraform's local plugins directory:

```sh
cp $GOPATH/bin/terraform-provider-herokux-dev ~/.terraform.d/plugins/
```

Set the HerokuX provider without a version constraint:

```hcl
provider "herokux" {}
```

Then, initialize Terraform:

```shell script
terraform init
```

### Testing

Please see the [TESTING](TESTING.md) guide for detailed instructions on running tests.

### Updating or adding dependencies

This project uses [Go Modules](https://github.com/golang/go/wiki/Modules) for dependency management.

This example will fetch a module at the release tag and record it in your project's `go.mod` and `go.sum` files.

If a module does not have release tags, then `module@SHA` can be used instead.
