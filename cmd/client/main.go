package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/svaan1/tcc-go/internal/config"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
	"github.com/svaan1/tcc-go/internal/grpc/client"
)

func main() {
	address := net.JoinHostPort(config.ServerHostName, config.ServerPortGRPC)

	profiles := []ffmpeg.EncodingProfile{
		{
			Name:  "264-slow",
			Codec: "libx264",
			Params: []string{
				"-preset", "slow",
				"-crf", "23",
			},

			EncodeTime: 1,
			DecodeTime: 1,
			Score:      1,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	client := client.New(address)

	err := client.Connect(ctx, config.ClientName, profiles)
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
