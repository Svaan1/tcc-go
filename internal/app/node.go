package app

import (
	"errors"
	"sort"
	"sync"
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

type NodeManager struct {
	ResourceUsagePollingTimeout time.Duration

	nodes map[uuid.UUID]*Node
	mu    sync.RWMutex
}

func NewNodeManager() *NodeManager {
	return &NodeManager{
		ResourceUsagePollingTimeout: 10 * time.Second,

		nodes: make(map[uuid.UUID]*Node),
	}
}

func (nm *NodeManager) RegisterNode(name string, codecs []string, now time.Time) (*Node, error) {
	n := &Node{
		ID:     uuid.New(),
		Name:   name,
		Codecs: codecs,
	}

	nm.mu.Lock()
	n.ResourceUsage.Timestamp = now
	nm.nodes[n.ID] = n
	nm.mu.Unlock()

	return n, nil
}

func (nm *NodeManager) RemoveNode(id uuid.UUID) {
	nm.mu.Lock()
	delete(nm.nodes, id)
	nm.mu.Unlock()
}

func (nm *NodeManager) GetNode(id uuid.UUID) (*Node, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	node, ok := nm.nodes[id]
	return node, ok
}

func (nm *NodeManager) ListNodes() []*Node {
	nm.mu.RLock()
	out := make([]*Node, 0, len(nm.nodes))
	for _, n := range nm.nodes {
		out = append(out, n)
	}
	nm.mu.RUnlock()

	// stable order (by Name) for deterministic APIs
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (nm *NodeManager) UpdateResourceUsage(id uuid.UUID, usage ResourceUsage, ts time.Time) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	node, ok := nm.nodes[id]
	if !ok {
		return errors.New("node not found")
	}

	usage.Timestamp = ts
	node.ResourceUsage = usage
	return nil
}

func (nm *NodeManager) GetTimedOutNodes(now time.Time) []uuid.UUID {
	nm.mu.RLock()
	ids := make([]uuid.UUID, 0)
	for id, n := range nm.nodes {
		if now.Sub(n.ResourceUsage.Timestamp) > nm.ResourceUsagePollingTimeout {
			ids = append(ids, id)
		}
	}
	nm.mu.RUnlock()
	return ids
}

// GetAvailableNodes returns nodes that support the given codec and meet resource criteria.
func (nm *NodeManager) GetAvailableNodes(requiredCodec string) []*Node {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	available := make([]*Node, 0)
	for _, node := range nm.nodes {
		if nm.nodeSupportsCodec(node, requiredCodec) {
			available = append(available, node)
		}
	}
	return available
}

func (nm *NodeManager) nodeSupportsCodec(node *Node, codec string) bool {
	for _, supportedCodec := range node.Codecs {
		if supportedCodec == codec {
			return true
		}
	}
	return false
}
