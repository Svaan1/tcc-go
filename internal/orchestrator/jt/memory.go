package jt

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrJobProgressNotFound = errors.New("job progress not found")
	ErrJobAlreadyTracked   = errors.New("job already being tracked")
)

type InMemoryJobTracker struct {
	activeJobs    map[uuid.UUID]*JobProgress
	completedJobs map[uuid.UUID]*JobHistory
	mu            sync.RWMutex
}

func NewInMemoryJobTracker() *InMemoryJobTracker {
	return &InMemoryJobTracker{
		activeJobs:    make(map[uuid.UUID]*JobProgress),
		completedJobs: make(map[uuid.UUID]*JobHistory),
	}
}

func (t *InMemoryJobTracker) TrackJob(ctx context.Context, jobID uuid.UUID, nodeID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.activeJobs[jobID]; exists {
		return ErrJobAlreadyTracked
	}

	now := time.Now()
	progress := &JobProgress{
		JobID:     jobID,
		NodeID:    nodeID,
		Status:    JobStatusAssigned,
		StartedAt: &now,
		UpdatedAt: now,
	}

	t.activeJobs[jobID] = progress
	return nil
}

func (t *InMemoryJobTracker) UpdateJobProgress(ctx context.Context, jobID uuid.UUID, progress *JobProgress) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	existingProgress, exists := t.activeJobs[jobID]
	if !exists {
		return ErrJobProgressNotFound
	}

	// Update fields
	existingProgress.Status = progress.Status
	existingProgress.UpdatedAt = time.Now()

	if progress.ErrorMessage != "" {
		existingProgress.ErrorMessage = progress.ErrorMessage
	}

	return nil
}

func (t *InMemoryJobTracker) GetJobProgress(ctx context.Context, jobID uuid.UUID) (*JobProgress, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	progress, exists := t.activeJobs[jobID]
	if !exists {
		return nil, ErrJobProgressNotFound
	}

	// Return a copy to prevent race conditions
	progressCopy := *progress
	return &progressCopy, nil
}

func (t *InMemoryJobTracker) CompleteJobTracking(ctx context.Context, jobID uuid.UUID, success bool, errorMsg string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	progress, exists := t.activeJobs[jobID]
	if !exists {
		return ErrJobProgressNotFound
	}

	now := time.Now()

	// Update progress status
	if success {
		progress.Status = JobStatusCompleted
	} else {
		progress.Status = JobStatusFailed
		progress.ErrorMessage = errorMsg
	}
	progress.CompletedAt = &now
	progress.UpdatedAt = now

	// Create history entry
	var duration time.Duration
	if progress.StartedAt != nil {
		duration = now.Sub(*progress.StartedAt)
	}

	history := &JobHistory{
		JobID:        jobID,
		NodeID:       progress.NodeID,
		Status:       progress.Status,
		StartedAt:    *progress.StartedAt,
		CompletedAt:  &now,
		Duration:     duration,
		RetryCount:   0, // TODO: Implement retry tracking
		ErrorMessage: progress.ErrorMessage,
	}

	// Move from active to completed
	t.completedJobs[jobID] = history
	delete(t.activeJobs, jobID)

	return nil
}

func (t *InMemoryJobTracker) GetActiveJobs(ctx context.Context) ([]*JobProgress, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	jobs := make([]*JobProgress, 0, len(t.activeJobs))
	for _, progress := range t.activeJobs {
		// Create copy to prevent race conditions
		progressCopy := *progress
		jobs = append(jobs, &progressCopy)
	}

	return jobs, nil
}

func (t *InMemoryJobTracker) GetJobsByNode(ctx context.Context, nodeID uuid.UUID) ([]*JobProgress, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var jobs []*JobProgress
	for _, progress := range t.activeJobs {
		if progress.NodeID == nodeID {
			progressCopy := *progress
			jobs = append(jobs, &progressCopy)
		}
	}

	return jobs, nil
}

func (t *InMemoryJobTracker) GetJobsByStatus(ctx context.Context, status JobStatus) ([]*JobProgress, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var jobs []*JobProgress
	for _, progress := range t.activeJobs {
		if progress.Status == status {
			progressCopy := *progress
			jobs = append(jobs, &progressCopy)
		}
	}

	return jobs, nil
}

func (t *InMemoryJobTracker) CleanupCompletedJobs(ctx context.Context, olderThan time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoffTime := time.Now().Add(-olderThan)

	for jobID, history := range t.completedJobs {
		if history.CompletedAt != nil && history.CompletedAt.Before(cutoffTime) {
			delete(t.completedJobs, jobID)
		}
	}

	return nil
}

func (t *InMemoryJobTracker) GetStaleJobs(ctx context.Context, timeout time.Duration) ([]*JobProgress, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	cutoffTime := time.Now().Add(-timeout)
	var staleJobs []*JobProgress

	for _, progress := range t.activeJobs {
		if progress.UpdatedAt.Before(cutoffTime) {
			progressCopy := *progress
			staleJobs = append(staleJobs, &progressCopy)
		}
	}

	return staleJobs, nil
}
