package verifycase

import (
	"testing"
	"time"

	"groundops/internal/roster"
	"groundops/internal/shift"
	"groundops/internal/store"
	"groundops/internal/task"
)

func TestPlannerUsesCurrentRoster(t *testing.T) {
	dir := t.TempDir()
	state := store.NewState(dir)
	t0 := time.Now().UTC()
	night, err := shift.Create(state, "night", t0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := roster.AddPersonnel(state, night.ID, "crew-a"); err != nil {
		t.Fatal(err)
	}
	_ = roster.CurrentShiftID(state)
	day, err := shift.Create(state, "day", t0.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := roster.AddPersonnel(state, day.ID, "crew-b"); err != nil {
		t.Fatal(err)
	}
	pool := task.PlannerPool(state)
	hasA, hasB := false, false
	for _, name := range pool {
		if name == "crew-a" {
			hasA = true
		}
		if name == "crew-b" {
			hasB = true
		}
	}
	if hasA || !hasB {
		t.Fatalf("planner used a stale roster pool: %v", pool)
	}
}
