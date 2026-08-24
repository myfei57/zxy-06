package verifycase

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"groundops/internal/flight"
	"groundops/internal/resource"
	"groundops/internal/store"
	"groundops/internal/task"
)

func TestResourceReleasedAfterCompletionDurable(t *testing.T) {
	dir := t.TempDir()
	state := store.NewState(dir)
	t0 := time.Now().UTC()
	f, err := flight.Register(state, "KA831", t0)
	if err != nil {
		t.Fatal(err)
	}
	res, err := resource.Register(state, "truck-1", "vehicle")
	if err != nil {
		t.Fatal(err)
	}
	tk, err := task.Create(state, f.ID, "fuel")
	if err != nil {
		t.Fatal(err)
	}
	if err := task.Assign(state, tk.ID, "crew-a", res.ID, t0); err != nil {
		t.Fatal(err)
	}
	if err := task.Start(state, tk.ID, t0); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "results"), []byte("blocker"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := &store.TaskResult{Signer: "crew-a", Note: "ok"}
	if err := task.Complete(state, tk.ID, result, t0.Add(time.Minute)); err == nil {
		t.Fatal("expected the result write to fail")
	}
	r, _ := state.Resource(res.ID)
	if r.Status != store.ResourceOccupied {
		t.Fatalf("resource released although completion was not durable")
	}
}
