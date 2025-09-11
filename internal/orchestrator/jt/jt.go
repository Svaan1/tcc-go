package jt

import (
	"github.com/google/uuid"
)

type JobProgress struct {
}

type JobTracker interface {
	StartTracking(jobID uuid.UUID, nodeID uuid.UUID) error
	UpdateProgress(jobID uuid.UUID, progress *JobProgress) error
	GetProgress(jobID uuid.UUID) (*JobProgress, error)

	// Completion handling
	// MarkCompleted(jobID uuid.UUID, result *JobResult) error
	MarkFailed(jobID uuid.UUID, err error) error

	// Failure and retry management
	HandleNodeFailure(nodeID uuid.UUID) error
	// GetJobsForRetry() ([]*Job, error)
	// ShouldRetry(job *Job) (bool, error)
}
