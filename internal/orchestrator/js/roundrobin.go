package js

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/orchestrator/jq"
	"github.com/svaan1/tcc-go/internal/orchestrator/jt"
	"github.com/svaan1/tcc-go/internal/orchestrator/np"
)

type RoundRobinScheduler struct {
	nextIdx int

	mu sync.Mutex
}

func NewRoundRobinScheduler() *RoundRobinScheduler {
	return &RoundRobinScheduler{}
}

func (rr *RoundRobinScheduler) RankNodes(ctx context.Context, job *jq.Job, nodes []*np.Node) ([]*np.Node, error) {
	// Round-robin does not rank nodes, so we return the original list.
	return nodes, nil
}

func (rr *RoundRobinScheduler) SelectBestNode(job *jq.Job, availableNodes []*np.Node, activeJobs map[uuid.UUID][]*jt.JobProgress) (*np.Node, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if len(availableNodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	// Count jobs per node
	nodeJobCount := make(map[uuid.UUID]int)
	for _, jobs := range activeJobs {
		for _, jobProgress := range jobs {
			nodeJobCount[jobProgress.NodeID]++
		}
	}

	// Try to find a node with zero jobs
	// Start from nextIdx and wrap around
	startIdx := rr.nextIdx
	for i := range availableNodes {
		idx := (startIdx + i) % len(availableNodes)
		node := availableNodes[idx]

		if nodeJobCount[node.ID] == 0 {
			rr.nextIdx = (idx + 1) % len(availableNodes)
			return node, nil
		}
	}

	// If all nodes have at least one job, return error to requeue
	return nil, fmt.Errorf("no available nodes without active jobs")
}
