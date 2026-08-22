package task

import (
	"errors"
	"time"

	"groundops/internal/flight"
	"groundops/internal/resource"
	"groundops/internal/roster"
	"groundops/internal/store"
)

// ErrFlightTerminal is returned when assigning a task to a terminal flight.
var ErrFlightTerminal = errors.New("flight is in a terminal state")

// ErrResourceBusy is returned when a resource is already occupied.
var ErrResourceBusy = errors.New("resource is occupied")

// Assign binds a task to an owner and a resource after checking the live
// flight state and the resource occupancy.
func Assign(state *store.State, taskID, owner, resourceID string, at time.Time) error {
	t, ok := state.Task(taskID)
	if !ok {
		return errors.New("task does not exist")
	}
	f, ok := state.Flight(t.FlightID)
	if !ok {
		return ErrUnknownFlight
	}
	// never dispatch personnel or vehicles to a flight that has already
	// departed or been cancelled.
	if !flight.IsServicable(f) {
		return ErrFlightTerminal
	}
	r, ok := state.Resource(resourceID)
	if !ok {
		return errors.New("resource does not exist")
	}
	if r.Status == store.ResourceOccupied {
		return ErrResourceBusy
	}
	if err := resource.Occupy(state, resourceID, t.ID); err != nil {
		return err
	}
	t.Status = store.TaskAssigned
	t.Owner = owner
	t.ResourceID = resourceID
	t.ShiftID = roster.CurrentShiftID(state)
	return state.PutTask(t)
}

// ReclaimByFlight recycles every assigned or in-progress task of a flight
// back to pending and releases its resources.
func ReclaimByFlight(state *store.State, flightID string) error {
	for _, t := range state.TasksByFlight(flightID) {
		if t.Status == store.TaskAssigned || t.Status == store.TaskInProgress {
			if t.ResourceID != "" {
				if err := resource.Release(state, t.ResourceID); err != nil {
					return err
				}
			}
			t.Status = store.TaskPending
			t.Owner = ""
			t.ResourceID = ""
			t.ShiftID = ""
			if err := state.PutTask(t); err != nil {
				return err
			}
		}
	}
	return nil
}

// PlannerPool returns personnel resources available for the current shift.
func PlannerPool(state *store.State) []string {
	roster.RefreshSnapshot(state)
	return roster.CurrentSnapshot(state)
}
