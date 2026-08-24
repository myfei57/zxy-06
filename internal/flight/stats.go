package flight

import "groundops/internal/store"

// StatusCounts groups flights by live status.
func StatusCounts(state *store.State) map[string]int {
	counts := make(map[string]int)
	for _, f := range state.Flights() {
		counts[f.Status]++
	}
	return counts
}

// DelayedFlights returns flights whose latest update is after the scheduled time.
func DelayedFlights(state *store.State) []*store.Flight {
	out := make([]*store.Flight, 0)
	for _, f := range state.Flights() {
		if f.UpdatedAt.After(f.ScheduledAt) && !IsTerminal(f) {
			out = append(out, f)
		}
	}
	return out
}
