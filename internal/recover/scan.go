package recover

import (
	"time"

	"groundops/internal/store"
)

// Sweep re-dispatches resources whose tasks are still active. Completed and
// cancelled tasks are left untouched, so a sweep never re-binds a finished or
// cancelled task's resource.
func Sweep(state *store.State, at time.Time) (int, error) {
	redispatched := 0
	for _, r := range state.Resources() {
		if r.Status != store.ResourceOccupied || r.TaskID == "" {
			continue
		}
		t, ok := state.Task(r.TaskID)
		if !ok {
			continue
		}
		if t.Status == store.TaskCompleted || t.Status == store.TaskCancelled {
			continue
		}
		// Re-dispatch active tasks that lost their binding.
		if t.Status == store.TaskPending || t.Status == store.TaskFailed {
			t.Status = store.TaskAssigned
			t.Owner = r.Name
			if err := state.PutTask(t); err != nil {
				return redispatched, err
			}
			redispatched++
		}
	}
	return redispatched, nil
}
