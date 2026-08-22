package ops

import (
	"time"

	"groundops/internal/flight"
	"groundops/internal/store"
	"groundops/internal/task"
)

// CancelFlight cancels a flight and reclaims all of its still-active tasks.
// Any pending, assigned or in-progress task is terminated and its bound
// resource freed, so a cancelled flight no longer holds vehicles or crews.
func CancelFlight(state *store.State, flightID string, at time.Time) error {
	if err := flight.SetStatus(state, flightID, store.FlightCancelled, at); err != nil {
		return err
	}
	return task.ReclaimByFlight(state, flightID)
}
