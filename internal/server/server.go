package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/svaan1/go-tcc/internal/transcoding"
	"google.golang.org/grpc"
)

type ServerConfig struct {
	Network string
	Address string

	ResourceUsagePollingTimeout time.Duration
}

type Server struct {
	pb.VideoTranscodingServer

	Config ServerConfig
	Nodes  map[uuid.UUID]*Node

	listener *net.Listener
	mu       sync.RWMutex
}

func New(address string) *Server {
	return &Server{
		Config: ServerConfig{
			Network: "tcp",
			Address: address,

			ResourceUsagePollingTimeout: 10 * time.Second,
		},
		Nodes:    map[uuid.UUID]*Node{},
		listener: nil,
	}

}

func (sv *Server) Serve() error {
	listener, err := net.Listen(sv.Config.Network, sv.Config.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s %s: %w", sv.Config.Network, sv.Config.Address, err)
	}

	sv.listener = &listener

	log.Printf("Server listening on %s %s", sv.Config.Network, sv.Config.Address)

	grpcServer := grpc.NewServer()
	pb.RegisterVideoTranscodingServer(grpcServer, sv)

	go sv.trackTimedOutNodes()

	log.Println("Starting gRPC server...")
	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}

func (sv *Server) trackTimedOutNodes() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sv.mu.Lock()
		for _, node := range sv.Nodes {
			if time.Since(node.ResourceUsage.Timestamp) > sv.Config.ResourceUsagePollingTimeout {
				select {
				case <-node.closedChan:
				default:
					close(node.closedChan)
				}
			}
		}
		sv.mu.Unlock()
	}
}
