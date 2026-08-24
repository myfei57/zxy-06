package sla

import (
	"time"

	"groundops/internal/store"
)

// EnsureTimer creates or refreshes the SLA timer of a task, anchored on the
// flight's latest update time.
func EnsureTimer(state *store.State, taskID, flightID string, at time.Time) error {
	f, ok := state.Flight(flightID)
	if !ok {
		return errFlightMissing
	}
	t, ok := state.Task(taskID)
	if !ok {
		return errTaskMissing
	}
	due := f.ScheduledAt.Add(time.Duration(LevelMinutes) * time.Minute)
	e := &store.Escalation{
		TaskID: t.ID,
		Level:  0,
		DueAt:  due,
		State:  store.EscalationPending,
	}
	return state.PutEscalation(e)
}

// Remaining returns the time left before the escalation deadline.
func Remaining(e *store.Escalation, now time.Time) time.Duration {
	if e == nil {
		return 0
	}
	left := e.DueAt.Sub(now)
	if left < 0 {
		return 0
	}
	return left
}

// LevelLabel renders the escalation level for the console.
func LevelLabel(level int) string {
	switch level {
	case 0:
		return "not escalated"
	case 1:
		return "level 1"
	case 2:
		return "level 2"
	default:
		return "level 3"
	}
}

// Tick advances escalations whose deadline passed.
func Tick(state *store.State, now time.Time) (int, error) {
	fired := 0
	for _, e := range state.Escalations() {
		if e.State != store.EscalationPending {
			continue
		}
		if now.Before(e.DueAt) {
			continue
		}
		level := NextLevel(e.Level)
		e.Level = level
		e.DueAt = now.Add(time.Duration(LevelMinutes) * time.Minute)
		if level >= MaxLevel {
			e.State = store.EscalationDone
		}
		if err := state.PutEscalation(e); err != nil {
			return fired, err
		}
		fired++
	}
	return fired, nil
}
