package jq

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
)

type JobStatus int

const (
	JobStatusPending JobStatus = iota
	JobStatusAssigned
	JobStatusRunning
	JobStatusCompleted
	JobStatusFailed
)

type Job struct {
	ID            uuid.UUID              `json:"id"`
	Status        JobStatus              `json:"status"`
	Params        *ffmpeg.EncodingParams `json:"params"`
	FailureReason string                 `json:"failure_reason,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type JobQueue interface {
	Enqueue(ctx context.Context, job *Job) error
	Dequeue(ctx context.Context) (*Job, error)
	Peek(ctx context.Context) (*Job, error)

	ListJobs(ctx context.Context, offset, limit int) ([]*Job, error)
	GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error)
	UpdateStatus(ctx context.Context, jobID uuid.UUID, status JobStatus, reason string) error
	GetQueueDepth(ctx context.Context) (int, error)
}
