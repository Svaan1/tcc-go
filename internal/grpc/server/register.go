package server

import (
	"fmt"
	"log"
	"time"

	"github.com/svaan1/go-tcc/internal/app"
	pb "github.com/svaan1/go-tcc/internal/grpc/transcoding"
)

func (sv *Server) RegisterNode(stream pb.VideoTranscoding_StreamServer) (*app.Node, *NodeConn, error) {
	msg, err := stream.Recv()
	if err != nil {
		log.Printf("Error receiving initial message: %v", err)
		return nil, nil, err
	}

	register := msg.GetRegisterRequest()
	if register == nil {
		log.Printf("Invalid message: expected RegisterRequest but got different message type")
		return nil, nil, fmt.Errorf("expected RegisterRequest")
	}

	node, err := sv.Service.RegisterNode(register.Name, register.Codecs, time.Now())
	if err != nil {
		log.Printf("Error registering node: %v", err)
		return nil, nil, err
	}

	nodeConn := newNodeConn(node.ID, stream)

	sv.mu.Lock()
	sv.NodeConns[node.ID] = nodeConn
	sv.mu.Unlock()

	if err := nodeConn.SendRegisterResponse(); err != nil {
		log.Printf("Error sending registration response to node %s: %v", node.Name, err)
		return nil, nil, err
	}

	log.Printf("Node %s (%s) successfully registered with %d codecs", node.Name, node.ID.String(), len(node.Codecs))

	return node, nodeConn, nil
}
