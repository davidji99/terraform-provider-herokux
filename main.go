package main

import (
	"github.com/davidji99/terraform-provider-herokuplus/herokuplus"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: herokuplus.New})
}
