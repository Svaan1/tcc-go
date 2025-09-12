package np

import (
	"context"

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
	RegisterNode(ctx context.Context, req *NodeRegistration) (*Node, error)
	UnregisterNode(ctx context.Context, nodeID uuid.UUID) error

	UpdateNodeMetrics(ctx context.Context, nodeID uuid.UUID, usage *metrics.ResourceUsage) error

	ListNodes(ctx context.Context, offset, limit int) ([]*Node, error)
	GetNode(ctx context.Context, nodeID uuid.UUID) (*Node, error)
	GetAvailableNodes(ctx context.Context, requirements *NodeFilter) ([]*Node, error)
}
