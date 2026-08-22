package verifycase

import (
	"testing"
	"time"

	"groundops/internal/flight"
	"groundops/internal/resource"
	"groundops/internal/store"
	"groundops/internal/task"
)

func TestTerminalFlightRejectsReceipt(t *testing.T) {
	dir := t.TempDir()
	state := store.NewState(dir)
	t0 := time.Now().UTC()
	f, err := flight.Register(state, "HO1252", t0)
	if err != nil {
		t.Fatal(err)
	}
	res, err := resource.Register(state, "cart-1", "vehicle")
	if err != nil {
		t.Fatal(err)
	}
	tk, err := task.Create(state, f.ID, "cleaning")
	if err != nil {
		t.Fatal(err)
	}
	if err := task.Assign(state, tk.ID, "crew-a", res.ID, t0); err != nil {
		t.Fatal(err)
	}
	if err := task.Start(state, tk.ID, t0); err != nil {
		t.Fatal(err)
	}
	if err := flight.SetStatus(state, f.ID, store.FlightDeparted, t0.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := task.Receipt(state, tk.ID, t0.Add(2*time.Hour)); err == nil {
		t.Fatal("receipt accepted for a departed flight")
	}
	after, _ := state.Task(tk.ID)
	if after.Status != store.TaskInProgress {
		t.Fatalf("terminal receipt advanced the task to %s", after.Status)
	}
}
