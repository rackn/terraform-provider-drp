package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/rackn/terraform-provider-drp/drp"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: drp.Provider})
}
