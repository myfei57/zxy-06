package sla

import (
	"testing"
	"time"

	"groundops/internal/store"
)

// registerFlight builds a flight whose UpdatedAt starts at the scheduled time
// (the same initial condition as flight.Register).
func registerFlight(t *testing.T, s *store.State, flightNo, id string, scheduledAt time.Time) *store.Flight {
	t.Helper()
	f := &store.Flight{
		ID:          id,
		FlightNo:    flightNo,
		Status:      store.FlightScheduled,
		ScheduledAt: scheduledAt,
		UpdatedAt:   scheduledAt,
		UpdatedSeq:  1,
	}
	if err := s.PutFlight(f); err != nil {
		t.Fatalf("PutFlight: %v", err)
	}
	return f
}

func registerTask(t *testing.T, s *store.State, id, flightID string, createdAt time.Time) *store.Task {
	t.Helper()
	tk := &store.Task{ID: id, FlightID: flightID, Status: store.TaskPending, CreatedAt: createdAt}
	if err := s.PutTask(tk); err != nil {
		t.Fatalf("PutTask: %v", err)
	}
	return tk
}

// TestEnsureTimer_AnchoredOnLatestDynamic is the regression test for the
// delayed-flight escalation bug: the SLA deadline must follow the flight's
// latest dynamic (UpdatedAt), not the original scheduled time, so a delayed
// flight pushes the escalation window out instead of counting down on a plan
// the flight no longer follows.
func TestEnsureTimer_AnchoredOnLatestDynamic(t *testing.T) {
	s := store.NewState(t.TempDir())
	scheduled := time.Date(2026, 8, 22, 1, 0, 0, 0, time.UTC)

	f := registerFlight(t, s, "KA0001", "flt-1", scheduled)
	tk := registerTask(t, s, "tsk-1", f.ID, scheduled)

	// On-time case: no delay, UpdatedAt == ScheduledAt, deadline is plan + window.
	if err := EnsureTimer(s, tk.ID, f.ID, scheduled); err != nil {
		t.Fatalf("EnsureTimer: %v", err)
	}
	e, ok := s.Escalation(tk.ID)
	if !ok {
		t.Fatal("escalation not created")
	}
	wantDue := scheduled.Add(LevelMinutes * time.Minute)
	if !e.DueAt.Equal(wantDue) {
		t.Fatalf("on-time DueAt = %v, want %v", e.DueAt, wantDue)
	}

	// Delayed case: flight gets a dynamics message two hours after the plan.
	// The escalation deadline must move with the latest dynamic.
	delayedAt := scheduled.Add(2 * time.Hour)
	if err := flightUpdate(s, f.ID, delayedAt); err != nil {
		t.Fatalf("flightUpdate: %v", err)
	}
	if err := EnsureTimer(s, tk.ID, f.ID, delayedAt); err != nil {
		t.Fatalf("EnsureTimer: %v", err)
	}
	e, _ = s.Escalation(tk.ID)
	wantDue = delayedAt.Add(LevelMinutes * time.Minute)
	if !e.DueAt.Equal(wantDue) {
		t.Fatalf("delayed DueAt = %v, want %v (deadline must track latest dynamic, not %v)",
			e.DueAt, wantDue, scheduled.Add(LevelMinutes*time.Minute))
	}

	// A tick at the original scheduled-window end must NOT fire the escalation,
	// because the deadline now tracks the delayed dynamic.
	staleTick := scheduled.Add(LevelMinutes * time.Minute)
	fired, err := Tick(s, staleTick)
	if err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if fired != 0 {
		t.Fatalf("Tick at stale plan-time fired %d escalations, want 0", fired)
	}
	e, _ = s.Escalation(tk.ID)
	if e.Level != 0 {
		t.Fatalf("level = %d, want 0 before the delayed deadline", e.Level)
	}

	// A tick past the delayed deadline fires the escalation.
	fired, err = Tick(s, delayedAt.Add(LevelMinutes*time.Minute))
	if err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if fired != 1 {
		t.Fatalf("Tick at delayed deadline fired %d, want 1", fired)
	}
	e, _ = s.Escalation(tk.ID)
	if e.Level != 1 {
		t.Fatalf("level = %d, want 1", e.Level)
	}
}

// flightUpdate applies a dynamics message by advancing UpdatedAt, mirroring
// flight.Update without pulling in cross-package dependencies from the sla test.
func flightUpdate(s *store.State, flightID string, at time.Time) error {
	f, ok := s.Flight(flightID)
	if !ok {
		return errFlightMissing
	}
	f.UpdatedAt = at
	f.UpdatedSeq++
	return s.PutFlight(f)
}
