package verifycase

import (
	"testing"
	"time"

	"groundops/internal/flight"
	"groundops/internal/resource"
	"groundops/internal/store"
	"groundops/internal/task"
)

func TestAssignSkipsTerminalFlights(t *testing.T) {
	dir := t.TempDir()
	state := store.NewState(dir)
	t0 := time.Now().UTC()
	f, err := flight.Register(state, "3U8890", t0)
	if err != nil {
		t.Fatal(err)
	}
	res, err := resource.Register(state, "fuel-1", "vehicle")
	if err != nil {
		t.Fatal(err)
	}
	if err := flight.SetStatus(state, f.ID, store.FlightDeparted, t0.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	tk, err := task.Create(state, f.ID, "fuelling")
	if err != nil {
		t.Fatal(err)
	}
	if err := task.Assign(state, tk.ID, "crew-a", res.ID, t0.Add(time.Hour)); err == nil {
		t.Fatal("assigned a task for a departed flight")
	}
}
