package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
	"github.com/svaan1/tcc-go/internal/metrics"
	"github.com/svaan1/tcc-go/internal/orchestrator/jq"
	"github.com/svaan1/tcc-go/internal/orchestrator/js"
	"github.com/svaan1/tcc-go/internal/orchestrator/jt"
	"github.com/svaan1/tcc-go/internal/orchestrator/np"
)

type Service struct {
	jq jq.JobQueue
	jt jt.JobTracker
	js js.JobScheduler
	np np.NodePool
}

func NewService() *Service {
	return &Service{
		jq: jq.NewInMemoryJobQueue(),
		jt: jt.NewInMemoryJobTracker(),
		js: js.NewRoundRobinScheduler(),
		np: np.NewInMemoryNodePool(),
	}
}

func (s *Service) ListNodes(ctx context.Context) ([]*np.Node, error) {
	return s.np.ListNodes(ctx, 0, 20)
}

func (s *Service) GetTimedOutNodes(ctx context.Context, timeout time.Duration) ([]*np.Node, error) {
	return s.np.GetTimedOutNodes(ctx, timeout)
}

func (s *Service) RegisterNode(ctx context.Context, name string, profiles []ffmpeg.EncodingProfile) (*np.Node, error) {
	return s.np.RegisterNode(ctx, &np.NodeRegistration{
		Name:     name,
		Profiles: profiles,
	})
}

func (s *Service) UnregisterNode(ctx context.Context, nodeID uuid.UUID) error {
	return s.np.UnregisterNode(ctx, nodeID)
}

func (s *Service) UpdateNodeMetrics(ctx context.Context, nodeID string, usage *metrics.ResourceUsage) error {
	parsedID, err := uuid.Parse(nodeID)
	if err != nil {
		return err
	}

	return s.np.UpdateNodeMetrics(ctx, parsedID, usage)
}

func (s *Service) EnqueueJob(ctx context.Context, params *ffmpeg.EncodingParams) error {
	job := &jq.Job{
		ID:     uuid.New(),
		Params: params,
	}

	if err := s.jq.Enqueue(ctx, job); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	nodes, err := s.np.GetAvailableNodes(ctx, &np.NodeFilter{Codec: params.VideoCodec})
	if err != nil {
		return fmt.Errorf("failed to get available nodes: %w", err)
	}

	node, err := s.js.SelectBestNode(job, nodes)
	if err != nil {
		return fmt.Errorf("failed to select best node: %w", err)
	}

	if err := s.jt.TrackJob(ctx, job.ID, node.ID); err != nil {
		return fmt.Errorf("failed to track job: %w", err)
	}

	return nil
}

func (s *Service) GetJob(ctx context.Context, jobID uuid.UUID) (*jq.Job, error) {
	return s.jq.GetJob(ctx, jobID)
}

func (s *Service) ListJobs(ctx context.Context) ([]*jq.Job, error) {
	return s.jq.ListJobs(ctx)
}

func (s *Service) GetJobProgress(ctx context.Context, jobID uuid.UUID) (*jt.JobProgress, error) {
	return s.jt.GetJobProgress(ctx, jobID)
}

func (s *Service) GetActiveJobs(ctx context.Context) ([]*jt.JobProgress, error) {
	return s.jt.GetActiveJobs(ctx)
}
