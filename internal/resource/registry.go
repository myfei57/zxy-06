package resource

import (
	"errors"
	"strings"

	"github.com/google/uuid"

	"groundops/internal/store"
)

// ErrUnknownResource is returned for operations on a missing resource.
var ErrUnknownResource = errors.New("resource does not exist")

// Register creates a vehicle or personnel resource.
func Register(state *store.State, name, kind string) (*store.Resource, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("resource name is empty")
	}
	r := &store.Resource{
		ID:     uuid.NewString(),
		Name:   name,
		Kind:   kind,
		Status: store.ResourceIdle,
	}
	if err := state.PutResource(r); err != nil {
		return nil, err
	}
	return r, nil
}

// BoundTask returns the id of the task currently bound to a resource.
func BoundTask(state *store.State, resourceID string) (string, bool) {
	r, ok := state.Resource(resourceID)
	if !ok {
		return "", false
	}
	return r.TaskID, r.TaskID != ""
}

// Available returns idle resources of a kind.
func Available(state *store.State, kind string) []*store.Resource {
	out := make([]*store.Resource, 0)
	for _, r := range state.Resources() {
		if r.Kind == kind && r.Status == store.ResourceIdle {
			out = append(out, r)
		}
	}
	return out
}

// Counts summarizes the resource registry.
func Counts(state *store.State) (total int, idle int, occupied int) {
	for _, r := range state.Resources() {
		total++
		if r.Status == store.ResourceOccupied {
			occupied++
		} else {
			idle++
		}
	}
	return total, idle, occupied
}
