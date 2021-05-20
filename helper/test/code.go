package test

import "fmt"

func HerokuProviderBlock() string {
	return `
terraform {
  required_providers {
    heroku = {
      source = "heroku/heroku"
      version = ">= 4.0"
    }
  }
}

provider "heroku" {}
`
}

func HerokuAppAddonBlock(appName, orgName, addonPlan string) string {
	return fmt.Sprintf(`
%s

resource "heroku_app" "foobar" {
  name   = "%s"
  region = "us"

  organization {
    name = "%s"
  }
}

resource "heroku_addon" "foobar" {
  app  = heroku_app.foobar.name
  plan = "%s"
}
`, HerokuProviderBlock(), appName, orgName, addonPlan)
}
