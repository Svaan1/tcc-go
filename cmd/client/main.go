package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/svaan1/go-tcc/internal/client"
	"github.com/svaan1/go-tcc/internal/config"
)

func main() {
	codecs := strings.Split(config.ClientCodecs, ";")

	address := net.JoinHostPort(config.ServerHostName, config.ServerPortGRPC)

	ctx, cancel := context.WithCancel(context.Background())
	client := client.New(address)

	err := client.Connect(ctx, config.ClientName, codecs)
	if err != nil {
		log.Fatalf("Failed to connect to server %v", err)
	}

	log.Printf("Connected to server at %s", address)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("Received shutdown signal, disconnecting...")
		cancel()
		client.Close()
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
		client.Close()
	}
}
