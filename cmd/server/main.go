package main

import (
	"log"
	"net"

	"github.com/svaan1/go-tcc/internal/config"
	"github.com/svaan1/go-tcc/internal/grpc/server"
)

func main() {
	grpcAddress := net.JoinHostPort("", config.ServerPortGRPC)
	httpAddress := net.JoinHostPort("", config.ServerPortHTTP)

	sv := server.New(grpcAddress)

	log.Printf("Starting HTTP server at %s", httpAddress)

	if err := sv.Serve(); err != nil {
		log.Fatalf("TCP server failed: %v", err)
	}
}
