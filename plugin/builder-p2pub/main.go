package main

import (
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/sishihara/packer-builder-p2pub/builder/p2pub"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(p2pub.Builder))
	server.Serve()
}
