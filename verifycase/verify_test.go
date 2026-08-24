package verifycase

import (
	"testing"
	"time"

	"groundops/internal/flight"
	"groundops/internal/sla"
	"groundops/internal/store"
	"groundops/internal/task"
)

func TestSlaRebasesOnLatestUpdate(t *testing.T) {
	dir := t.TempDir()
	state := store.NewState(dir)
	t0 := time.Now().UTC()
	t1 := t0.Add(2 * time.Hour)
	f, err := flight.Register(state, "CA1821", t0)
	if err != nil {
		t.Fatal(err)
	}
	if err := flight.Update(state, f.ID, store.FlightUpdate{Seq: 2, EventAt: t1}); err != nil {
		t.Fatal(err)
	}
	tk, err := task.Create(state, f.ID, "cleaning")
	if err != nil {
		t.Fatal(err)
	}
	if err := sla.EnsureTimer(state, tk.ID, f.ID, t0); err != nil {
		t.Fatal(err)
	}
	e, ok := state.Escalation(tk.ID)
	if !ok {
		t.Fatal("escalation timer missing")
	}
	want := t1.Add(30 * time.Minute)
	if !e.DueAt.Equal(want) {
		t.Fatalf("due %v, want %v (latest update based)", e.DueAt, want)
	}
}
