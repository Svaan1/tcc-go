package scheduling

import (
	"github.com/svaan1/go-tcc/internal/orchestrator/jq"
	"github.com/svaan1/go-tcc/internal/orchestrator/np"
)

type NodeScore struct {
	Node  *np.Node `json:"node"`
	Score float64  `json:"score"` // Higher is better
}

type SelectionStrategy interface {
	ScoreNode(job *jq.Job, node *np.Node) *NodeScore
}
