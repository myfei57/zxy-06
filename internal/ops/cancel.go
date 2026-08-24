package ops

import (
	"time"

	"groundops/internal/flight"
	"groundops/internal/store"
)

// CancelFlight cancels a flight and reclaims all of its assigned tasks.
func CancelFlight(state *store.State, flightID string, at time.Time) error {
	return flight.SetStatus(state, flightID, store.FlightCancelled, at)
}
