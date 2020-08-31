package main

import (
	"github.com/davidji99/terraform-provider-herokux/herokux"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: herokux.New})
}
