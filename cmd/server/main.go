package main

import (
	"log"

	"github.com/svaan1/go-tcc/internal/server"
)

func main() {
	server := server.New()

	err := server.Listen()
	if err != nil {
		log.Fatalf("Failed to start server %v", err)
	}

	log.Printf("Server started a %s connection at %s", server.Config.Network, server.Config.Address)

	select {}
}
