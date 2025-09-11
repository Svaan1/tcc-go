package orchestrator

import (
	"time"

	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
	"github.com/svaan1/tcc-go/internal/orchestrator/jq"
	"github.com/svaan1/tcc-go/internal/orchestrator/np"
)

type Service struct {
	jq jq.JobQueue
	np np.NodePool
}

func NewService() *Service {
	return &Service{
		jq: jq.NewInMemoryJobQueue(),
		np: np.NewInMemoryNodePool(),
	}
}

func (s *Service) ListNodes() []*np.Node {
	return nil
}

func (s *Service) RegisterNode(name string, codecs []string, ts time.Time) (*np.Node, error) {
	return nil, nil
}

func (s *Service) UnregisterNode(ID uuid.UUID) error {
	return nil
}

func (s *Service) EnqueueJob(params *ffmpeg.EncodingParams) error {
	return nil
}

func (s *Service) GetTimedOutNodes(ts time.Time) []uuid.UUID {
	return nil
}
