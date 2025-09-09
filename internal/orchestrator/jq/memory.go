package jq

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/svaan1/go-tcc/internal/ffmpeg"
)

type InMemoryJobQueue struct {
	JobQueue

	jobs []*Job
	mu   sync.RWMutex
}

func NewInMemoryJobQueue() *InMemoryJobQueue {
	return &InMemoryJobQueue{
		jobs: []*Job{},
	}
}

func (jq *InMemoryJobQueue) Enqueue(params *ffmpeg.EncodingParams) (*Job, error) {
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

func (jq *InMemoryJobQueue) GetNextPendingJob() *Job {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	for _, job := range jq.jobs {
		if job.Status == JobStatusPending {
			return job
		}
	}
	return nil
}

func (jq *InMemoryJobQueue) UpdateJobStatus(jobID uuid.UUID, status JobStatus) error {
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

func (jq *InMemoryJobQueue) GetJob(jobID uuid.UUID) (*Job, bool) {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	for _, job := range jq.jobs {
		if job.ID == jobID {
			return job, true
		}
	}
	return nil, false
}

func (jq *InMemoryJobQueue) ListJobs() []*Job {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	// Return a copy to avoid external modification
	jobs := make([]*Job, len(jq.jobs))
	copy(jobs, jq.jobs)
	return jobs
}
