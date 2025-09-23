package np

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
	"github.com/svaan1/tcc-go/internal/metrics"
)

type NodeRegistration struct {
	Name     string                   `json:"name"`
	Profiles []ffmpeg.EncodingProfile `json:"codecs"`
}

type Node struct {
	ID            uuid.UUID
	Name          string                   `json:"name"`
	Profiles      []ffmpeg.EncodingProfile `json:"codecs"`
	ResourceUsage *metrics.ResourceUsage   `json:"resourceUsage"`
	HeartBeat     time.Time
}

type NodeFilter struct {
	Codec string `json:"codec"`
}

type NodePool interface {
	RegisterNode(ctx context.Context, req *NodeRegistration) (uuid.UUID, error)
	UnregisterNode(ctx context.Context, nodeID uuid.UUID) error

	UpdateNodeMetrics(ctx context.Context, nodeID uuid.UUID, usage *metrics.ResourceUsage) error

	ListNodes(ctx context.Context, offset, limit int) ([]*Node, error)
	GetNode(ctx context.Context, nodeID uuid.UUID) (*Node, error)
	GetAvailableNodes(ctx context.Context, requirements *NodeFilter) ([]*Node, error)
	GetTimedOutNodes(ctx context.Context, timeout time.Duration) ([]*Node, error)
}
