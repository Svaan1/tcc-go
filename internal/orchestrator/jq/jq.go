package jq

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type JobParams struct {
	InputPath  string
	OutputPath string
	VideoCodec string
}

type Job struct {
	ID        uuid.UUID `json:"id"`
	Params    JobParams `json:"params"`
	Priority  int
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type JobQueue interface {
	Enqueue(ctx context.Context, params JobParams) (uuid.UUID, error)
	Requeue(ctx context.Context, job *Job) error
	Dequeue(ctx context.Context) (*Job, error)
	Peek(ctx context.Context) (*Job, error)

	ListJobs(ctx context.Context) ([]*Job, error)
	GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error)
	GetQueueDepth(ctx context.Context) (int, error)
}
