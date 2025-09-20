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

func (sv *Server) Serve() error {
	listener, err := net.Listen(sv.Config.Network, sv.Config.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s %s: %w", sv.Config.Network, sv.Config.Address, err)
	}

	log.Printf("Server listening on %s %s", sv.Config.Network, sv.Config.Address)

	server := grpc.NewServer()
	pb.RegisterVideoTranscodingServer(server, sv)

	reflection.Register(server)

	go sv.pollJobs()
	go sv.trackTimedOutNodes()

	log.Println("Starting gRPC server...")
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}

func (sv *Server) pollJobs() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		job, node, err := sv.Service.DequeueJob(context.TODO())
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
			InputPath:  job.Params.InputPath,
			OutputPath: job.Params.OutputPath,
			VideoCodec: job.Params.VideoCodec,
			AudioCodec: job.Params.AudioCodec,
			Crf:        job.Params.Crf,
			Preset:     job.Params.Preset,
		})

		sv.mu.Unlock()
	}
}

func (sv *Server) trackTimedOutNodes() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		nodes, err := sv.Service.GetTimedOutNodes(context.TODO(), 15*time.Second)

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
