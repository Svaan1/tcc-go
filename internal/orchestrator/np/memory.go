package np

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type InMemoryNodePool struct {
	NodePool

	nodes map[uuid.UUID]*Node
	mu    sync.RWMutex
}

func NewInMemoryNodePool() *InMemoryNodePool {
	return &InMemoryNodePool{
		nodes: make(map[uuid.UUID]*Node),
	}
}

func (p *InMemoryNodePool) RegisterNode(req *NodeRegistration) (*Node, error) {
	if req == nil || req.Name == "" {
		return nil, fmt.Errorf("invalid node registration: name is required")
	}

	nodeID := uuid.New()
	node := &Node{
		Name:   req.Name,
		Codecs: req.Codecs,
	}

	p.mu.Lock()
	p.nodes[nodeID] = node
	p.mu.Unlock()

	return node, nil
}

func (p *InMemoryNodePool) UnregisterNode(nodeID uuid.UUID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.nodes[nodeID]; !exists {
		return fmt.Errorf("node with ID %s not found", nodeID)
	}

	delete(p.nodes, nodeID)
	return nil
}

func (p *InMemoryNodePool) GetNode(nodeID uuid.UUID) (*Node, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	node, exists := p.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node with ID %s not found", nodeID)
	}

	return node, nil
}

func (p *InMemoryNodePool) ListNodes() []*Node {
	p.mu.RLock()
	defer p.mu.RUnlock()

	nodes := make([]*Node, 0, len(p.nodes))
	for _, node := range p.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

func (p *InMemoryNodePool) GetAvailableNodes(requirements *NodeFilter) []*Node {
	if requirements == nil {
		return p.ListNodes()
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	var availableNodes []*Node

	for _, node := range p.nodes {
		if p.nodeSupportsCodecs(node, requirements.RequiredCodecs) {
			availableNodes = append(availableNodes, node)
		}
	}

	return availableNodes
}

func (p *InMemoryNodePool) nodeSupportsCodecs(node *Node, requiredCodecs []string) bool {
	if len(requiredCodecs) == 0 {
		return true
	}

	codecMap := make(map[string]bool)
	for _, benchmark := range node.Codecs {
		codecMap[benchmark.Codec.Name] = true
	}

	for _, required := range requiredCodecs {
		if !codecMap[required] {
			return false
		}
	}

	return true
}
