package main

import (
	"context"
	"log"

	"github.com/svaan1/go-tcc/internal/client"
)

func main() {
	ctx := context.Background()
	client := client.New()

	err := client.Connect(ctx, "node-1", []string{"x265"})
	if err != nil {
		log.Fatalf("Failed to connect to server %v", err)
	}

	select {}
}
