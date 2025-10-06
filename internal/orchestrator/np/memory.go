package np

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/metrics"
)

type InMemoryNodePool struct {
	nodes map[uuid.UUID]*Node

	mu sync.RWMutex
}

func NewInMemoryNodePool() *InMemoryNodePool {
	return &InMemoryNodePool{
		nodes: make(map[uuid.UUID]*Node),
	}
}

func (p *InMemoryNodePool) RegisterNode(ctx context.Context, req *NodeRegistration) (uuid.UUID, error) {
	if req == nil || req.Name == "" {
		return uuid.Nil, fmt.Errorf("invalid node registration")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	node := &Node{
		ID:       uuid.New(),
		Name:     req.Name,
		Profiles: req.Profiles,
		ResourceUsage: &metrics.ResourceUsage{
			CPUUsagePercent:    0,
			MemoryUsagePercent: 0,
			DiskUsagePercent:   0,
		},
		HeartBeat: time.Now(),
	}

	p.nodes[node.ID] = node
	return node.ID, nil
}

func (p *InMemoryNodePool) UnregisterNode(ctx context.Context, nodeID uuid.UUID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.nodes[nodeID]; !exists {
		return fmt.Errorf("node not found")
	}

	delete(p.nodes, nodeID)
	return nil
}

func (p *InMemoryNodePool) UpdateNodeMetrics(ctx context.Context, nodeID uuid.UUID, usage *metrics.ResourceUsage) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	node, exists := p.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node not found")
	}

	node.ResourceUsage.CPUUsagePercent = usage.CPUUsagePercent
	node.ResourceUsage.MemoryUsagePercent = usage.MemoryUsagePercent
	node.ResourceUsage.DiskUsagePercent = usage.DiskUsagePercent

	node.HeartBeat = time.Now()
	return nil
}

func (p *InMemoryNodePool) ListNodes(ctx context.Context, offset, limit int) ([]*Node, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	nodes := make([]*Node, 0, len(p.nodes))
	for _, node := range p.nodes {
		nodes = append(nodes, node)
	}

	start := offset
	if start > len(nodes) {
		return []*Node{}, nil
	}

	end := start + limit
	if end > len(nodes) {
		end = len(nodes)
	}

	return nodes[start:end], nil
}

func (p *InMemoryNodePool) GetNode(ctx context.Context, nodeID uuid.UUID) (*Node, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	node, exists := p.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found")
	}

	return node, nil
}

func (p *InMemoryNodePool) GetAvailableNodes(ctx context.Context, requirements *NodeFilter) ([]*Node, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var availableNodes []*Node

	for _, node := range p.nodes {
		if p.nodeMatchesRequirements(node, requirements) {
			availableNodes = append(availableNodes, node)
		}
	}

	return availableNodes, nil
}

func (p *InMemoryNodePool) nodeMatchesRequirements(node *Node, requirements *NodeFilter) bool {
	if requirements == nil || len(requirements.ProfileName) == 0 {
		return true
	}

	for _, profile := range node.Profiles {
		if profile.Name == requirements.ProfileName {
			return true
		}
	}

	return false
}

func (p *InMemoryNodePool) GetTimedOutNodes(ctx context.Context, timeout time.Duration) ([]*Node, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("invalid timeout")
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	now := time.Now()
	var timedOut []*Node
	for _, node := range p.nodes {
		if now.Sub(node.HeartBeat) > timeout {
			timedOut = append(timedOut, node)
		}
	}

	return timedOut, nil
}
