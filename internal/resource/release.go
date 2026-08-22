package resource

import (
	"errors"

	"groundops/internal/store"
)

// ErrAlreadyIdle is returned when releasing an idle resource.
var ErrAlreadyIdle = errors.New("resource is already idle")

// Release frees a resource back to the idle pool.
func Release(state *store.State, resourceID string) error {
	r, ok := state.Resource(resourceID)
	if !ok {
		return ErrUnknownResource
	}
	if r.Status == store.ResourceIdle {
		return ErrAlreadyIdle
	}
	r.Status = store.ResourceIdle
	r.TaskID = ""
	return state.PutResource(r)
}
