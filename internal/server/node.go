package server

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/svaan1/go-tcc/internal/transcoding"
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

	stream     *pb.VideoTranscoding_StreamServer
	logger     *log.Logger
	closedChan chan struct{}
	closed     bool
	mu         sync.Mutex
}

func (n *Node) SetResourceUsage(req *pb.ResourceUsageRequest, timestamp time.Time) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ResourceUsage = ResourceUsage{
		ResourceUsageRequest: req,
		Timestamp:            timestamp,
	}
}
