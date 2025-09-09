package server

import (
	"fmt"
	"time"

	"github.com/svaan1/go-tcc/internal/app"
	pb "github.com/svaan1/go-tcc/internal/grpc/transcoding"
)

func (sv *Server) RegisterNode(stream pb.VideoTranscoding_StreamServer) (*app.Node, *NodeConn, error) {
	msg, err := stream.Recv()
	if err != nil {
		return nil, nil, fmt.Errorf("error receiving initial message: %v", err)
	}

	register := msg.GetRegisterRequest()
	if register == nil {
		return nil, nil, fmt.Errorf("expected RegisterRequest but got different message type")
	}

	node, err := sv.Service.RegisterNode(register.Name, register.Codecs, time.Now())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to register node: %v", err)
	}

	nodeConn := newNodeConn(node.ID, stream)

	sv.mu.Lock()
	sv.NodeConns[node.ID] = nodeConn
	sv.mu.Unlock()

	if err := nodeConn.SendRegisterResponse(); err != nil {
		return nil, nil, fmt.Errorf("error sending registration response to %s: %v", node.Name, err)
	}

	return node, nodeConn, nil
}
