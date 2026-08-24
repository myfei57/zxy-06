package flight

import "groundops/internal/store"

// IsServicable reports whether a flight may still receive ground tasks.
func IsServicable(f *store.Flight) bool {
	if f == nil {
		return false
	}
	switch f.Status {
	case store.FlightScheduled, store.FlightArrived, store.FlightServicing, store.FlightReady:
		return true
	}
	return false
}

// IsTerminal reports whether a flight reached a terminal state.
func IsTerminal(f *store.Flight) bool {
	if f == nil {
		return true
	}
	return f.Status == store.FlightDeparted || f.Status == store.FlightCancelled
}

// LiveStatus returns the flight status.
func LiveStatus(f *store.Flight) string {
	if f == nil {
		return ""
	}
	return f.Status
}

// Counts summarizes the flight registry.
func Counts(state *store.State) (total int, active int, terminal int) {
	for _, f := range state.Flights() {
		total++
		if IsTerminal(f) {
			terminal++
		} else {
			active++
		}
	}
	return total, active, terminal
}
