package js

import (
	"context"
	"fmt"
	"sync"

	"github.com/svaan1/tcc-go/internal/orchestrator/jq"
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

func (rr *RoundRobinScheduler) SelectBestNode(job *jq.Job, availableNodes []*np.Node) (*np.Node, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if len(availableNodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	if rr.nextIdx >= len(availableNodes) {
		rr.nextIdx = 0
	}

	bestNode := availableNodes[rr.nextIdx]
	rr.nextIdx++

	return bestNode, nil
}
