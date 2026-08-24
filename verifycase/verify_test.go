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

func TestCancelReclaimsAssignedTasks(t *testing.T) {
	dir := t.TempDir()
	state := store.NewState(dir)
	t0 := time.Now().UTC()
	f, err := flight.Register(state, "MU5102", t0)
	if err != nil {
		t.Fatal(err)
	}
	res, err := resource.Register(state, "truck-1", "vehicle")
	if err != nil {
		t.Fatal(err)
	}
	tk, err := task.Create(state, f.ID, "luggage")
	if err != nil {
		t.Fatal(err)
	}
	if err := task.Assign(state, tk.ID, "crew-a", res.ID, t0); err != nil {
		t.Fatal(err)
	}
	if err := ops.CancelFlight(state, f.ID, t0.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	after, _ := state.Task(tk.ID)
	if after.Status == store.TaskAssigned || after.Status == store.TaskInProgress {
		t.Fatalf("task still active (%s) after cancellation", after.Status)
	}
	r, _ := state.Resource(res.ID)
	if r.Status != store.ResourceIdle {
		t.Fatalf("resource still %s after cancellation", r.Status)
	}
}
