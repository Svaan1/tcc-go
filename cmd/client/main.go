package main

import (
	"log"

	"github.com/svaan1/go-tcc/internal/client"
)

func main() {
	client := client.New()

	err := client.Connect("node-1", []string{"x265"})
	if err != nil {
		log.Fatalf("Failed to connect to server %v", err)
	}

	select {}
}
