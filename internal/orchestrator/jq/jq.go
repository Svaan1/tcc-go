package jq

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
)

type Job struct {
	ID        uuid.UUID              `json:"id"`
	Params    *ffmpeg.EncodingParams `json:"params"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type JobQueue interface {
	Enqueue(ctx context.Context, job *Job) error
	Dequeue(ctx context.Context) (*Job, error)
	Peek(ctx context.Context) (*Job, error)

	ListJobs(ctx context.Context) ([]*Job, error)
	GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error)
	GetQueueDepth(ctx context.Context) (int, error)
}
