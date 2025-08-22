package server

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	pb "github.com/svaan1/go-tcc/internal/transcoding"
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

	node := &Node{
		ID:     uuid.New(),
		Name:   register.Name,
		Codecs: register.Codecs,
		ResourceUsage: ResourceUsage{
			Timestamp: time.Now(),
		},

		logger:     log.New(os.Stdout, fmt.Sprintf("[%s] ", register.Name), log.LstdFlags),
		stream:     stream,
		closedChan: make(chan struct{}),
	}

	sv.mu.Lock()
	sv.Nodes[node.ID] = node
	sv.mu.Unlock()

	if err := node.SendRegisterResponse(); err != nil {
		log.Printf("Error sending registration response to node %s: %v", node.Name, err)
		return nil, err
	}

	log.Printf("Node %s (%s) successfully registered with %d codecs", node.Name, node.ID.String(), len(node.Codecs))

	return node, nil
}
