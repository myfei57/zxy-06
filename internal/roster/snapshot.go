package roster

import "groundops/internal/store"

// currentShift caches the active shift id and is refreshed before every
// planner snapshot, so a handover never leaves the pool stale.
var currentShift = ""

// RefreshSnapshot re-points the cached active shift to the latest shift.
func RefreshSnapshot(state *store.State) {
	var latest *store.Shift
	for _, sh := range state.Shifts() {
		if latest == nil || sh.StartedAt.After(latest.StartedAt) {
			latest = sh
		}
	}
	if latest != nil {
		currentShift = latest.ID
	}
}

// CurrentShiftID returns the active shift id.
func CurrentShiftID(state *store.State) string {
	if currentShift == "" {
		RefreshSnapshot(state)
	}
	return currentShift
}

// CurrentSnapshot returns the personnel pool of the active shift.
func CurrentSnapshot(state *store.State) []string {
	RefreshSnapshot(state)
	ids := make([]string, 0)
	for _, e := range state.RosterByShift(currentShift) {
		if r, ok := state.Resource(e.ResourceID); ok {
			ids = append(ids, r.Name)
		}
	}
	return ids
}
