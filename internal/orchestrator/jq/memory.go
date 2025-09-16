package jq

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrQueueEmpty  = errors.New("queue is empty")
	ErrJobNotFound = errors.New("job not found")
)

type InMemoryJobQueue struct {
	mu    sync.RWMutex
	jobs  map[uuid.UUID]*Job
	queue []*Job
}

func NewInMemoryJobQueue() *InMemoryJobQueue {
	return &InMemoryJobQueue{
		jobs:  make(map[uuid.UUID]*Job),
		queue: make([]*Job, 0),
	}
}

func (q *InMemoryJobQueue) Enqueue(ctx context.Context, job *Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	q.jobs[job.ID] = job
	q.queue = append(q.queue, job)

	return nil
}

func (q *InMemoryJobQueue) Dequeue(ctx context.Context) (*Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil, ErrQueueEmpty
	}

	job := q.queue[0]
	q.queue = q.queue[1:]

	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	return job, nil
}

func (q *InMemoryJobQueue) Peek(ctx context.Context) (*Job, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.queue) == 0 {
		return nil, ErrQueueEmpty
	}

	return q.queue[0], nil
}

func (q *InMemoryJobQueue) ListJobs(ctx context.Context) ([]*Job, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	jobs := make([]*Job, 0, len(q.jobs))
	for _, job := range q.jobs {
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (q *InMemoryJobQueue) GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	job, exists := q.jobs[jobID]
	if !exists {
		return nil, ErrJobNotFound
	}

	return job, nil
}

func (q *InMemoryJobQueue) GetQueueDepth(ctx context.Context) (int, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return len(q.queue), nil
}
