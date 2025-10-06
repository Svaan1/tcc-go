package js

import (
	"context"

	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/orchestrator/jq"
	"github.com/svaan1/tcc-go/internal/orchestrator/jt"
	"github.com/svaan1/tcc-go/internal/orchestrator/np"
)

type JobScheduler interface {
	RankNodes(ctx context.Context, job *jq.Job, nodes []*np.Node) ([]*np.Node, error)
	SelectBestNode(job *jq.Job, availableNodes []*np.Node, activeJobs map[uuid.UUID][]*jt.JobProgress) (*np.Node, error)
}
