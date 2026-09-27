package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/zone-eu/terraform-provider-zone/internal/provider"
)

var (
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address:         "registry.terraform.io/kepsic/zoneeu",
		Debug:           debug,
		ProtocolVersion: 5,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

