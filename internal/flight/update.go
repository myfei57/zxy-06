package flight

import (
	"errors"
	"time"

	"groundops/internal/store"
)

// ErrStaleUpdate is returned when a flight update is older than the current one.
var ErrStaleUpdate = errors.New("flight update is stale")

// Update applies one flight dynamics message. Updates whose sequence is not
// newer than the last accepted sequence are rejected, so an old message can
// never overwrite a newer flight or task state.
func Update(state *store.State, flightID string, update store.FlightUpdate) error {
	f, ok := state.Flight(flightID)
	if !ok {
		return errors.New("flight does not exist")
	}
	if update.Seq <= f.UpdatedSeq {
		return ErrStaleUpdate
	}
	if update.Status != "" {
		f.Status = update.Status
	}
	if !update.EventAt.IsZero() {
		f.UpdatedAt = update.EventAt
	}
	f.UpdatedSeq = update.Seq
	return state.PutFlight(f)
}

// SetStatus changes the flight status at time at.
func SetStatus(state *store.State, flightID, status string, at time.Time) error {
	f, ok := state.Flight(flightID)
	if !ok {
		return errors.New("flight does not exist")
	}
	f.Status = status
	f.UpdatedAt = at
	f.UpdatedSeq++
	return state.PutFlight(f)
}
