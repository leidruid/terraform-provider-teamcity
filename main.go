package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"

	"github.com/leidruid/terraform-provider-teamcity/teamcity"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: teamcity.Provider})
}
