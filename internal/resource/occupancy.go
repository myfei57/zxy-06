package resource

import (
	"errors"

	"groundops/internal/store"
)

// Occupy binds a resource to a task.
func Occupy(state *store.State, resourceID, taskID string) error {
	r, ok := state.Resource(resourceID)
	if !ok {
		return ErrUnknownResource
	}
	if r.Status == store.ResourceOccupied {
		return errors.New("resource is occupied")
	}
	r.Status = store.ResourceOccupied
	r.TaskID = taskID
	return state.PutResource(r)
}
