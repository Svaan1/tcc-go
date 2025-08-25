package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/svaan1/go-tcc/internal/app"
	pb "github.com/svaan1/go-tcc/internal/transcoding"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	server := grpc.NewServer()
	pb.RegisterVideoTranscodingServer(server, sv)

	reflection.Register(server)

	go sv.trackTimedOutNodes()

	log.Println("Starting gRPC server...")
	if err := server.Serve(listener); err != nil {
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

// GetAllNodes returns all registered nodes with their current resource usage
func (sv *Server) GetAllNodes(ctx context.Context, req *pb.GetAllNodesRequest) (*pb.GetAllNodesResponse, error) {
	nodes := sv.App.ListNodes(ctx)

	nodeInfos := make([]*pb.NodeInfo, len(nodes))
	for i, node := range nodes {
		nodeInfos[i] = &pb.NodeInfo{
			NodeId:        node.ID.String(),
			Name:          node.Name,
			Codecs:        node.Codecs,
			CpuPercent:    node.ResourceUsage.CPUPercent,
			MemoryPercent: node.ResourceUsage.MemoryPercent,
			DiskPercent:   node.ResourceUsage.DiskPercent,
			LastSeen:      timestamppb.New(node.ResourceUsage.Timestamp),
		}
	}

	return &pb.GetAllNodesResponse{
		Nodes:      nodeInfos,
		TotalCount: int32(len(nodes)),
	}, nil
}

// EnqueueJob enqueues a new transcoding job to be assigned to an available node
func (sv *Server) EnqueueJob(ctx context.Context, req *pb.EnqueueJobRequest) (*pb.EnqueueJobResponse, error) {
	jobID := uuid.New().String()

	// Pick an available node for the job
	nodeID, err := sv.App.PickNodeForJob(ctx)
	if err != nil {
		return &pb.EnqueueJobResponse{
			JobId:   jobID,
			Success: false,
			Message: fmt.Sprintf("No available nodes for job: %v", err),
		}, nil
	}

	// Find the node connection to send the job assignment
	sv.mu.RLock()
	nodeConn, exists := sv.NodeConns[nodeID]
	sv.mu.RUnlock()

	if !exists {
		return &pb.EnqueueJobResponse{
			JobId:   jobID,
			Success: false,
			Message: "Selected node is not connected",
		}, nil
	}

	// Create job assignment request
	jobAssignment := &pb.OrchestratorMessage{
		Base: &pb.MessageBase{
			MessageId: "job-assignment-" + jobID,
			Timestamp: timestamppb.Now(),
		},
		Payload: &pb.OrchestratorMessage_JobAssignmentRequest{
			JobAssignmentRequest: &pb.JobAssignmentRequest{
				JobId:      jobID,
				InputPath:  req.InputPath,
				OutputPath: req.OutputPath,
				Crf:        req.Crf,
				Preset:     req.Preset,
				AudioCodec: req.AudioCodec,
				VideoCodec: req.VideoCodec,
			},
		},
	}

	// Send job assignment to the selected node
	err = nodeConn.stream.Send(jobAssignment)
	if err != nil {
		return &pb.EnqueueJobResponse{
			JobId:   jobID,
			Success: false,
			Message: fmt.Sprintf("Failed to send job to node: %v", err),
		}, nil
	}

	log.Printf("Job %s assigned to node %s", jobID, nodeID.String())

	return &pb.EnqueueJobResponse{
		JobId:          jobID,
		Success:        true,
		Message:        "Job successfully enqueued and assigned",
		AssignedNodeId: nodeID.String(),
	}, nil
}
