package shift

import (
	"errors"

	"groundops/internal/roster"
	"groundops/internal/store"
)

// Handover migrates in-flight tasks of the outgoing shift to the incoming
// shift, including the executing owner and the bound resource.
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
		t.ShiftID = toID
		if err := state.PutTask(t); err != nil {
			return err
		}
	}
	return nil
}
