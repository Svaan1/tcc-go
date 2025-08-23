package grpcserver

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/svaan1/go-tcc/internal/transcoding"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ResourceUsage struct {
	ResourceUsageRequest *pb.ResourceUsageRequest
	Timestamp            time.Time
}

type Node struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Codecs        []string  `json:"codecs"`
	ResourceUsage ResourceUsage

	logger     *log.Logger
	stream     pb.VideoTranscoding_StreamServer
	closedChan chan struct{}
	mu         sync.Mutex
}

func (n *Node) SendRegisterResponse() error {
	msg := &pb.OrchestratorMessage{
		Base: &pb.MessageBase{
			MessageId: "register-response-" + n.Name,
			Timestamp: timestamppb.Now(),
		},
		Payload: &pb.OrchestratorMessage_RegisterResponse{
			RegisterResponse: &pb.RegisterResponse{
				NodeId:  n.ID.String(),
				Success: true,
				Message: "Registered successfuly.",
			},
		},
	}

	return n.stream.Send(msg)
}

func (n *Node) SendJobAssignmentRequest(req *pb.JobAssignmentRequest) error {
	msg := pb.OrchestratorMessage{
		Base: &pb.MessageBase{
			MessageId: "job-assignment-" + req.JobId,
			Timestamp: timestamppb.Now(),
		},
		Payload: &pb.OrchestratorMessage_JobAssignmentRequest{
			JobAssignmentRequest: req,
		},
	}

	return n.stream.Send(&msg)
}

func (n *Node) SetResourceUsage(req *pb.ResourceUsageRequest, timestamp time.Time) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ResourceUsage = ResourceUsage{
		ResourceUsageRequest: req,
		Timestamp:            timestamp,
	}
}
