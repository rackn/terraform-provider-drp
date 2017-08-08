package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/rackn/terraform-provider-drp/provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: provider.Provider})
}
