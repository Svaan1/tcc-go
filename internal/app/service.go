package app

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/svaan1/go-tcc/internal/ffmpeg"
)

// Service is the job scheduling service that coordinates between job queue and node management.
type Service struct {
	nodeManager *NodeManager
	jobQueue    *JobQueue
}

func NewService() *Service {
	return &Service{
		nodeManager: NewNodeManager(),
		jobQueue:    NewJobQueue(),
	}
}

// Node management methods (delegated to NodeManager)
func (s *Service) RegisterNode(name string, codecs []string, now time.Time) (*Node, error) {
	return s.nodeManager.RegisterNode(name, codecs, now)
}

func (s *Service) RemoveNode(id uuid.UUID) {
	s.nodeManager.RemoveNode(id)
}

func (s *Service) ListNodes() []*Node {
	return s.nodeManager.ListNodes()
}

func (s *Service) UpdateResourceUsage(id uuid.UUID, usage ResourceUsage, ts time.Time) error {
	return s.nodeManager.UpdateResourceUsage(id, usage, ts)
}

func (s *Service) GetTimedOutNodes(now time.Time, timeout time.Duration) []uuid.UUID {
	return s.nodeManager.GetTimedOutNodes(now, timeout)
}

// Job management methods (delegated to JobQueue)
func (s *Service) EnqueueJob(params *ffmpeg.EncodingParams) (*Job, error) {
	return s.jobQueue.Enqueue(params)
}

func (s *Service) GetJob(jobID uuid.UUID) (*Job, bool) {
	return s.jobQueue.GetJob(jobID)
}

func (s *Service) ListJobs() []*Job {
	return s.jobQueue.ListJobs()
}

// Scheduling logic - this is where the intelligence happens
func (s *Service) AssignJobToNode(jobID uuid.UUID) (*Node, error) {
	job, exists := s.jobQueue.GetJob(jobID)
	if !exists {
		return nil, errors.New("job not found")
	}

	if job.Status != JobStatusPending {
		return nil, errors.New("job is not in pending status")
	}

	// Get the required codec from job params
	requiredCodec := job.Params.VideoCodec
	availableNodes := s.nodeManager.GetAvailableNodes(requiredCodec)

	if len(availableNodes) == 0 {
		return nil, errors.New("no available nodes support the required codec")
	}

	// Select best node (placeholder for more sophisticated logic)
	bestNode := s.selectBestNode(availableNodes)

	// Update job status and assign node
	err := s.jobQueue.UpdateJobStatus(jobID, JobStatusAssigned)
	if err != nil {
		return nil, err
	}

	job.AssignedNode = &bestNode.ID

	return bestNode, nil
}

// selectBestNode implements node selection algorithm
func (s *Service) selectBestNode(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}

	// Simple algorithm: select node with lowest CPU usage
	// TODO: Implement more sophisticated scoring based on multiple factors
	bestNode := nodes[0]
	lowestCPU := bestNode.ResourceUsage.CPUPercent

	for _, node := range nodes[1:] {
		if node.ResourceUsage.CPUPercent < lowestCPU {
			lowestCPU = node.ResourceUsage.CPUPercent
			bestNode = node
		}
	}

	return bestNode
}

// ProcessPendingJobs attempts to assign pending jobs to available nodes
func (s *Service) ProcessPendingJobs() (int, error) {
	assigned := 0

	for {
		pendingJob := s.jobQueue.GetNextPendingJob()
		if pendingJob == nil {
			break // No more pending jobs
		}

		_, err := s.AssignJobToNode(pendingJob.ID)
		if err != nil {
			// Log error but continue with other jobs
			continue
		}

		assigned++
	}

	return assigned, nil
}

// UpdateJobStatus updates a job's status
func (s *Service) UpdateJobStatus(jobID uuid.UUID, status JobStatus) error {
	return s.jobQueue.UpdateJobStatus(jobID, status)
}
