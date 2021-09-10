package main

import (
	"github.com/netrisai/terraform-provider-netris/netris"

	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: netris.Provider,
	})
}
