package jq

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
)

func TestInMemoryJobQueue_Enqueue(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	job := &Job{
		ID:     uuid.New(),
		Params: &ffmpeg.EncodingParams{},
	}

	err := queue.Enqueue(ctx, job)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if job.Status != JobStatusPending {
		t.Errorf("expected status %v, got %v", JobStatusPending, job.Status)
	}

	if job.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if job.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}

	depth, err := queue.GetQueueDepth(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if depth != 1 {
		t.Errorf("expected queue depth 1, got %d", depth)
	}
}

func TestInMemoryJobQueue_Dequeue(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Test empty queue
	_, err := queue.Dequeue(ctx)
	if err != ErrQueueEmpty {
		t.Errorf("expected ErrQueueEmpty, got %v", err)
	}

	// Add job and dequeue
	job := &Job{
		ID:     uuid.New(),
		Params: &ffmpeg.EncodingParams{},
	}

	queue.Enqueue(ctx, job)

	dequeuedJob, err := queue.Dequeue(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if dequeuedJob.ID != job.ID {
		t.Errorf("expected job ID %v, got %v", job.ID, dequeuedJob.ID)
	}

	if dequeuedJob.Status != JobStatusAssigned {
		t.Errorf("expected status %v, got %v", JobStatusAssigned, dequeuedJob.Status)
	}

	depth, err := queue.GetQueueDepth(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if depth != 0 {
		t.Errorf("expected queue depth 0, got %d", depth)
	}
}

func TestInMemoryJobQueue_Peek(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Test empty queue
	_, err := queue.Peek(ctx)
	if err != ErrQueueEmpty {
		t.Errorf("expected ErrQueueEmpty, got %v", err)
	}

	// Add job and peek
	job := &Job{
		ID:     uuid.New(),
		Params: &ffmpeg.EncodingParams{},
	}

	queue.Enqueue(ctx, job)

	peekedJob, err := queue.Peek(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if peekedJob.ID != job.ID {
		t.Errorf("expected job ID %v, got %v", job.ID, peekedJob.ID)
	}

	// Queue should still have the job
	depth, err := queue.GetQueueDepth(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if depth != 1 {
		t.Errorf("expected queue depth 1, got %d", depth)
	}
}

func TestInMemoryJobQueue_GetJob(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	jobID := uuid.New()

	// Test non-existent job
	_, err := queue.GetJob(ctx, jobID)
	if err != ErrJobNotFound {
		t.Errorf("expected ErrJobNotFound, got %v", err)
	}

	// Add job and retrieve
	job := &Job{
		ID:     jobID,
		Params: &ffmpeg.EncodingParams{},
	}

	queue.Enqueue(ctx, job)

	retrievedJob, err := queue.GetJob(ctx, jobID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrievedJob.ID != jobID {
		t.Errorf("expected job ID %v, got %v", jobID, retrievedJob.ID)
	}
}

func TestInMemoryJobQueue_UpdateStatus(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	jobID := uuid.New()

	// Test non-existent job
	err := queue.UpdateStatus(ctx, jobID, JobStatusRunning)
	if err != ErrJobNotFound {
		t.Errorf("expected ErrJobNotFound, got %v", err)
	}

	// Add job and update status
	job := &Job{
		ID:     jobID,
		Params: &ffmpeg.EncodingParams{},
	}

	queue.Enqueue(ctx, job)
	originalUpdatedAt := job.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond)

	err = queue.UpdateStatus(ctx, jobID, JobStatusRunning)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updatedJob, _ := queue.GetJob(ctx, jobID)
	if updatedJob.Status != JobStatusRunning {
		t.Errorf("expected status %v, got %v", JobStatusRunning, updatedJob.Status)
	}

	if !updatedJob.UpdatedAt.After(originalUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestInMemoryJobQueue_ListJobs(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Test empty list
	jobs, err := queue.ListJobs(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(jobs))
	}

	// Add multiple jobs
	job1 := &Job{ID: uuid.New(), Params: &ffmpeg.EncodingParams{}}
	job2 := &Job{ID: uuid.New(), Params: &ffmpeg.EncodingParams{}}

	queue.Enqueue(ctx, job1)
	queue.Enqueue(ctx, job2)

	jobs, err = queue.ListJobs(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestInMemoryJobQueue_FIFO(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Add jobs in order
	job1 := &Job{ID: uuid.New(), Params: &ffmpeg.EncodingParams{}}
	job2 := &Job{ID: uuid.New(), Params: &ffmpeg.EncodingParams{}}
	job3 := &Job{ID: uuid.New(), Params: &ffmpeg.EncodingParams{}}

	queue.Enqueue(ctx, job1)
	queue.Enqueue(ctx, job2)
	queue.Enqueue(ctx, job3)

	// Dequeue should return jobs in FIFO order
	first, err := queue.Dequeue(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if first.ID != job1.ID {
		t.Errorf("expected first job ID %v, got %v", job1.ID, first.ID)
	}

	second, err := queue.Dequeue(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if second.ID != job2.ID {
		t.Errorf("expected second job ID %v, got %v", job2.ID, second.ID)
	}

	third, err := queue.Dequeue(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if third.ID != job3.ID {
		t.Errorf("expected third job ID %v, got %v", job3.ID, third.ID)
	}
}

func TestInMemoryJobQueue_GetQueueDepth(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Test initial depth
	depth, err := queue.GetQueueDepth(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if depth != 0 {
		t.Errorf("expected queue depth 0, got %d", depth)
	}

	// Add jobs
	job1 := &Job{ID: uuid.New(), Params: &ffmpeg.EncodingParams{}}
	job2 := &Job{ID: uuid.New(), Params: &ffmpeg.EncodingParams{}}

	queue.Enqueue(ctx, job1)
	depth, _ = queue.GetQueueDepth(ctx)
	if depth != 1 {
		t.Errorf("expected queue depth 1, got %d", depth)
	}

	queue.Enqueue(ctx, job2)
	depth, _ = queue.GetQueueDepth(ctx)
	if depth != 2 {
		t.Errorf("expected queue depth 2, got %d", depth)
	}

	// Dequeue one
	queue.Dequeue(ctx)
	depth, _ = queue.GetQueueDepth(ctx)
	if depth != 1 {
		t.Errorf("expected queue depth 1, got %d", depth)
	}
}

func TestInMemoryJobQueue_ConcurrentAccess(t *testing.T) {
	queue := NewInMemoryJobQueue()
	ctx := context.Background()

	// Test basic concurrent safety
	done := make(chan bool, 2)

	// Goroutine 1: Enqueue jobs
	go func() {
		for i := 0; i < 100; i++ {
			job := &Job{ID: uuid.New(), Params: &ffmpeg.EncodingParams{}}
			queue.Enqueue(ctx, job)
		}
		done <- true
	}()

	// Goroutine 2: Check queue depth
	go func() {
		for i := 0; i < 100; i++ {
			queue.GetQueueDepth(ctx)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	depth, err := queue.GetQueueDepth(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if depth != 100 {
		t.Errorf("expected queue depth 100, got %d", depth)
	}
}
