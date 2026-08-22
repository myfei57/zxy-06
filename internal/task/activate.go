package task

import (
	"groundops/internal/store"
)

// ActivateForFlight starts the assigned tasks of a flight. Tasks of terminal
// flights are never re-activated, so a stale flight update cannot revive
// completed work.
func ActivateForFlight(state *store.State, flightID string) error {
	f, ok := state.Flight(flightID)
	if !ok {
		return nil
	}
	for _, t := range state.TasksByFlight(flightID) {
		if t.Status == store.TaskAssigned {
			t.Status = store.TaskInProgress
			t.StartedAt = f.UpdatedAt
			if err := state.PutTask(t); err != nil {
				return err
			}
		}
	}
	return nil
}
