package jt

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInMemoryJobTracker_TrackJob(t *testing.T) {
	tracker := NewInMemoryJobTracker()
	ctx := context.Background()
	jobID := uuid.New()
	nodeID := uuid.New()

	t.Run("successful tracking", func(t *testing.T) {
		err := tracker.TrackJob(ctx, jobID, nodeID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		progress, err := tracker.GetJobProgress(ctx, jobID)
		if err != nil {
			t.Fatalf("expected no error getting progress, got %v", err)
		}

		if progress.JobID != jobID {
			t.Errorf("expected job ID %v, got %v", jobID, progress.JobID)
		}
		if progress.NodeID != nodeID {
			t.Errorf("expected node ID %v, got %v", nodeID, progress.NodeID)
		}
		if progress.Status != JobStatusAssigned {
			t.Errorf("expected status %v, got %v", JobStatusAssigned, progress.Status)
		}
		if progress.StartedAt == nil {
			t.Error("expected StartedAt to be set")
		}
	})

	t.Run("duplicate tracking", func(t *testing.T) {
		err := tracker.TrackJob(ctx, jobID, nodeID)
		if err != ErrJobAlreadyTracked {
			t.Errorf("expected ErrJobAlreadyTracked, got %v", err)
		}
	})
}

func TestInMemoryJobTracker_UpdateJobProgress(t *testing.T) {
	tracker := NewInMemoryJobTracker()
	ctx := context.Background()
	jobID := uuid.New()
	nodeID := uuid.New()

	t.Run("update non-existent job", func(t *testing.T) {
		progress := &JobProgress{Status: JobStatusRunning}
		err := tracker.UpdateJobProgress(ctx, jobID, progress)
		if err != ErrJobProgressNotFound {
			t.Errorf("expected ErrJobProgressNotFound, got %v", err)
		}
	})

	t.Run("successful update", func(t *testing.T) {
		tracker.TrackJob(ctx, jobID, nodeID)

		progress := &JobProgress{
			Status:       JobStatusRunning,
			ErrorMessage: "test error",
		}

		err := tracker.UpdateJobProgress(ctx, jobID, progress)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		updatedProgress, err := tracker.GetJobProgress(ctx, jobID)
		if err != nil {
			t.Fatalf("expected no error getting progress, got %v", err)
		}

		if updatedProgress.Status != JobStatusRunning {
			t.Errorf("expected status %v, got %v", JobStatusRunning, updatedProgress.Status)
		}
		if updatedProgress.ErrorMessage != "test error" {
			t.Errorf("expected error message 'test error', got %v", updatedProgress.ErrorMessage)
		}
	})
}

func TestInMemoryJobTracker_CompleteJobTracking(t *testing.T) {
	tracker := NewInMemoryJobTracker()
	ctx := context.Background()
	jobID := uuid.New()
	nodeID := uuid.New()

	t.Run("complete non-existent job", func(t *testing.T) {
		err := tracker.CompleteJobTracking(ctx, jobID, true, "")
		if err != ErrJobProgressNotFound {
			t.Errorf("expected ErrJobProgressNotFound, got %v", err)
		}
	})

	t.Run("successful completion", func(t *testing.T) {
		tracker.TrackJob(ctx, jobID, nodeID)
		time.Sleep(time.Millisecond) // Ensure duration > 0

		err := tracker.CompleteJobTracking(ctx, jobID, true, "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Job should no longer be active
		_, err = tracker.GetJobProgress(ctx, jobID)
		if err != ErrJobProgressNotFound {
			t.Errorf("expected ErrJobProgressNotFound for completed job, got %v", err)
		}

		// Job should be in completed jobs
		if _, exists := tracker.completedJobs[jobID]; !exists {
			t.Error("expected job to be in completed jobs")
		}
	})

	t.Run("failed completion", func(t *testing.T) {
		failedJobID := uuid.New()
		tracker.TrackJob(ctx, failedJobID, nodeID)

		err := tracker.CompleteJobTracking(ctx, failedJobID, false, "encoding failed")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		history := tracker.completedJobs[failedJobID]
		if history.Status != JobStatusFailed {
			t.Errorf("expected status %v, got %v", JobStatusFailed, history.Status)
		}
		if history.ErrorMessage != "encoding failed" {
			t.Errorf("expected error message 'encoding failed', got %v", history.ErrorMessage)
		}
	})
}

func TestInMemoryJobTracker_GetActiveJobs(t *testing.T) {
	tracker := NewInMemoryJobTracker()
	ctx := context.Background()

	t.Run("empty tracker", func(t *testing.T) {
		jobs, err := tracker.GetActiveJobs(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(jobs) != 0 {
			t.Errorf("expected 0 jobs, got %d", len(jobs))
		}
	})

	t.Run("multiple active jobs", func(t *testing.T) {
		job1ID := uuid.New()
		job2ID := uuid.New()
		nodeID := uuid.New()

		tracker.TrackJob(ctx, job1ID, nodeID)
		tracker.TrackJob(ctx, job2ID, nodeID)

		jobs, err := tracker.GetActiveJobs(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(jobs) != 2 {
			t.Errorf("expected 2 jobs, got %d", len(jobs))
		}
	})
}

func TestInMemoryJobTracker_GetJobsByNode(t *testing.T) {
	tracker := NewInMemoryJobTracker()
	ctx := context.Background()

	node1ID := uuid.New()
	node2ID := uuid.New()
	job1ID := uuid.New()
	job2ID := uuid.New()
	job3ID := uuid.New()

	tracker.TrackJob(ctx, job1ID, node1ID)
	tracker.TrackJob(ctx, job2ID, node1ID)
	tracker.TrackJob(ctx, job3ID, node2ID)

	t.Run("get jobs for node1", func(t *testing.T) {
		jobs, err := tracker.GetJobsByNode(ctx, node1ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(jobs) != 2 {
			t.Errorf("expected 2 jobs for node1, got %d", len(jobs))
		}
	})

	t.Run("get jobs for node2", func(t *testing.T) {
		jobs, err := tracker.GetJobsByNode(ctx, node2ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(jobs) != 1 {
			t.Errorf("expected 1 job for node2, got %d", len(jobs))
		}
	})

	t.Run("get jobs for non-existent node", func(t *testing.T) {
		jobs, err := tracker.GetJobsByNode(ctx, uuid.New())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(jobs) != 0 {
			t.Errorf("expected 0 jobs for non-existent node, got %d", len(jobs))
		}
	})
}

func TestInMemoryJobTracker_GetJobsByStatus(t *testing.T) {
	tracker := NewInMemoryJobTracker()
	ctx := context.Background()

	job1ID := uuid.New()
	job2ID := uuid.New()
	job3ID := uuid.New()
	nodeID := uuid.New()

	tracker.TrackJob(ctx, job1ID, nodeID)
	tracker.TrackJob(ctx, job2ID, nodeID)
	tracker.TrackJob(ctx, job3ID, nodeID)

	// Update one job to running
	tracker.UpdateJobProgress(ctx, job2ID, &JobProgress{Status: JobStatusRunning})

	t.Run("get assigned jobs", func(t *testing.T) {
		jobs, err := tracker.GetJobsByStatus(ctx, JobStatusAssigned)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(jobs) != 2 {
			t.Errorf("expected 2 assigned jobs, got %d", len(jobs))
		}
	})

	t.Run("get running jobs", func(t *testing.T) {
		jobs, err := tracker.GetJobsByStatus(ctx, JobStatusRunning)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(jobs) != 1 {
			t.Errorf("expected 1 running job, got %d", len(jobs))
		}
	})
}

func TestInMemoryJobTracker_CleanupCompletedJobs(t *testing.T) {
	tracker := NewInMemoryJobTracker()
	ctx := context.Background()

	job1ID := uuid.New()
	job2ID := uuid.New()
	nodeID := uuid.New()

	// Complete two jobs
	tracker.TrackJob(ctx, job1ID, nodeID)
	tracker.TrackJob(ctx, job2ID, nodeID)

	tracker.CompleteJobTracking(ctx, job1ID, true, "")
	tracker.CompleteJobTracking(ctx, job2ID, true, "")

	// Manually set completion time for one job to be old
	history := tracker.completedJobs[job1ID]
	oldTime := time.Now().Add(-2 * time.Hour)
	history.CompletedAt = &oldTime

	err := tracker.CleanupCompletedJobs(ctx, time.Hour)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Old job should be cleaned up
	if _, exists := tracker.completedJobs[job1ID]; exists {
		t.Error("expected old job to be cleaned up")
	}

	// Recent job should remain
	if _, exists := tracker.completedJobs[job2ID]; !exists {
		t.Error("expected recent job to remain")
	}
}

func TestInMemoryJobTracker_GetStaleJobs(t *testing.T) {
	tracker := NewInMemoryJobTracker()
	ctx := context.Background()

	job1ID := uuid.New()
	job2ID := uuid.New()
	nodeID := uuid.New()

	tracker.TrackJob(ctx, job1ID, nodeID)
	tracker.TrackJob(ctx, job2ID, nodeID)

	// Manually set one job's updated time to be stale
	job1 := tracker.activeJobs[job1ID]
	job1.UpdatedAt = time.Now().Add(-2 * time.Hour)

	staleJobs, err := tracker.GetStaleJobs(ctx, time.Hour)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(staleJobs) != 1 {
		t.Errorf("expected 1 stale job, got %d", len(staleJobs))
	}

	if staleJobs[0].JobID != job1ID {
		t.Errorf("expected stale job ID %v, got %v", job1ID, staleJobs[0].JobID)
	}
}

func TestInMemoryJobTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewInMemoryJobTracker()
	ctx := context.Background()
	nodeID := uuid.New()

	done := make(chan bool, 10)

	// Concurrent job tracking
	for i := 0; i < 5; i++ {
		go func() {
			jobID := uuid.New()
			tracker.TrackJob(ctx, jobID, nodeID)
			tracker.UpdateJobProgress(ctx, jobID, &JobProgress{Status: JobStatusRunning})
			tracker.CompleteJobTracking(ctx, jobID, true, "")
			done <- true
		}()
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func() {
			tracker.GetActiveJobs(ctx)
			tracker.GetJobsByNode(ctx, nodeID)
			tracker.GetJobsByStatus(ctx, JobStatusRunning)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic or cause data races
}
