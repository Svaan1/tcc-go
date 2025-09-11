package scheduling

import (
	"fmt"

	"github.com/svaan1/tcc-go/internal/orchestrator/jq"
	"github.com/svaan1/tcc-go/internal/orchestrator/np"
)

// NodeScheduler implements the strategy for choosing optimal nodes
type NodeScheduler struct {
	strategy SelectionStrategy
}

func NewNodeScheduler(strategy SelectionStrategy) *NodeScheduler {
	return &NodeScheduler{
		strategy: strategy,
	}
}

func (ns *NodeScheduler) SelectBestNode(job *jq.Job, availableNodes []*np.Node) (*NodeScore, error) {
	var bestScore *NodeScore

	for _, node := range availableNodes {
		nodeScore := ns.strategy.ScoreNode(job, node)

		if bestScore == nil || nodeScore.Score > bestScore.Score {
			bestScore = nodeScore
		}
	}

	if bestScore == nil {
		return nil, fmt.Errorf("could not find any node for the job")
	}

	return bestScore, nil
}

func (ns *NodeScheduler) RankNodes(job *jq.Job, nodes []*np.Node) ([]*NodeScore, error) {
	var scores []*NodeScore
	for _, node := range nodes {
		scores = append(scores, ns.strategy.ScoreNode(job, node))
	}

	return scores, nil
}

func (ns *NodeScheduler) SetStrategy(strategy SelectionStrategy) {
	ns.strategy = strategy
}

func (ns *NodeScheduler) GetStrategy() SelectionStrategy {
	return ns.strategy
}

// Learning and adaptation (TODO)
// RecordJobResult( nodeID uuid.UUID, jobResult *JobResult) error
// GetNodePerformanceHistory( nodeID uuid.UUID) (*PerformanceHistory, error)
