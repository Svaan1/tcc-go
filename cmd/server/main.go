package main

import (
	"context"
	"log"
	"net"

	"github.com/svaan1/tcc-go/internal/config"
	"github.com/svaan1/tcc-go/internal/grpc/server"
)

func main() {
	grpcAddress := net.JoinHostPort("", config.ServerPortGRPC)

	ctx := context.Background()

	sv := server.New(grpcAddress)

	if err := sv.Serve(ctx); err != nil {
		log.Fatalf("TCP server failed: %v", err)
	}
}
