package shift

import (
	"errors"

	"groundops/internal/resource"
	"groundops/internal/roster"
	"groundops/internal/store"
)

// Handover migrates in-flight tasks of the outgoing shift to the incoming
// shift, including the executing owner and the bound resource. The crew and
// vehicle bound to each task are handed over together so that confirmation and
// exception reporting stay accountable to the shift that actually took over.
func Handover(state *store.State, fromID, toID string) error {
	if _, ok := state.Shift(fromID); !ok {
		return errors.New("from shift does not exist")
	}
	if _, ok := state.Shift(toID); !ok {
		return errors.New("to shift does not exist")
	}
	personnel := roster.PersonnelOf(state, toID)
	if len(personnel) == 0 {
		return errors.New("incoming shift has no personnel")
	}
	for _, t := range state.InProgressTasks() {
		if t.ShiftID != fromID {
			continue
		}
		// Release the binding held by the outgoing crew before rebinding the
		// task to the incoming crew, so the resource is not left occupied by a
		// shift that has already handed over.
		if t.ResourceID != "" {
			if r, ok := state.Resource(t.ResourceID); ok && r.Status == store.ResourceOccupied {
				_ = resource.Release(state, t.ResourceID)
			}
		}
		// Bind an idle crew member of the incoming shift as the new executing
		// owner, so the task is carried over together with its owner and
		// bound resource rather than the shift field alone.
		crew, ok := idleCrewMember(state, personnel)
		if !ok {
			return errors.New("incoming shift has no available crew")
		}
		if err := resource.Occupy(state, crew.ID, t.ID); err != nil {
			return err
		}
		t.ShiftID = toID
		t.Owner = crew.Name
		t.ResourceID = crew.ID
		if err := state.PutTask(t); err != nil {
			return err
		}
	}
	return nil
}

// idleCrewMember returns the first idle personnel resource of a shift.
func idleCrewMember(state *store.State, personnel []*store.Resource) (*store.Resource, bool) {
	for _, p := range personnel {
		if r, ok := state.Resource(p.ID); ok && r.Status == store.ResourceIdle {
			return r, true
		}
	}
	return nil, false
}
