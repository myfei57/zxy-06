package ops

import (
	"time"

	"groundops/internal/flight"
	"groundops/internal/store"
	"groundops/internal/task"
)

// CancelFlight cancels a flight and reclaims all of its assigned tasks.
func CancelFlight(state *store.State, flightID string, at time.Time) error {
	if err := flight.SetStatus(state, flightID, store.FlightCancelled, at); err != nil {
		return err
	}
	return task.ReclaimByFlight(state, flightID)
}
