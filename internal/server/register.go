package server

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	pb "github.com/svaan1/go-tcc/internal/transcoding"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (sv *Server) RegisterNode(stream pb.VideoTranscoding_StreamServer) (*Node, error) {
	msg, err := stream.Recv()
	if err != nil {
		log.Printf("Error receiving initial message: %v", err)
		return nil, err
	}

	register := msg.GetRegisterRequest()
	if register == nil {
		log.Printf("Invalid message: expected RegisterRequest but got different message type")
		return nil, fmt.Errorf("expected RegisterRequest")
	}

	log.Printf("Processing registration request for node: %s with codecs: %v", register.Name, register.Codecs)

	node := &Node{
		ID:     uuid.New(),
		Name:   register.Name,
		Codecs: register.Codecs,
		ResourceUsage: ResourceUsage{
			Timestamp: time.Now(),
		},

		stream:     &stream,
		logger:     log.New(os.Stdout, fmt.Sprintf("[%s] ", register.Name), log.LstdFlags),
		closedChan: make(chan struct{}),
	}
	sv.Nodes[node.ID] = node

	log.Printf("Node %s (%s) successfully registered with %d codecs", node.Name, node.ID.String(), len(node.Codecs))

	registerResponse := &pb.OrchestratorMessage{
		Base: &pb.MessageBase{
			MessageId: "register-response-" + node.Name,
			Timestamp: timestamppb.Now(),
		},
		Payload: &pb.OrchestratorMessage_RegisterResponse{
			RegisterResponse: &pb.RegisterResponse{
				NodeId:  node.ID.String(),
				Success: true,
				Message: "Registered successfuly.",
			},
		},
	}

	if err := stream.Send(registerResponse); err != nil {
		log.Printf("Error sending registration response to node %s: %v", node.Name, err)
		return nil, err
	}
	log.Printf("Registration response sent successfully to node %s", node.Name)

	return node, nil
}
