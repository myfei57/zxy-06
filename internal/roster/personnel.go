package roster

import (
	"errors"

	"github.com/google/uuid"

	"groundops/internal/resource"
	"groundops/internal/store"
)

// AddPersonnel binds a personnel resource to a shift.
func AddPersonnel(state *store.State, shiftID, name string) (*store.RosterEntry, error) {
	if _, ok := state.Shift(shiftID); !ok {
		return nil, errors.New("shift does not exist")
	}
	r, err := resource.Register(state, name, "personnel")
	if err != nil {
		return nil, err
	}
	e := &store.RosterEntry{
		ID:         uuid.NewString(),
		ShiftID:    shiftID,
		ResourceID: r.ID,
	}
	if err := state.PutRoster(e); err != nil {
		return nil, err
	}
	return e, nil
}

// PersonnelOf returns the personnel resources of a shift.
func PersonnelOf(state *store.State, shiftID string) []*store.Resource {
	out := make([]*store.Resource, 0)
	for _, e := range state.RosterByShift(shiftID) {
		if r, ok := state.Resource(e.ResourceID); ok {
			out = append(out, r)
		}
	}
	return out
}
