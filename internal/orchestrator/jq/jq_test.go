package jq

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
)

func runJobQueueTests(t *testing.T, newQueue func() JobQueue) {
	t.Run("Enqueue", func(t *testing.T) {
		testEnqueue(t, newQueue())
	})

	t.Run("Dequeue", func(t *testing.T) {
		testDequeue(t, newQueue())
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		testUpdateStatus(t, newQueue())
	})

	t.Run("GetJob", func(t *testing.T) {
		testGetJob(t, newQueue())
	})

	t.Run("ListJobs", func(t *testing.T) {
		testListJobs(t, newQueue())
	})

	t.Run("Peek", func(t *testing.T) {
		testPeek(t, newQueue())
	})

	t.Run("GetQueueDepth", func(t *testing.T) {
		testGetQueueDepth(t, newQueue())
	})

	t.Run("EmptyQueue", func(t *testing.T) {
		testEmptyQueue(t, newQueue())
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		testConcurrentOperations(t, newQueue())
	})
}

func testEnqueue(t *testing.T, q JobQueue) {
	job := createTestJob()

	err := q.Enqueue(job)
	assert.NoError(t, err)

	// Verify job was added
	depth, err := q.GetQueueDepth()
	require.NoError(t, err)
	assert.Equal(t, 1, depth)

	// Verify we can retrieve the job
	retrievedJob, err := q.GetJob(job.ID)
	require.NoError(t, err)
	assert.Equal(t, job.ID, retrievedJob.ID)
	assert.Equal(t, JobStatusPending, retrievedJob.Status)
}

func testDequeue(t *testing.T, q JobQueue) {
	// Test dequeue from empty queue
	job, err := q.Dequeue()
	assert.Error(t, err)
	assert.Nil(t, job)

	// Add jobs and test FIFO behavior
	job1 := createTestJob()
	job2 := createTestJob()

	require.NoError(t, q.Enqueue(job1))
	require.NoError(t, q.Enqueue(job2))

	// Dequeue should return first job
	dequeuedJob, err := q.Dequeue()
	require.NoError(t, err)
	assert.Equal(t, job1.ID, dequeuedJob.ID)
	assert.Equal(t, JobStatusAssigned, dequeuedJob.Status)

	// Queue depth should decrease
	depth, err := q.GetQueueDepth()
	require.NoError(t, err)
	assert.Equal(t, 1, depth)

	// Second dequeue should return second job
	dequeuedJob, err = q.Dequeue()
	require.NoError(t, err)
	assert.Equal(t, job2.ID, dequeuedJob.ID)
}

func testUpdateStatus(t *testing.T, q JobQueue) {
	job := createTestJob()
	require.NoError(t, q.Enqueue(job))

	// Update to running
	err := q.UpdateStatus(job.ID, JobStatusRunning)
	assert.NoError(t, err)

	// Verify status was updated
	updatedJob, err := q.GetJob(job.ID)
	require.NoError(t, err)
	assert.Equal(t, JobStatusRunning, updatedJob.Status)
	assert.True(t, updatedJob.UpdatedAt.After(job.UpdatedAt))

	// Test updating non-existent job
	nonExistentID := uuid.New()
	err = q.UpdateStatus(nonExistentID, JobStatusCompleted)
	assert.Error(t, err)
}

func testGetJob(t *testing.T, q JobQueue) {
	job := createTestJob()
	require.NoError(t, q.Enqueue(job))

	// Test getting existing job
	retrievedJob, err := q.GetJob(job.ID)
	require.NoError(t, err)
	assert.Equal(t, job.ID, retrievedJob.ID)
	assert.Equal(t, job.Params, retrievedJob.Params)

	// Test getting non-existent job
	nonExistentID := uuid.New()
	_, err = q.GetJob(nonExistentID)
	assert.Error(t, err)
}

func testListJobs(t *testing.T, q JobQueue) {
	// Test empty list
	jobs := q.ListJobs()
	assert.Empty(t, jobs)

	// Add multiple jobs
	job1 := createTestJob()
	job2 := createTestJob()
	job3 := createTestJob()

	require.NoError(t, q.Enqueue(job1))
	require.NoError(t, q.Enqueue(job2))
	require.NoError(t, q.Enqueue(job3))

	// Update one job status
	require.NoError(t, q.UpdateStatus(job2.ID, JobStatusRunning))

	// List all jobs
	jobs = q.ListJobs()
	assert.Len(t, jobs, 3)

	// Verify job IDs are present
	jobIDs := make(map[uuid.UUID]bool)
	for _, job := range jobs {
		jobIDs[job.ID] = true
	}
	assert.True(t, jobIDs[job1.ID])
	assert.True(t, jobIDs[job2.ID])
	assert.True(t, jobIDs[job3.ID])
}

func testPeek(t *testing.T, q JobQueue) {
	// Test peek on empty queue
	job, err := q.Peek()
	assert.Error(t, err)
	assert.Nil(t, job)

	// Add jobs
	job1 := createTestJob()
	job2 := createTestJob()

	require.NoError(t, q.Enqueue(job1))
	require.NoError(t, q.Enqueue(job2))

	// Peek should return first job without removing it
	peekedJob, err := q.Peek()
	require.NoError(t, err)
	assert.Equal(t, job1.ID, peekedJob.ID)
	assert.Equal(t, JobStatusPending, peekedJob.Status)

	// Queue depth should remain unchanged
	depth, err := q.GetQueueDepth()
	require.NoError(t, err)
	assert.Equal(t, 2, depth)

	// Multiple peeks should return the same job
	peekedJob2, err := q.Peek()
	require.NoError(t, err)
	assert.Equal(t, peekedJob.ID, peekedJob2.ID)
}

func testGetQueueDepth(t *testing.T, q JobQueue) {
	// Test empty queue depth
	depth, err := q.GetQueueDepth()
	require.NoError(t, err)
	assert.Equal(t, 0, depth)

	// Add jobs and verify depth increases
	job1 := createTestJob()
	job2 := createTestJob()

	require.NoError(t, q.Enqueue(job1))
	depth, err = q.GetQueueDepth()
	require.NoError(t, err)
	assert.Equal(t, 1, depth)

	require.NoError(t, q.Enqueue(job2))
	depth, err = q.GetQueueDepth()
	require.NoError(t, err)
	assert.Equal(t, 2, depth)

	// Dequeue and verify depth decreases
	_, err = q.Dequeue()
	require.NoError(t, err)
	depth, err = q.GetQueueDepth()
	require.NoError(t, err)
	assert.Equal(t, 1, depth)
}

func testEmptyQueue(t *testing.T, q JobQueue) {
	// Test all operations on empty queue
	depth, err := q.GetQueueDepth()
	require.NoError(t, err)
	assert.Equal(t, 0, depth)

	job, err := q.Dequeue()
	assert.Error(t, err)
	assert.Nil(t, job)

	job, err = q.Peek()
	assert.Error(t, err)
	assert.Nil(t, job)

	jobs := q.ListJobs()
	assert.Empty(t, jobs)
}

func testConcurrentOperations(t *testing.T, q JobQueue) {
	const numGoroutines = 10
	const jobsPerGoroutine = 5

	// Concurrently enqueue jobs
	jobChan := make(chan *Job, numGoroutines*jobsPerGoroutine)
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < jobsPerGoroutine; j++ {
				job := createTestJob()
				if err := q.Enqueue(job); err != nil {
					errChan <- err
					return
				}
				jobChan <- job
			}
			errChan <- nil
		}()
	}

	// Wait for all enqueue operations
	var enqueuedJobs []*Job
	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		require.NoError(t, err)
	}

	close(jobChan)
	for job := range jobChan {
		enqueuedJobs = append(enqueuedJobs, job)
	}

	// Verify final queue depth
	depth, err := q.GetQueueDepth()
	require.NoError(t, err)
	assert.Equal(t, len(enqueuedJobs), depth)

	// Concurrently dequeue all jobs
	dequeuedJobs := make(chan *Job, len(enqueuedJobs))
	dequeueErrChan := make(chan error, len(enqueuedJobs))

	for i := 0; i < len(enqueuedJobs); i++ {
		go func() {
			job, err := q.Dequeue()
			if err != nil {
				dequeueErrChan <- err
				return
			}
			dequeuedJobs <- job
			dequeueErrChan <- nil
		}()
	}

	// Wait for all dequeue operations
	var dequeuedJobList []*Job
	for i := 0; i < len(enqueuedJobs); i++ {
		err := <-dequeueErrChan
		require.NoError(t, err)
	}

	close(dequeuedJobs)
	for job := range dequeuedJobs {
		dequeuedJobList = append(dequeuedJobList, job)
	}

	// Verify all jobs were dequeued
	assert.Len(t, dequeuedJobList, len(enqueuedJobs))

	// Verify queue is empty
	depth, err = q.GetQueueDepth()
	require.NoError(t, err)
	assert.Equal(t, 0, depth)
}

// Helper function to create a test job
func createTestJob() *Job {
	return &Job{
		ID:        uuid.New(),
		Status:    JobStatusPending,
		Params:    &ffmpeg.EncodingParams{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Example of how to test a database implementation
func TestDatabaseJobQueue(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping database tests in short mode")
	}

	runJobQueueTests(t, func() JobQueue {
		// Setup test database
		// db := setupTestDB()
		// return NewDatabaseJobQueue(db)
		t.Skip("Database implementation not yet available")
		return nil
	})
}
