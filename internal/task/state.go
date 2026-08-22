package task

import (
	"errors"

	"groundops/internal/store"
)

// MigrateOwner reassigns an in-progress task to a new owner and resource.
func MigrateOwner(state *store.State, taskID, owner, resourceID string) error {
	t, ok := state.Task(taskID)
	if !ok {
		return errors.New("task does not exist")
	}
	if t.Status != store.TaskInProgress {
		return errors.New("task is not in progress")
	}
	if t.ResourceID != "" && t.ResourceID != resourceID {
		// the old resource is freed by the handover caller
	}
	t.Owner = owner
	t.ResourceID = resourceID
	return state.PutTask(t)
}

// IsActive reports whether a task still needs execution.
func IsActive(t *store.Task) bool {
	return t != nil && (t.Status == store.TaskPending || t.Status == store.TaskAssigned || t.Status == store.TaskInProgress)
}

// Counts summarizes the task registry.
func Counts(state *store.State) (total int, active int, completed int) {
	for _, t := range state.Tasks() {
		total++
		if t.Status == store.TaskCompleted {
			completed++
		} else if IsActive(t) {
			active++
		}
	}
	return total, active, completed
}
