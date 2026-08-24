package shift

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"groundops/internal/resource"
	"groundops/internal/roster"
	"groundops/internal/store"
	"groundops/internal/task"
)

func newTestState(t *testing.T) *store.State {
	t.Helper()
	dir := t.TempDir()
	return store.NewState(dir)
}

// registerFlightAndTask sets up a flight with one in-progress task owned by the
// outgoing crew. It mirrors how the control plane assigns then activates a task.
func seedInProgressTask(t *testing.T, state *store.State, shiftID string) *store.Task {
	t.Helper()
	f := &store.Flight{
		ID:          uuid.NewString(),
		FlightNo:    "CA1234",
		Status:      store.FlightServicing,
		ScheduledAt: time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		UpdatedSeq:  1,
	}
	if err := state.PutFlight(f); err != nil {
		t.Fatalf("put flight: %v", err)
	}

	tk := &store.Task{
		ID:         uuid.NewString(),
		FlightID:   f.ID,
		TaskType:   "boarding",
		Status:     store.TaskAssigned,
		ShiftID:    shiftID,
		CreatedAt:  time.Now().UTC(),
	}
	if err := state.PutTask(tk); err != nil {
		t.Fatalf("put task: %v", err)
	}

	// Outgoing crew member occupies the task.
	crew, err := resource.Register(state, "night-driver", "personnel")
	if err != nil {
		t.Fatalf("register crew: %v", err)
	}
	if err := resource.Occupy(state, crew.ID, tk.ID); err != nil {
		t.Fatalf("occupy: %v", err)
	}
	tk.Owner = crew.Name
	tk.ResourceID = crew.ID
	tk.Status = store.TaskInProgress
	tk.StartedAt = time.Now().UTC()
	if err := state.PutTask(tk); err != nil {
		t.Fatalf("put task in-progress: %v", err)
	}
	return tk
}

// TestHandoverMigratesOwnerAndResource asserts the regression fix: a handover
// must carry over the executing owner and bound resource together with the
// shift, not flip the shift field alone.
func TestHandoverMigratesOwnerAndResource(t *testing.T) {
	state := newTestState(t)

	from, err := Create(state, "night", time.Now().UTC())
	if err != nil {
		t.Fatalf("create from shift: %v", err)
	}
	to, err := Create(state, "day", time.Now().UTC())
	if err != nil {
		t.Fatalf("create to shift: %v", err)
	}
	// Incoming shift must have personnel to take over.
	dayCrew, err := roster.AddPersonnel(state, to.ID, "day-driver")
	if err != nil {
		t.Fatalf("add day crew: %v", err)
	}

	tk := seedInProgressTask(t, state, from.ID)
	oldResourceID := tk.ResourceID
	oldOwner := tk.Owner

	if err := Handover(state, from.ID, to.ID); err != nil {
		t.Fatalf("handover: %v", err)
	}

	got, ok := state.Task(tk.ID)
	if !ok {
		t.Fatalf("task missing after handover")
	}
	if got.ShiftID != to.ID {
		t.Errorf("shift not migrated: got %q want %q", got.ShiftID, to.ID)
	}
	if got.Owner == "" || got.Owner == oldOwner {
		t.Errorf("owner not migrated: got %q want a day-shift crew member (was %q)",
			got.Owner, oldOwner)
	}
	if got.ResourceID == "" || got.ResourceID == oldResourceID {
		t.Errorf("resource not migrated: got %q want the day-shift resource (was %q)",
			got.ResourceID, oldResourceID)
	}
	if got.ResourceID != dayCrew.ResourceID {
		t.Errorf("task not bound to incoming crew: got %q want %q",
			got.ResourceID, dayCrew.ResourceID)
	}
	// The new owner must be accountable: the bound resource is occupied by the task.
	bound, ok := state.Resource(got.ResourceID)
	if !ok {
		t.Fatalf("new bound resource missing")
	}
	if bound.Status != store.ResourceOccupied || bound.TaskID != got.ID {
		t.Errorf("new crew not occupied by task: status=%q taskID=%q",
			bound.Status, bound.TaskID)
	}
	// The old binding is released so the outgoing crew is not left occupied.
	if old, ok := state.Resource(oldResourceID); ok {
		if old.Status != store.ResourceIdle {
			t.Errorf("old resource not released: status=%q", old.Status)
		}
	}
}

// TestHandoverRejectsNoAvailableCrew ensures a handover fails loudly when the
// incoming shift has no idle crew to take the work, instead of silently
// leaving tasks without an owner.
func TestHandoverRejectsNoAvailableCrew(t *testing.T) {
	state := newTestState(t)

	from, err := Create(state, "night", time.Now().UTC())
	if err != nil {
		t.Fatalf("create from shift: %v", err)
	}
	to, err := Create(state, "day", time.Now().UTC())
	if err != nil {
		t.Fatalf("create to shift: %v", err)
	}
	// Day shift has a crew member, but they are already busy on another task.
	dayCrew, err := roster.AddPersonnel(state, to.ID, "day-driver")
	if err != nil {
		t.Fatalf("add day crew: %v", err)
	}
	other := &store.Task{
		ID:         uuid.NewString(),
		Status:     store.TaskInProgress,
		ShiftID:    to.ID,
		Owner:      dayCrew.ResourceID,
		ResourceID: dayCrew.ResourceID,
		CreatedAt:  time.Now().UTC(),
	}
	if err := state.PutTask(other); err != nil {
		t.Fatalf("put other task: %v", err)
	}
	if err := resource.Occupy(state, dayCrew.ResourceID, other.ID); err != nil {
		t.Fatalf("occupy day crew: %v", err)
	}

	_ = seedInProgressTask(t, state, from.ID)

	err = Handover(state, from.ID, to.ID)
	if err == nil {
		t.Fatalf("expected error when no idle crew, got nil")
	}
}

// TestHandoverLeavesOtherShiftsAlone ensures tasks belonging to a third shift
// are not swept into the handover.
func TestHandoverLeavesOtherShiftsAlone(t *testing.T) {
	state := newTestState(t)

	from, err := Create(state, "night", time.Now().UTC())
	if err != nil {
		t.Fatalf("create from shift: %v", err)
	}
	to, err := Create(state, "day", time.Now().UTC())
	if err != nil {
		t.Fatalf("create to shift: %v", err)
	}
	third, err := Create(state, "swing", time.Now().UTC())
	if err != nil {
		t.Fatalf("create third shift: %v", err)
	}
	if _, err := roster.AddPersonnel(state, to.ID, "day-driver"); err != nil {
		t.Fatalf("add day crew: %v", err)
	}

	other := seedInProgressTask(t, state, third.ID)
	_ = seedInProgressTask(t, state, from.ID)

	if err := Handover(state, from.ID, to.ID); err != nil {
		t.Fatalf("handover: %v", err)
	}
	got, ok := state.Task(other.ID)
	if !ok {
		t.Fatalf("third-shift task missing")
	}
	if got.ShiftID != third.ID {
		t.Errorf("third-shift task migrated: got %q want %q", got.ShiftID, third.ID)
	}
	if got.Owner == "" {
		t.Errorf("third-shift task lost its owner")
	}
}

// sanity: the task package is imported to guard against import pruning.
var _ = task.IsActive
