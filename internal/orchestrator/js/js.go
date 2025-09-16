package js

import (
	"context"

	"github.com/svaan1/tcc-go/internal/orchestrator/jq"
	"github.com/svaan1/tcc-go/internal/orchestrator/np"
)

type JobScheduler interface {
	RankNodes(ctx context.Context, job *jq.Job, nodes []*np.Node) ([]*np.Node, error)
	SelectBestNode(job *jq.Job, availableNodes []*np.Node) (*np.Node, error)
}
