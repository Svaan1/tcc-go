package orchestrator

import (
	"context"
	"fmt"
	"sync"
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

	mu sync.Mutex
}

func NewService() *Service {
	return &Service{
		jq: jq.NewInMemoryJobQueue(),
		jt: jt.NewInMemoryJobTracker(),
		js: js.NewRoundRobinScheduler(),
		np: np.NewInMemoryNodePool(),
	}
}

// Nodes
func (s *Service) ListNodes(ctx context.Context) ([]*np.Node, error) {
	return s.np.ListNodes(ctx, 0, 20)
}

func (s *Service) RegisterNode(ctx context.Context, name string, profiles []ffmpeg.EncodingProfile) (uuid.UUID, error) {
	return s.np.RegisterNode(ctx, &np.NodeRegistration{
		Name:     name,
		Profiles: profiles,
	})
}

func (s *Service) UnregisterNode(ctx context.Context, nodeID uuid.UUID) error {
	return s.np.UnregisterNode(ctx, nodeID)
}

func (s *Service) GetTimedOutNodes(ctx context.Context, timeout time.Duration) ([]*np.Node, error) {
	return s.np.GetTimedOutNodes(ctx, timeout)
}

func (s *Service) UpdateNodeMetrics(ctx context.Context, nodeID string, usage *metrics.ResourceUsage) error {
	parsedID, err := uuid.Parse(nodeID)
	if err != nil {
		return err
	}

	return s.np.UpdateNodeMetrics(ctx, parsedID, usage)
}

// Queue
func (s *Service) EnqueueJob(ctx context.Context, params jq.JobParams) (uuid.UUID, error) {
	jobID, err := s.jq.Enqueue(ctx, params)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to enqueue job: %w", err)
	}

	return jobID, nil
}

func (s *Service) DequeueJob(ctx context.Context) (*jq.Job, *np.Node, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, err := s.jq.Dequeue(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to peek job: %w", err)
	}

	nodes, err := s.np.GetAvailableNodes(ctx, &np.NodeFilter{ProfileName: job.Params.ProfileName})
	if err != nil {
		s.jq.Requeue(ctx, job)
		return nil, nil, fmt.Errorf("failed to get available nodes: %w", err)
	}

	activeJobs, err := s.jt.GetActiveJobs(ctx)
	if err != nil {
		s.jq.Requeue(ctx, job)
		return nil, nil, fmt.Errorf("failed to get current active jobs %v", err)
	}

	node, err := s.js.SelectBestNode(job, nodes, activeJobs)
	if err != nil {
		s.jq.Requeue(ctx, job)
		return nil, nil, fmt.Errorf("failed to select best node: %w", err)
	}

	if err := s.jt.TrackJob(ctx, job.ID, node.ID); err != nil {
		s.jq.Requeue(ctx, job)
		return nil, nil, fmt.Errorf("failed to track job: %w", err)
	}

	return job, node, nil
}

func (s *Service) CompleteJob(ctx context.Context, jobID string, success bool, errorMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	parsedID, err := uuid.Parse(jobID)
	if err != nil {
		return err
	}

	return s.jt.CompleteJobTracking(ctx, parsedID, success, errorMsg)
}

func (s *Service) RejectJob(ctx context.Context, jobID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	parsedID, err := uuid.Parse(jobID)
	if err != nil {
		return err
	}

	// Remove the job from tracking
	if err := s.jt.CompleteJobTracking(ctx, parsedID, false, fmt.Sprintf("Job rejected: %s", reason)); err != nil {
		return fmt.Errorf("failed to complete job tracking: %w", err)
	}

	// Re-queue the job so it can be assigned to another node
	job := &jq.Job{
		ID: parsedID,
	}
	if err := s.jq.Requeue(ctx, job); err != nil {
		return fmt.Errorf("failed to requeue job: %w", err)
	}

	return nil
}

func (s *Service) GetQueueInfo(ctx context.Context) ([]*jq.Job, map[uuid.UUID][]*jt.JobProgress, error) {
	// Get pending jobs from queue
	pendingJobs, err := s.jq.ListJobs(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}

	// Get active (processing) jobs from tracker
	activeJobs, err := s.jt.GetActiveJobs(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get active jobs: %w", err)
	}

	return pendingJobs, activeJobs, nil
}

func (s *Service) GetJobHistory(ctx context.Context, statusFilter string, limit int) ([]*jt.JobHistory, error) {
	var jobs []*jt.JobHistory
	var err error

	if statusFilter == "" {
		jobs, err = s.jt.GetCompletedJobs(ctx)
	} else {
		var status jt.JobStatus
		switch statusFilter {
		case "completed":
			status = jt.JobStatusCompleted
		case "failed":
			status = jt.JobStatusFailed
		default:
			return nil, fmt.Errorf("invalid status filter: %s", statusFilter)
		}
		jobs, err = s.jt.GetCompletedJobsByStatus(ctx, status)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get job history: %w", err)
	}

	// Apply limit if specified
	if limit > 0 && len(jobs) > limit {
		jobs = jobs[:limit]
	}

	return jobs, nil
}

func (s *Service) ClearQueue(ctx context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count, err := s.jq.ClearQueue(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to clear queue: %w", err)
	}

	return count, nil
}
