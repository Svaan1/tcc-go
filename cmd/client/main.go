package main

import (
	"context"
	"log"
	"strings"

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

	ctx := context.Background()
	client := client.New(address)

	err := client.Connect(ctx, name, codecs)
	if err != nil {
		log.Fatalf("Failed to connect to server %v", err)
	}

	log.Printf("Connected to server at %s", address)

	select {}
}
