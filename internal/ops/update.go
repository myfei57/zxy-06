package ops

import (
	"groundops/internal/flight"
	"groundops/internal/store"
	"groundops/internal/task"
)

// ApplyUpdate applies one flight dynamics message and re-evaluates tasks when
// the flight moves into servicing.
func ApplyUpdate(state *store.State, flightID string, update store.FlightUpdate) error {
	if err := flight.Update(state, flightID, update); err != nil {
		return err
	}
	if update.Status == store.FlightServicing {
		return task.ActivateForFlight(state, flightID)
	}
	return nil
}
