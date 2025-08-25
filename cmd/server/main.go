package main

import (
	"log"
	"net"

	"github.com/svaan1/go-tcc/internal/config"
	"github.com/svaan1/go-tcc/internal/grpcserver"
)

func main() {
	grpcAddress := net.JoinHostPort("", config.ServerPortGRPC)
	httpAddress := net.JoinHostPort("", config.ServerPortHTTP)

	sv := grpcserver.New(grpcAddress)

	log.Printf("Starting HTTP server at %s", httpAddress)

	if err := sv.Serve(); err != nil {
		log.Fatalf("TCP server failed: %v", err)
	}
}
