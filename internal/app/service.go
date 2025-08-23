package app

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Service owns domain state and policies for the orchestrator.
type Service struct {
	nodes map[uuid.UUID]*Node

	mu sync.RWMutex
}

func NewService() *Service {
	return &Service{
		nodes: make(map[uuid.UUID]*Node),
	}
}

func (s *Service) ListNodes(ctx context.Context) []*Node {
	_ = ctx
	s.mu.RLock()
	out := make([]*Node, 0, len(s.nodes))
	for _, n := range s.nodes {
		out = append(out, n)
	}
	s.mu.RUnlock()

	// stable order (by Name) for deterministic APIs
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *Service) RegisterNode(ctx context.Context, name string, codecs []string, now time.Time) (*Node, error) {
	_ = ctx // reserved for auth/cancel in future
	n := &Node{
		ID:     uuid.New(),
		Name:   name,
		Codecs: codecs,
	}
	s.mu.Lock()
	n.ResourceUsage.Timestamp = now
	s.nodes[n.ID] = n
	s.mu.Unlock()
	return n, nil
}

func (s *Service) RemoveNode(id uuid.UUID) {
	s.mu.Lock()
	delete(s.nodes, id)
	s.mu.Unlock()
}

func (s *Service) UpdateResourceUsage(ctx context.Context, id uuid.UUID, usage ResourceUsage, ts time.Time) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.nodes[id]
	if !ok {
		return errors.New("node not found")
	}

	usage.Timestamp = ts
	node.ResourceUsage = usage
	return nil
}

// TimedOutIDs returns node IDs whose lastSeen exceed timeout.
func (s *Service) TimedOutIDs(now time.Time, timeout time.Duration) []uuid.UUID {
	s.mu.RLock()
	ids := make([]uuid.UUID, 0)
	for id, n := range s.nodes {
		if now.Sub(n.ResourceUsage.Timestamp) > timeout {
			ids = append(ids, id)
		}
	}
	s.mu.RUnlock()
	return ids
}

// PickNodeForJob returns a node id using a simple policy (lowest CPU usage).
func (s *Service) PickNodeForJob(ctx context.Context) (uuid.UUID, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestID uuid.UUID
	bestCPU := 200.0 // larger than any real percent
	for id, n := range s.nodes {
		if n.ResourceUsage.CPUPercent < bestCPU {
			bestCPU = n.ResourceUsage.CPUPercent
			bestID = id
		}
	}

	if bestID == (uuid.UUID{}) {
		return uuid.Nil, errors.New("no node available")
	}

	return bestID, nil
}
