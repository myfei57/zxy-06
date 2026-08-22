package verifycase

import (
	"testing"
	"time"

	"groundops/internal/flight"
	"groundops/internal/recover"
	"groundops/internal/resource"
	"groundops/internal/store"
	"groundops/internal/task"
)

func TestRecoverySkipsCompletedTasks(t *testing.T) {
	dir := t.TempDir()
	state := store.NewState(dir)
	t0 := time.Now().UTC()
	f, err := flight.Register(state, "MF8361", t0)
	if err != nil {
		t.Fatal(err)
	}
	res, err := resource.Register(state, "truck-2", "vehicle")
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
	if err := task.Start(state, tk.ID, t0); err != nil {
		t.Fatal(err)
	}
	if err := task.Receipt(state, tk.ID, t0.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := recover.Sweep(state, t0.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	after, _ := state.Task(tk.ID)
	if after.Status != store.TaskCompleted {
		t.Fatalf("recovery re-dispatched a completed task to %s", after.Status)
	}
}
