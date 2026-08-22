package task

import "groundops/internal/store"

// OpenForFlight returns the non-terminal tasks of a flight.
func OpenForFlight(state *store.State, flightID string) []*store.Task {
	out := make([]*store.Task, 0)
	for _, t := range state.TasksByFlight(flightID) {
		if IsActive(t) {
			out = append(out, t)
		}
	}
	return out
}

// CompletedCountForFlight counts completed tasks of a flight.
func CompletedCountForFlight(state *store.State, flightID string) int {
	count := 0
	for _, t := range state.TasksByFlight(flightID) {
		if t.Status == store.TaskCompleted {
			count++
		}
	}
	return count
}

// FailedCount counts tasks that ended in failure.
func FailedCount(state *store.State) int {
	count := 0
	for _, t := range state.Tasks() {
		if t.Status == store.TaskFailed {
			count++
		}
	}
	return count
}
