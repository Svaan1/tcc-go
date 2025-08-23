package grpcserver

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/svaan1/go-tcc/internal/app"
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

	Config    ServerConfig
	App       *app.Service
	NodeConns map[uuid.UUID]*NodeConn

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
		App:       app.NewService(),
		NodeConns: map[uuid.UUID]*NodeConn{},
		listener:  nil,
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
		now := time.Now()
		ids := sv.App.TimedOutIDs(now, sv.Config.ResourceUsagePollingTimeout)
		if len(ids) == 0 {
			continue
		}

		sv.mu.Lock()
		for _, id := range ids {
			if conn, ok := sv.NodeConns[id]; ok {
				select {
				case <-conn.closedChan:
				default:
					close(conn.closedChan)
				}
			}
		}
		sv.mu.Unlock()
	}
}
