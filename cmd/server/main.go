package main

import (
	"log"
	"net"
	"net/http"

	"github.com/svaan1/go-tcc/internal/api"
	"github.com/svaan1/go-tcc/internal/config"
	"github.com/svaan1/go-tcc/internal/grpcserver"
)

func main() {
	grpcAddress := net.JoinHostPort("", config.ServerPortGRPC)
	httpAddress := net.JoinHostPort("", config.ServerPortHTTP)

	sv := grpcserver.New(grpcAddress)

	go func() {
		if err := sv.Serve(); err != nil {
			log.Fatalf("TCP server failed: %v", err)
		}
	}()

	log.Printf("Starting HTTP server at %s", httpAddress)

	handlers := api.NewHandlers(sv)
	http.HandleFunc("/nodes", handlers.GetNodes)
	http.HandleFunc("/job", handlers.AssignJob)

	if err := http.ListenAndServe(httpAddress, nil); err != nil {
		log.Fatalf("Failed to start HTTP server %v", err)
	}
}
