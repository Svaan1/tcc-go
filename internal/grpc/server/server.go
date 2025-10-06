package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/svaan1/tcc-go/internal/grpc/proto"
	"github.com/svaan1/tcc-go/internal/orchestrator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type ServerConfig struct {
	Network string
	Address string
}

type Server struct {
	pb.VideoTranscodingServer

	Config    ServerConfig
	Service   *orchestrator.Service
	NodeConns map[uuid.UUID]*NodeConn

	mu sync.RWMutex
}

func New(address string) *Server {
	return &Server{
		Config: ServerConfig{
			Network: "tcp",
			Address: address,
		},
		Service:   orchestrator.NewService(),
		NodeConns: map[uuid.UUID]*NodeConn{},
	}

}

func (sv *Server) Serve(ctx context.Context) error {
	listener, err := net.Listen(sv.Config.Network, sv.Config.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s %s: %w", sv.Config.Network, sv.Config.Address, err)
	}

	server := grpc.NewServer()
	pb.RegisterVideoTranscodingServer(server, sv)

	reflection.Register(server)

	go sv.pollJobs(ctx)
	go sv.trackTimedOutNodes(ctx)

	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	log.Printf("Starting gRPC server at %s %s", sv.Config.Network, sv.Config.Address)

	return nil
}

func (sv *Server) pollJobs(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		job, node, err := sv.Service.DequeueJob(ctx)
		if err != nil {
			continue
		}

		if job == nil || node == nil {
			continue
		}

		sv.mu.Lock()

		conn, exists := sv.NodeConns[node.ID]
		if !exists {
			sv.mu.Unlock()
			continue
		}

		conn.SendJobAssignmentRequest(&pb.JobAssignmentRequest{
			JobId:       job.ID.String(),
			InputPath:   job.Params.InputPath,
			OutputPath:  job.Params.OutputPath,
			ProfileName: job.Params.ProfileName,
		})

		sv.mu.Unlock()
	}
}

func (sv *Server) trackTimedOutNodes(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		nodes, err := sv.Service.GetTimedOutNodes(ctx, 15*time.Second)

		if err != nil {
			log.Printf("Failed to get timed out nodes: %v", err)
		}

		if len(nodes) == 0 {
			continue
		}

		sv.mu.Lock()
		for _, node := range nodes {
			if conn, ok := sv.NodeConns[node.ID]; ok {
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
