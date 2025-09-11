package jq

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
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

func (jq *InMemoryJobQueue) Enqueue(job *Job) error {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	jq.jobs = append(jq.jobs, job)
	return nil
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

func (jq *InMemoryJobQueue) GetJob(jobID uuid.UUID) (*Job, error) {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	for _, job := range jq.jobs {
		if job.ID == jobID {
			return job, nil
		}
	}

	return nil, fmt.Errorf("job not found")
}

func (jq *InMemoryJobQueue) ListJobs() []*Job {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	// Return a copy to avoid external modification
	jobs := make([]*Job, len(jq.jobs))
	copy(jobs, jq.jobs)
	return jobs
}
