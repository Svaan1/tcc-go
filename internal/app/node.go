package app

import (
	"time"

	"github.com/google/uuid"
)

type ResourceUsage struct {
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryPercent float64   `json:"memory_percent"`
	DiskPercent   float64   `json:"disk_percent"`
	Timestamp     time.Time `json:"timestamp"`
}

type Node struct {
	ID            uuid.UUID     `json:"id"`
	Name          string        `json:"name"`
	Codecs        []string      `json:"codecs"`
	ResourceUsage ResourceUsage `json:"resource_usage"`
}
