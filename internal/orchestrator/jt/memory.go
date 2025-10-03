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
	activeJobs    map[uuid.UUID][]*JobProgress
	completedJobs map[uuid.UUID][]*JobHistory
	mu            sync.RWMutex
}

func NewInMemoryJobTracker() *InMemoryJobTracker {
	return &InMemoryJobTracker{
		activeJobs:    make(map[uuid.UUID][]*JobProgress),
		completedJobs: make(map[uuid.UUID][]*JobHistory),
	}
}

func (t *InMemoryJobTracker) TrackJob(ctx context.Context, jobID uuid.UUID, nodeID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if job already exists in any node's list
	for _, jobs := range t.activeJobs {
		for _, job := range jobs {
			if job.JobID == jobID {
				return ErrJobAlreadyTracked
			}
		}
	}

	now := time.Now()
	progress := &JobProgress{
		JobID:     jobID,
		NodeID:    nodeID,
		Status:    JobStatusAssigned,
		StartedAt: &now,
		UpdatedAt: now,
	}

	t.activeJobs[nodeID] = append(t.activeJobs[nodeID], progress)
	return nil
}

func (t *InMemoryJobTracker) UpdateJobProgress(ctx context.Context, jobID uuid.UUID, progress *JobProgress) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Find the job in any node's list
	for _, jobs := range t.activeJobs {
		for _, existingProgress := range jobs {
			if existingProgress.JobID == jobID {
				// Update fields
				existingProgress.Status = progress.Status
				existingProgress.UpdatedAt = time.Now()

				if progress.ErrorMessage != "" {
					existingProgress.ErrorMessage = progress.ErrorMessage
				}

				return nil
			}
		}
	}

	return ErrJobProgressNotFound
}

func (t *InMemoryJobTracker) GetJobProgress(ctx context.Context, jobID uuid.UUID) (*JobProgress, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Find the job in any node's list
	for _, jobs := range t.activeJobs {
		for _, progress := range jobs {
			if progress.JobID == jobID {
				// Return a copy to prevent race conditions
				progressCopy := *progress
				return &progressCopy, nil
			}
		}
	}

	return nil, ErrJobProgressNotFound
}

func (t *InMemoryJobTracker) CompleteJobTracking(ctx context.Context, jobID uuid.UUID, success bool, errorMsg string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Find and remove the job from active jobs
	var progress *JobProgress
	var nodeID uuid.UUID
	var jobIndex int
	found := false

	for nid, jobs := range t.activeJobs {
		for i, job := range jobs {
			if job.JobID == jobID {
				progress = job
				nodeID = nid
				jobIndex = i
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
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
	t.completedJobs[nodeID] = append(t.completedJobs[nodeID], history)

	// Remove from active jobs
	t.activeJobs[nodeID] = append(t.activeJobs[nodeID][:jobIndex], t.activeJobs[nodeID][jobIndex+1:]...)
	if len(t.activeJobs[nodeID]) == 0 {
		delete(t.activeJobs, nodeID)
	}

	return nil
}

func (t *InMemoryJobTracker) GetActiveJobs(ctx context.Context) (map[uuid.UUID][]*JobProgress, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[uuid.UUID][]*JobProgress)
	for nodeID, nodeJobs := range t.activeJobs {
		jobs := make([]*JobProgress, 0, len(nodeJobs))
		for _, progress := range nodeJobs {
			// Create copy to prevent race conditions
			progressCopy := *progress
			jobs = append(jobs, &progressCopy)
		}
		result[nodeID] = jobs
	}

	return result, nil
}

func (t *InMemoryJobTracker) GetJobsByNode(ctx context.Context, nodeID uuid.UUID) ([]*JobProgress, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var jobs []*JobProgress
	if nodeJobs, exists := t.activeJobs[nodeID]; exists {
		for _, progress := range nodeJobs {
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
	for _, nodeJobs := range t.activeJobs {
		for _, progress := range nodeJobs {
			if progress.Status == status {
				progressCopy := *progress
				jobs = append(jobs, &progressCopy)
			}
		}
	}

	return jobs, nil
}

func (t *InMemoryJobTracker) CleanupCompletedJobs(ctx context.Context, olderThan time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoffTime := time.Now().Add(-olderThan)

	for nodeID, histories := range t.completedJobs {
		var remainingJobs []*JobHistory
		for _, history := range histories {
			if history.CompletedAt == nil || history.CompletedAt.After(cutoffTime) {
				remainingJobs = append(remainingJobs, history)
			}
		}

		if len(remainingJobs) == 0 {
			delete(t.completedJobs, nodeID)
		} else {
			t.completedJobs[nodeID] = remainingJobs
		}
	}

	return nil
}

func (t *InMemoryJobTracker) GetStaleJobs(ctx context.Context, timeout time.Duration) ([]*JobProgress, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	cutoffTime := time.Now().Add(-timeout)
	var staleJobs []*JobProgress

	for _, nodeJobs := range t.activeJobs {
		for _, progress := range nodeJobs {
			if progress.UpdatedAt.Before(cutoffTime) {
				progressCopy := *progress
				staleJobs = append(staleJobs, &progressCopy)
			}
		}
	}

	return staleJobs, nil
}
