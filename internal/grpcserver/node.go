package grpcserver

import (
	"github.com/google/uuid"
	pb "github.com/svaan1/go-tcc/internal/transcoding"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NodeConn struct {
	ID uuid.UUID

	stream     pb.VideoTranscoding_StreamServer
	closedChan chan struct{}
}

func newNodeConn(id uuid.UUID, stream pb.VideoTranscoding_StreamServer) *NodeConn {
	return &NodeConn{
		ID:         id,
		stream:     stream,
		closedChan: make(chan struct{}),
	}
}

func (n *NodeConn) SendRegisterResponse() error {
	msg := &pb.OrchestratorMessage{
		Base: &pb.MessageBase{
			MessageId: "register-response-" + n.ID.String(),
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

func (n *NodeConn) SendJobAssignmentRequest(req *pb.JobAssignmentRequest) error {
	msg := &pb.OrchestratorMessage{
		Base: &pb.MessageBase{
			MessageId: "job-assignment-" + req.JobId,
			Timestamp: timestamppb.Now(),
		},
		Payload: &pb.OrchestratorMessage_JobAssignmentRequest{
			JobAssignmentRequest: req,
		},
	}
	return n.stream.Send(msg)
}
