package verifycase

import (
	"testing"
	"time"

	"groundops/internal/flight"
	"groundops/internal/ops"
	"groundops/internal/resource"
	"groundops/internal/store"
	"groundops/internal/task"
)

func TestStaleUpdateKeepsTaskTerminal(t *testing.T) {
	dir := t.TempDir()
	state := store.NewState(dir)
	t0 := time.Now().UTC()
	f, err := flight.Register(state, "SC4789", t0)
	if err != nil {
		t.Fatal(err)
	}
	res, err := resource.Register(state, "bus-1", "vehicle")
	if err != nil {
		t.Fatal(err)
	}
	tk, err := task.Create(state, f.ID, "transfer")
	if err != nil {
		t.Fatal(err)
	}
	if err := task.Assign(state, tk.ID, "crew-a", res.ID, t0); err != nil {
		t.Fatal(err)
	}
	if err := task.Start(state, tk.ID, t0); err != nil {
		t.Fatal(err)
	}
	result := &store.TaskResult{Signer: "crew-a", Note: "done"}
	if err := task.Complete(state, tk.ID, result, t0.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := flight.SetStatus(state, f.ID, store.FlightDeparted, t0.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	stale := store.FlightUpdate{Seq: 1, Status: store.FlightServicing, EventAt: t0}
	if err := ops.ApplyUpdate(state, f.ID, stale); err == nil {
		t.Fatal("expected the stale update to be rejected")
	}
	after, _ := state.Task(tk.ID)
	if after.Status != store.TaskCompleted {
		t.Fatalf("stale update re-activated the task to %s", after.Status)
	}
	ff, _ := state.Flight(f.ID)
	if ff.Status != store.FlightDeparted {
		t.Fatalf("stale update reverted flight state to %s", ff.Status)
	}
}
