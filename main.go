package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/trotttrotttrott/terraform-provider-gmailfilter/gmailfilter"
)

var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/hashicorp/gmailfilter",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), gmailfilter.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
