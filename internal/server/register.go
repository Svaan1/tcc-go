package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/svaan1/go-tcc/internal/transcoding"
)

type ResourceUsage struct {
	ResourceUsageRequest pb.ResourceUsageRequest
	Timestamp            time.Time
}

type Node struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Codecs        []string  `json:"codecs"`
	ResourceUsage ResourceUsage

	stream     *pb.VideoTranscoding_StreamServer
	logger     *log.Logger
	closedOnce sync.Once
	closedChan chan struct{}
}

func (sv *Server) Stream(stream pb.VideoTranscoding_StreamServer) error {
	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	register := msg.GetRegisterRequest()
	if register == nil {
		return fmt.Errorf("expected RegisterRequest")
	}

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

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		log.Print(msg)

		orchestratorMsg := &pb.OrchestratorMessage{}

		if err := stream.Send(orchestratorMsg); err != nil {
			return err
		}
	}
}
