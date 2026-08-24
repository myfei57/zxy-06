package task

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"groundops/internal/store"
)

// ErrUnknownFlight is returned when creating a task for a missing flight.
var ErrUnknownFlight = errors.New("flight does not exist")

// Create creates a pending ground task for a flight.
func Create(state *store.State, flightID, taskType string) (*store.Task, error) {
	taskType = strings.TrimSpace(taskType)
	if taskType == "" {
		return nil, errors.New("task type is empty")
	}
	if _, ok := state.Flight(flightID); !ok {
		return nil, ErrUnknownFlight
	}
	t := &store.Task{
		ID:        uuid.NewString(),
		FlightID:  flightID,
		TaskType:  taskType,
		Status:    store.TaskPending,
		CreatedAt: time.Now().UTC(),
	}
	if err := state.PutTask(t); err != nil {
		return nil, err
	}
	return t, nil
}

// Start moves an assigned task into execution.
func Start(state *store.State, taskID string, at time.Time) error {
	t, ok := state.Task(taskID)
	if !ok {
		return errors.New("task does not exist")
	}
	if t.Status != store.TaskAssigned {
		return errors.New("task is not assigned")
	}
	t.Status = store.TaskInProgress
	t.StartedAt = at
	return state.PutTask(t)
}
