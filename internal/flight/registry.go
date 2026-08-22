package flight

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"groundops/internal/store"
)

// Register creates a scheduled flight.
func Register(state *store.State, flightNo string, scheduledAt time.Time) (*store.Flight, error) {
	flightNo = strings.TrimSpace(flightNo)
	if flightNo == "" {
		return nil, errors.New("flight number is empty")
	}
	f := &store.Flight{
		ID:          uuid.NewString(),
		FlightNo:    flightNo,
		Status:      store.FlightScheduled,
		ScheduledAt: scheduledAt,
		UpdatedAt:   scheduledAt,
		UpdatedSeq:  1,
	}
	if err := state.PutFlight(f); err != nil {
		return nil, err
	}
	return f, nil
}
