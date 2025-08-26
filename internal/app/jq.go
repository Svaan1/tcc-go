package app

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/svaan1/go-tcc/internal/ffmpeg"
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
	ID           uuid.UUID
	Params       *ffmpeg.EncodingParams
	Status       JobStatus
	AssignedNode *uuid.UUID // nil if not assigned
}

type JobQueue struct {
	jobs []*Job
	mu   sync.RWMutex
}

func NewJobQueue() *JobQueue {
	return &JobQueue{
		jobs: []*Job{},
	}
}

func (jq *JobQueue) Enqueue(params *ffmpeg.EncodingParams) (*Job, error) {
	job := &Job{
		ID:     uuid.New(),
		Params: params,
		Status: JobStatusPending,
	}

	jq.mu.Lock()
	defer jq.mu.Unlock()

	jq.jobs = append(jq.jobs, job)
	return job, nil
}

func (jq *JobQueue) GetNextPendingJob() *Job {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	for _, job := range jq.jobs {
		if job.Status == JobStatusPending {
			return job
		}
	}
	return nil
}

func (jq *JobQueue) UpdateJobStatus(jobID uuid.UUID, status JobStatus) error {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	for _, job := range jq.jobs {
		if job.ID == jobID {
			job.Status = status
			return nil
		}
	}
	return errors.New("job not found")
}

func (jq *JobQueue) GetJob(jobID uuid.UUID) (*Job, bool) {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	for _, job := range jq.jobs {
		if job.ID == jobID {
			return job, true
		}
	}
	return nil, false
}

func (jq *JobQueue) ListJobs() []*Job {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	// Return a copy to avoid external modification
	jobs := make([]*Job, len(jq.jobs))
	copy(jobs, jq.jobs)
	return jobs
}
