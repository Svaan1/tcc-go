package np

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type NodeRegistration struct {
	Name         string            `json:"name"`
	Capabilities *NodeCapabilities `json:"capabilities"`
}

type NodeCapabilities struct {
	Codecs            []string             `json:"codecs"`
	MaxConcurrentJobs int                  `json:"max_concurrent_jobs"`
	GPU               *GPUCapabilities     `json:"gpu,omitempty"`
	CPU               *CPUCapabilities     `json:"cpu"`
	Memory            *MemoryCapabilities  `json:"memory"`
	Storage           *StorageCapabilities `json:"storage"`
	Network           *NetworkCapabilities `json:"network"`
}

type NodeRequirements struct {
	RequiredCodecs []string      `json:"required_codecs"`
	MinCPU         float64       `json:"min_cpu"`
	MinMemory      int64         `json:"min_memory"`
	MinStorage     int64         `json:"min_storage"`
	GPURequired    bool          `json:"gpu_required"`
	MaxLatency     time.Duration `json:"max_latency"`
}

type HealthStatus struct {
	Status          HealthStatusType `json:"status"`
	CPUUsage        float64          `json:"cpu_usage"`
	MemoryUsage     float64          `json:"memory_usage"`
	DiskUsage       float64          `json:"disk_usage"`
	ActiveJobs      int              `json:"active_jobs"`
	LastHealthcheck time.Time        `json:"last_healthcheck"`
	Errors          []string         `json:"errors,omitempty"`
}

type NodePool interface {
	// Node registration
	RegisterNode(ctx context.Context, req *NodeRegistration) (*Node, error)
	UnregisterNode(ctx context.Context, nodeID uuid.UUID) error

	// Node discovery and querying
	GetNode(ctx context.Context, nodeID uuid.UUID) (*Node, error)
	ListNodes(ctx context.Context, filter NodeFilter) ([]*Node, error)
	GetAvailableNodes(ctx context.Context, requirements *NodeRequirements) ([]*Node, error)

	// Node health and status
	UpdateNodeHealth(ctx context.Context, nodeID uuid.UUID, health *HealthStatus) error
	GetNodeHealth(ctx context.Context, nodeID uuid.UUID) (*HealthStatus, error)
	MarkNodeOffline(ctx context.Context, nodeID uuid.UUID, reason string) error

	// Capability management
	UpdateNodeCapabilities(ctx context.Context, nodeID uuid.UUID, caps *NodeCapabilities) error
	GetNodesByCapability(ctx context.Context, capability string) ([]*Node, error)
}
