package task

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"groundops/internal/audit"
	"groundops/internal/resource"
	"groundops/internal/store"
)

// Complete persists the work result first, then flips the task to completed
// and only then releases the resource.
func Complete(state *store.State, taskID string, result *store.TaskResult, at time.Time) error {
	t, ok := state.Task(taskID)
	if !ok {
		return errors.New("task does not exist")
	}
	if t.Status != store.TaskInProgress {
		return errors.New("task is not in progress")
	}
	if result.ID == "" {
		result.ID = uuid.NewString()
	}
	result.TaskID = taskID
	result.At = at
	// Persist the work result and flip the task to completed before
	// releasing the resource. Releasing first would return the resource to
	// the idle pool while the completion record is not yet durable; if the
	// record write then failed, the task would still show in progress while
	// the resource is already free for another task to grab.
	if err := audit.RecordResult(state, result); err != nil {
		return err
	}
	t.Status = store.TaskCompleted
	t.CompletedAt = at
	if err := state.PutTask(t); err != nil {
		return err
	}
	if t.ResourceID != "" {
		if err := resource.Release(state, t.ResourceID); err != nil {
			return err
		}
	}
	return nil
}
