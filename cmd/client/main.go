package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/svaan1/go-tcc/internal/client"
	"github.com/svaan1/go-tcc/pkg/utils"
)

var (
	address     = utils.GetEnv("SERVER_ADDRESS", "localhost:8080")
	name        = utils.GetEnv("NODE_NAME", "node")
	codecString = utils.GetEnv("CODECS", "x264;x265")
)

func main() {
	codecs := strings.Split(codecString, ";")

	ctx, cancel := context.WithCancel(context.Background())
	client := client.New(address)

	err := client.Connect(ctx, name, codecs)
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
