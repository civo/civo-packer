package main

import (
	"github.com/civo/civo-packer/builder/civo"
	"github.com/hashicorp/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	if err := server.RegisterBuilder(new(civo.Builder)); err != nil {
		panic(err)
	}
	server.Serve()
}
