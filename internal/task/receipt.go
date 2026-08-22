package task

import (
	"errors"
	"time"

	"groundops/internal/flight"
	"groundops/internal/store"
)

// ErrTerminalReceipt is returned when a receipt arrives for a terminal flight.
var ErrTerminalReceipt = errors.New("flight is terminal, receipt rejected")

// Receipt accepts a ground-service completion receipt. Receipts for terminal
// flights are rejected so a task can never be completed twice.
func Receipt(state *store.State, taskID string, at time.Time) error {
	t, ok := state.Task(taskID)
	if !ok {
		return errors.New("task does not exist")
	}
	f, ok := state.Flight(t.FlightID)
	if !ok {
		return ErrUnknownFlight
	}
	if flight.IsTerminal(f) {
		return ErrTerminalReceipt
	}
	if t.Status == store.TaskCompleted {
		return errors.New("task is already completed")
	}
	t.Status = store.TaskCompleted
	t.CompletedAt = at
	return state.PutTask(t)
}
