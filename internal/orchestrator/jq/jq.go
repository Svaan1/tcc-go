package jq

import (
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
	ID     uuid.UUID              `json:"id"`
	Status JobStatus              `json:"status"`
	Params *ffmpeg.EncodingParams `json:"params"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type JobQueue interface {
	Enqueue(job *Job) error
	Dequeue() (*Job, error)
	UpdateStatus(jobID uuid.UUID, status JobStatus) error

	GetJob(jobID uuid.UUID) (*Job, error)
	ListJobs() []*Job
	Peek() (*Job, error)
	GetQueueDepth() (int, error)
}
