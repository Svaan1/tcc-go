package jq

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryJobQueue(t *testing.T) {
	queue := NewInMemoryJobQueue()

	assert.NotNil(t, queue)
	assert.NotNil(t, queue.jobs)
	assert.NotNil(t, queue.queue)
	assert.Equal(t, 0, len(queue.jobs))
	assert.Equal(t, 0, len(queue.queue))
}

func TestEnqueue(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()
	params := JobParams{InputPath: "test-job"}

	jobID, err := queue.Enqueue(ctx, params)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, jobID)

	// Verify job was added to jobs map
	job, exists := queue.jobs[jobID]
	assert.True(t, exists)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, params, job.Params)
	assert.Equal(t, DefaultPriority, job.Priority)
	assert.False(t, job.CreatedAt.IsZero())
	assert.False(t, job.UpdatedAt.IsZero())

	// Verify job was added to queue
	assert.Equal(t, 1, len(queue.queue))
	assert.Equal(t, job, queue.queue[0])
}

func TestEnqueueWithPriority(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()
	params := JobParams{InputPath: "priority-job"}
	priority := 5

	jobID, err := queue.EnqueueWithPriority(ctx, params, priority)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, jobID)

	job, exists := queue.jobs[jobID]
	assert.True(t, exists)
	assert.Equal(t, priority, job.Priority)
}

func TestEnqueueMultiplePriorities(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Add jobs with different priorities
	lowID, _ := queue.EnqueueWithPriority(ctx, JobParams{InputPath: "low"}, 1)
	highID, _ := queue.EnqueueWithPriority(ctx, JobParams{InputPath: "high"}, 10)
	medID, _ := queue.EnqueueWithPriority(ctx, JobParams{InputPath: "medium"}, 5)

	// Verify queue is ordered by priority (highest first)
	assert.Equal(t, 3, len(queue.queue))
	assert.Equal(t, highID, queue.queue[0].ID)
	assert.Equal(t, medID, queue.queue[1].ID)
	assert.Equal(t, lowID, queue.queue[2].ID)
}

func TestDequeue(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()
	params := JobParams{InputPath: "dequeue-test"}

	// Enqueue a job first
	jobID, _ := queue.Enqueue(ctx, params)
	originalJob := queue.jobs[jobID]
	originalUpdateTime := originalJob.UpdatedAt

	// Small delay to ensure UpdatedAt changes
	time.Sleep(time.Millisecond)

	// Dequeue the job
	dequeuedJob, err := queue.Dequeue(ctx)

	require.NoError(t, err)
	assert.Equal(t, jobID, dequeuedJob.ID)
	assert.Equal(t, params, dequeuedJob.Params)
	assert.True(t, dequeuedJob.UpdatedAt.After(originalUpdateTime))

	// Verify job is still in jobs map but removed from queue
	assert.Equal(t, 1, len(queue.jobs))
	assert.Equal(t, 0, len(queue.queue))
}

func TestDequeueEmptyQueue(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	job, err := queue.Dequeue(ctx)

	assert.Nil(t, job)
	assert.Equal(t, ErrQueueEmpty, err)
}

func TestDequeueOrder(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Enqueue jobs with different priorities
	queue.EnqueueWithPriority(ctx, JobParams{InputPath: "low"}, 1)
	queue.EnqueueWithPriority(ctx, JobParams{InputPath: "high"}, 10)
	queue.EnqueueWithPriority(ctx, JobParams{InputPath: "medium"}, 5)

	// Dequeue should return highest priority first
	job1, _ := queue.Dequeue(ctx)
	assert.Equal(t, JobParams{InputPath: "high"}, job1.Params)

	job2, _ := queue.Dequeue(ctx)
	assert.Equal(t, JobParams{InputPath: "medium"}, job2.Params)

	job3, _ := queue.Dequeue(ctx)
	assert.Equal(t, JobParams{InputPath: "low"}, job3.Params)
}

func TestRequeue(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()
	params := JobParams{InputPath: "requeue-test"}

	// Enqueue and dequeue a job
	jobID, _ := queue.Enqueue(ctx, params)
	job, _ := queue.Dequeue(ctx)
	originalPriority := job.Priority
	originalUpdateTime := job.UpdatedAt

	// Small delay to ensure UpdatedAt changes
	time.Sleep(time.Millisecond)

	// Requeue the job
	err := queue.Requeue(ctx, job)

	require.NoError(t, err)
	assert.Equal(t, originalPriority-1, job.Priority)
	assert.True(t, job.UpdatedAt.After(originalUpdateTime))

	// Verify job is back in queue
	assert.Equal(t, 1, len(queue.queue))
	assert.Equal(t, jobID, queue.queue[0].ID)
}

func TestRequeueNonExistentJob(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Create a job that doesn't exist in the queue
	fakeJob := &Job{
		ID:       uuid.New(),
		Params:   JobParams{InputPath: "fake"},
		Priority: 5,
	}

	err := queue.Requeue(ctx, fakeJob)

	assert.Equal(t, ErrJobNotFound, err)
}

func TestPeek(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()
	params := JobParams{InputPath: "peek-test"}

	// Enqueue a job
	jobID, _ := queue.Enqueue(ctx, params)

	// Peek should return the job without removing it
	peekedJob, err := queue.Peek(ctx)

	require.NoError(t, err)
	assert.Equal(t, jobID, peekedJob.ID)
	assert.Equal(t, params, peekedJob.Params)

	// Verify job is still in queue
	assert.Equal(t, 1, len(queue.queue))
}

func TestPeekEmptyQueue(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	job, err := queue.Peek(ctx)

	assert.Nil(t, job)
	assert.Equal(t, ErrQueueEmpty, err)
}

func TestListJobs(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Enqueue multiple jobs
	id1, _ := queue.Enqueue(ctx, JobParams{InputPath: "job1"})
	id2, _ := queue.Enqueue(ctx, JobParams{InputPath: "job2"})
	id3, _ := queue.Enqueue(ctx, JobParams{InputPath: "job3"})

	// List all jobs
	jobs, err := queue.ListJobs(ctx)

	require.NoError(t, err)
	assert.Equal(t, 3, len(jobs))

	// Verify all job IDs are present
	jobIDs := make(map[uuid.UUID]bool)
	for _, job := range jobs {
		jobIDs[job.ID] = true
	}
	assert.True(t, jobIDs[id1])
	assert.True(t, jobIDs[id2])
	assert.True(t, jobIDs[id3])
}

func TestListJobsEmpty(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	jobs, err := queue.ListJobs(ctx)

	require.NoError(t, err)
	assert.Equal(t, 0, len(jobs))
}

func TestGetJob(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()
	params := JobParams{InputPath: "get-job-test"}

	// Enqueue a job
	jobID, _ := queue.Enqueue(ctx, params)

	// Get the job by ID
	job, err := queue.GetJob(ctx, jobID)

	require.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, params, job.Params)
}

func TestGetJobNotFound(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Try to get a non-existent job
	fakeID := uuid.New()
	job, err := queue.GetJob(ctx, fakeID)

	assert.Nil(t, job)
	assert.Equal(t, ErrJobNotFound, err)
}

func TestGetQueueDepth(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Initially empty
	depth, err := queue.GetQueueDepth(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, depth)

	// Add jobs
	queue.Enqueue(ctx, JobParams{InputPath: "job1"})
	queue.Enqueue(ctx, JobParams{InputPath: "job2"})
	queue.Enqueue(ctx, JobParams{InputPath: "job3"})

	depth, err = queue.GetQueueDepth(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, depth)

	// Dequeue one job
	queue.Dequeue(ctx)

	depth, err = queue.GetQueueDepth(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, depth)
}

func TestConcurrency(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Test concurrent enqueues
	const numGoroutines = 10
	const jobsPerGoroutine = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < jobsPerGoroutine; j++ {
				queue.Enqueue(ctx, JobParams{
					InputPath: "concurrent-job",
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all jobs were enqueued
	depth, _ := queue.GetQueueDepth(ctx)
	assert.Equal(t, numGoroutines*jobsPerGoroutine, depth)

	jobs, _ := queue.ListJobs(ctx)
	assert.Equal(t, numGoroutines*jobsPerGoroutine, len(jobs))
}

func TestInsertByPriorityEdgeCases(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Test with same priorities
	queue.EnqueueWithPriority(ctx, JobParams{InputPath: "first"}, 5)
	queue.EnqueueWithPriority(ctx, JobParams{InputPath: "second"}, 5)
	queue.EnqueueWithPriority(ctx, JobParams{InputPath: "third"}, 5)

	// All should have same priority, order should be maintained (FIFO for same priority)
	assert.Equal(t, 3, len(queue.queue))
	assert.Equal(t, JobParams{InputPath: "first"}, queue.queue[0].Params)
	assert.Equal(t, JobParams{InputPath: "second"}, queue.queue[1].Params)
	assert.Equal(t, JobParams{InputPath: "third"}, queue.queue[2].Params)
}

// Benchmark tests
func BenchmarkEnqueue(b *testing.B) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()
	params := JobParams{InputPath: "benchmark-job"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Enqueue(ctx, params)
	}
}

func BenchmarkDequeue(b *testing.B) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()
	params := JobParams{InputPath: "benchmark-job"}

	// Pre-populate queue
	for i := 0; i < b.N; i++ {
		queue.Enqueue(ctx, params)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Dequeue(ctx)
	}
}
