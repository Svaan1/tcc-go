package jq

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type JobStatus int

const (
	JobStatusPending JobStatus = iota
	JobStatusAssigned
	JobStatusRunning
	JobStatusCompleted
	JobStatusFailed
)

type JobProgress struct {
	JobID  uuid.UUID `json:"job_id"`
	NodeID uuid.UUID `json:"node_id"`
	Status JobStatus `json:"status"`

	StartedAt   *time.Time `json:"started_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	ErrorMessage string `json:"error_message,omitempty"`
}

type JobHistory struct {
	JobID  uuid.UUID `json:"job_id"`
	NodeID uuid.UUID `json:"node_id"`
	Status JobStatus `json:"status"`

	StartedAt   time.Time     `json:"started_at"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	Duration    time.Duration `json:"duration"`
	RetryCount  int           `json:"retry_count"`

	ErrorMessage string `json:"error_message,omitempty"`
}

type JobTracker interface {
	TrackJob(ctx context.Context, jobID uuid.UUID, nodeID uuid.UUID) error
	UpdateJobProgress(ctx context.Context, jobID uuid.UUID, progress *JobProgress) error
	GetJobProgress(ctx context.Context, jobID uuid.UUID) (*JobProgress, error)
	CompleteJobTracking(ctx context.Context, jobID uuid.UUID, success bool, errorMsg string) error

	GetActiveJobs(ctx context.Context) ([]*JobProgress, error)
	GetJobsByNode(ctx context.Context, nodeID uuid.UUID) ([]*JobProgress, error)
	GetJobsByStatus(ctx context.Context, status JobStatus) ([]*JobProgress, error)

	CleanupCompletedJobs(ctx context.Context, olderThan time.Duration) error
	GetStaleJobs(ctx context.Context, timeout time.Duration) ([]*JobProgress, error)
}
