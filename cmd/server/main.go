package main

import (
	"log"
	"net/http"

	"github.com/svaan1/go-tcc/internal/api"
	"github.com/svaan1/go-tcc/internal/server"
)

func main() {
	sv := server.New()

	log.Printf("Starting %s connection at %s", sv.Config.Network, sv.Config.Address)

	err := sv.Listen()
	if err != nil {
		log.Fatalf("Failed to start server %v", err)
	}

	log.Println("Starting API server at localhost:8082")

	handlers := api.NewHandlers(sv)
	http.HandleFunc("/nodes", handlers.GetNodes)

	log.Fatal(http.ListenAndServe(":8082", nil))
}
