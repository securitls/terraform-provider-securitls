package main

import (
	"context"
	"terraform-provider-securitls/securitls"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version = "dev"
var commit = "none"

func main() {
	providerserver.Serve(context.Background(), securitls.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/securitls/securitls",
	})
}
