package console

import (
	"strconv"

	"groundops/internal/flight"
	"groundops/internal/store"
	"groundops/internal/task"
)

func formatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

func flightCounts(state *store.State) (int, int, int) {
	return flight.Counts(state)
}

func taskCounts(state *store.State) (int, int, int) {
	return task.Counts(state)
}
