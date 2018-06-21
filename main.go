package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/plugin"
	"github.com/rackn/terraform-provider-drp/drp"
)

func main() {
	args := os.Args

	if len(args) > 1 && args[1] == "define-json" {
		p := drp.Provider().(*schema.Provider)
		b, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			fmt.Printf("GREG: marshal err := %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(b))
	} else {
		plugin.Serve(&plugin.ServeOpts{ProviderFunc: drp.Provider})
	}
}
