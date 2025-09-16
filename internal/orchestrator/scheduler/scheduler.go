package scheduler

import (
	"fmt"
	"sort"

	"github.com/svaan1/tcc-go/internal/orchestrator/jq"
	"github.com/svaan1/tcc-go/internal/orchestrator/np"
)

type JobScheduler struct {
	jp JobPriorityStrategy
	ns NodeSelectionStrategy
}

func NewJobScheduler(jp JobPriorityStrategy, ns NodeSelectionStrategy) *JobScheduler {
	return &JobScheduler{
		jp: jp,
		ns: ns,
	}
}

func (js *JobScheduler) RankJobs(jobs []*jq.Job) ([]JobPriorityScore, error) {
	var scores []JobPriorityScore
	for _, job := range jobs {
		scores = append(scores, js.jp.ScoreJobPriority(job))
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return scores, nil
}

func (js *JobScheduler) RankNodes(job *jq.Job, nodes []*np.Node) ([]NodeScore, error) {
	var scores []NodeScore
	for _, node := range nodes {
		scores = append(scores, js.ns.ScoreNode(job, node))
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return scores, nil
}

func (js *JobScheduler) SelectBestNode(job *jq.Job, availableNodes []*np.Node) (NodeScore, error) {
	var bestScore NodeScore

	for _, node := range availableNodes {
		nodeScore := js.ns.ScoreNode(job, node)

		if bestScore.Node == nil || nodeScore.Score > bestScore.Score {
			bestScore = nodeScore
		}
	}

	if bestScore.Node == nil {
		return NodeScore{}, fmt.Errorf("could not find any node for the job")
	}

	return bestScore, nil
}
