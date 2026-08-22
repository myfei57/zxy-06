package shift

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"groundops/internal/store"
)

// Create opens a new shift.
func Create(state *store.State, name string, at time.Time) (*store.Shift, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("shift name is empty")
	}
	sh := &store.Shift{
		ID:        uuid.NewString(),
		Name:      name,
		StartedAt: at,
	}
	if err := state.PutShift(sh); err != nil {
		return nil, err
	}
	return sh, nil
}

// ActiveShift returns the most recently opened shift.
func ActiveShift(state *store.State) (*store.Shift, bool) {
	shifts := state.Shifts()
	if len(shifts) == 0 {
		return nil, false
	}
	return shifts[len(shifts)-1], true
}

// Close ends a shift.
func Close(state *store.State, shiftID string, at time.Time) error {
	sh, ok := state.Shift(shiftID)
	if !ok {
		return errors.New("shift does not exist")
	}
	sh.EndedAt = at
	return state.PutShift(sh)
}
