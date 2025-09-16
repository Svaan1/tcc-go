package scheduler

import (
	"github.com/svaan1/tcc-go/internal/orchestrator/jq"
	"github.com/svaan1/tcc-go/internal/orchestrator/np"
)

type JobPriorityScore struct {
	Job   *jq.Job `json:"job"`
	Score float64 `json:"score"`
}

type NodeScore struct {
	Node  *np.Node `json:"node"`
	Score float64  `json:"score"` // Higher is better
}

type JobPriorityStrategy interface {
	ScoreJobPriority(job *jq.Job) JobPriorityScore
}

type NodeSelectionStrategy interface {
	ScoreNode(job *jq.Job, node *np.Node) NodeScore
}
