package np

import (
	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
	"github.com/svaan1/tcc-go/internal/metrics"
)

type NodeRegistration struct {
	Name   string                  `json:"name"`
	Codecs []ffmpeg.CodecBenchmark `json:"codecs"`
}

type Node struct {
	ID            uuid.UUID
	Name          string                  `json:"name"`
	Codecs        []ffmpeg.CodecBenchmark `json:"codecs"`
	ResourceUsage *metrics.ResourceUsage  `json:"resourceUsage"`
}

type NodeFilter struct {
	RequiredCodecs []string `json:"required_codecs"`
}

type NodePool interface {
	// Node registration
	RegisterNode(req *NodeRegistration) (*Node, error)
	UnregisterNode(nodeID uuid.UUID) error

	// Node discovery and querying
	GetNode(nodeID uuid.UUID) (*Node, error)
	ListNodes() []*Node
	GetAvailableNodes(requirements *NodeFilter) []*Node
}
