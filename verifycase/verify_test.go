package verifycase

import (
	"testing"
	"time"

	"groundops/internal/flight"
	"groundops/internal/roster"
	"groundops/internal/resource"
	"groundops/internal/shift"
	"groundops/internal/store"
	"groundops/internal/task"
)

func TestHandoverMigratesInflightOwner(t *testing.T) {
	dir := t.TempDir()
	state := store.NewState(dir)
	t0 := time.Now().UTC()
	f, err := flight.Register(state, "HU7171", t0)
	if err != nil {
		t.Fatal(err)
	}
	res, err := resource.Register(state, "bus-2", "vehicle")
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
	night, err := shift.Create(state, "night", t0)
	if err != nil {
		t.Fatal(err)
	}
	day, err := shift.Create(state, "day", t0.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := roster.AddPersonnel(state, night.ID, "crew-a"); err != nil {
		t.Fatal(err)
	}
	if _, err := roster.AddPersonnel(state, day.ID, "crew-b"); err != nil {
		t.Fatal(err)
	}
	// bind the task to the night shift
	tkAfter, _ := state.Task(tk.ID)
	tkAfter.ShiftID = night.ID
	if err := state.PutTask(tkAfter); err != nil {
		t.Fatal(err)
	}
	if err := shift.Handover(state, night.ID, day.ID); err != nil {
		t.Fatal(err)
	}
	after, _ := state.Task(tk.ID)
	if after.Owner != "crew-b" {
		t.Fatalf("owner %q not migrated to incoming crew", after.Owner)
	}
}
