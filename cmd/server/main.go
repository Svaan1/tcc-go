package main

import (
	"log"
	"net"
	"net/http"

	"github.com/svaan1/go-tcc/internal/api"
	"github.com/svaan1/go-tcc/internal/server"
	"github.com/svaan1/go-tcc/pkg/utils"
)

var (
	tcpPort  = utils.GetEnv("TCP_PORT", "8080")
	httpPort = utils.GetEnv("HTTP_PORT", "8081")
)

func main() {
	tcpAddress := net.JoinHostPort("", tcpPort)
	httpAddress := net.JoinHostPort("", httpPort)

	sv := server.New(tcpAddress)

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
