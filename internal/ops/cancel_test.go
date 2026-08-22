package ops

import (
	"testing"
	"time"

	"groundops/internal/flight"
	"groundops/internal/recover"
	"groundops/internal/resource"
	"groundops/internal/store"
	"groundops/internal/task"
)

// registerServicedFlight sets up a flight already in the servicing state so its
// tasks can be assigned and started.
func registerServicedFlight(t *testing.T, state *store.State, no string, at time.Time) *store.Flight {
	t.Helper()
	f, err := flight.Register(state, no, at)
	if err != nil {
		t.Fatalf("register flight: %v", err)
	}
	if err := flight.SetStatus(state, f.ID, store.FlightServicing, at); err != nil {
		t.Fatalf("service flight: %v", err)
	}
	return f
}

func TestCancelFlightReclaimsAssignedTasksAndResources(t *testing.T) {
	state := store.NewState(t.TempDir())
	at := time.Date(2026, 8, 22, 3, 0, 0, 0, time.UTC)
	f := registerServicedFlight(t, state, "MU9999", at)

	// Two vehicles for the cancelled flight: a shuttle bus and a baggage cart.
	bus, _ := resource.Register(state, "shuttle-1", "vehicle")
	cart, _ := resource.Register(state, "baggage-1", "vehicle")

	busTask, _ := task.Create(state, f.ID, "shuttle")
	cartTask, _ := task.Create(state, f.ID, "baggage")
	if err := task.Assign(state, busTask.ID, "crew-A", bus.ID, at); err != nil {
		t.Fatalf("assign bus task: %v", err)
	}
	if err := task.Assign(state, cartTask.ID, "crew-B", cart.ID, at); err != nil {
		t.Fatalf("assign cart task: %v", err)
	}
	// Both tasks are in progress when the cancellation lands: the vehicles are
	// already on their way to the stand.
	if err := task.Start(state, busTask.ID, at); err != nil {
		t.Fatalf("start bus task: %v", err)
	}
	if err := task.Start(state, cartTask.ID, at); err != nil {
		t.Fatalf("start cart task: %v", err)
	}

	if err := CancelFlight(state, f.ID, at.Add(time.Minute)); err != nil {
		t.Fatalf("cancel flight: %v", err)
	}

	if got := flight.LiveStatus(f); got != store.FlightCancelled {
		t.Fatalf("flight status = %q, want %q", got, store.FlightCancelled)
	}

	for _, c := range []struct {
		label string
		task  *store.Task
		res   *store.Resource
	}{
		{"shuttle", mustTask(state, busTask.ID), mustResource(state, bus.ID)},
		{"baggage", mustTask(state, cartTask.ID), mustResource(state, cart.ID)},
	} {
		if c.task.Status != store.TaskCancelled {
			t.Errorf("%s task status = %q, want %q", c.label, c.task.Status, store.TaskCancelled)
		}
		if task.IsActive(c.task) {
			t.Errorf("%s task still active: status=%q", c.label, c.task.Status)
		}
		if c.task.Owner != "" || c.task.ResourceID != "" || c.task.ShiftID != "" {
			t.Errorf("%s task still bound: owner=%q resource=%q shift=%q", c.label, c.task.Owner, c.task.ResourceID, c.task.ShiftID)
		}
		if c.res.Status != store.ResourceIdle {
			t.Errorf("%s resource not released: status=%q task=%q", c.label, c.res.Status, c.res.TaskID)
		}
		if c.res.TaskID != "" {
			t.Errorf("%s resource still bound to task %q", c.label, c.res.TaskID)
		}
	}

	// Cancelled tasks must not surface as open work on the flight.
	if open := task.OpenForFlight(state, f.ID); len(open) != 0 {
		t.Errorf("cancelled flight still has open tasks: %d", len(open))
	}
}

// TestCancelFlightViaDynamicsUpdate exercises the dynamics-message path: when
// an inbound flight update carries the cancelled status, the same reclaim must
// happen, not just a status flip.
func TestCancelFlightViaDynamicsUpdate(t *testing.T) {
	state := store.NewState(t.TempDir())
	at := time.Date(2026, 8, 22, 3, 0, 0, 0, time.UTC)
	f := registerServicedFlight(t, state, "MU1234", at)

	bus, _ := resource.Register(state, "shuttle-2", "vehicle")
	tk, _ := task.Create(state, f.ID, "shuttle")
	if err := task.Assign(state, tk.ID, "crew-C", bus.ID, at); err != nil {
		t.Fatalf("assign: %v", err)
	}
	if err := task.Start(state, tk.ID, at); err != nil {
		t.Fatalf("start: %v", err)
	}

	// The dynamics feed reports the flight as cancelled one seq later.
	update := store.FlightUpdate{Seq: f.UpdatedSeq + 1, Status: store.FlightCancelled, EventAt: at.Add(time.Minute)}
	if err := ApplyUpdate(state, f.ID, update); err != nil {
		t.Fatalf("apply update: %v", err)
	}

	if got := flight.LiveStatus(f); got != store.FlightCancelled {
		t.Fatalf("flight status = %q, want %q", got, store.FlightCancelled)
	}
	tk2, _ := state.Task(tk.ID)
	if tk2.Status != store.TaskCancelled {
		t.Fatalf("task status = %q, want %q", tk2.Status, store.TaskCancelled)
	}
	bus2, _ := state.Resource(bus.ID)
	if bus2.Status != store.ResourceIdle {
		t.Fatalf("resource status = %q, want %q", bus2.Status, store.ResourceIdle)
	}
}

// TestRecoverSweepDoesNotReviveCancelledTasks guards against the recovery sweep
// resurrecting cancelled tasks back into assignment. With resources released,
// nothing should be redispatched and the task must stay cancelled.
func TestRecoverSweepDoesNotReviveCancelledTasks(t *testing.T) {
	state := store.NewState(t.TempDir())
	at := time.Date(2026, 8, 22, 3, 0, 0, 0, time.UTC)
	f := registerServicedFlight(t, state, "MU5555", at)

	// Park a task still bound to a busy resource, then cancel: release happens
	// during reclaim, so the sweep has nothing to re-dispatch.
	bus, _ := resource.Register(state, "shuttle-3", "vehicle")
	tk, _ := task.Create(state, f.ID, "shuttle")
	if err := task.Assign(state, tk.ID, "crew-D", bus.ID, at); err != nil {
		t.Fatalf("assign: %v", err)
	}
	if err := task.Start(state, tk.ID, at); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := CancelFlight(state, f.ID, at.Add(time.Minute)); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	// Stale escalation timer pointing at the now-cancelled task must not be
	// treated as active work either.
	_ = state.PutEscalation(&store.Escalation{TaskID: tk.ID, Level: 0, DueAt: at, State: store.EscalationPending})

	if _, err := recover.Sweep(state, at.Add(2*time.Minute)); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	tk2, _ := state.Task(tk.ID)
	if tk2.Status != store.TaskCancelled {
		t.Fatalf("sweep revived cancelled task: status=%q", tk2.Status)
	}
	bus2, _ := state.Resource(bus.ID)
	if bus2.Status != store.ResourceIdle {
		t.Fatalf("sweep re-occupied released resource: status=%q", bus2.Status)
	}
}

func mustTask(state *store.State, id string) *store.Task {
	t, ok := state.Task(id)
	if !ok {
		panic("task not found: " + id)
	}
	return t
}

func mustResource(state *store.State, id string) *store.Resource {
	r, ok := state.Resource(id)
	if !ok {
		panic("resource not found: " + id)
	}
	return r
}
