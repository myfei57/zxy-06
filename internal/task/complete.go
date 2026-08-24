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
	if t.ResourceID != "" {
		if err := resource.Release(state, t.ResourceID); err != nil {
			return err
		}
	}
	if err := audit.RecordResult(state, result); err != nil {
		return err
	}
	t.Status = store.TaskCompleted
	t.CompletedAt = at
	return state.PutTask(t)
}
