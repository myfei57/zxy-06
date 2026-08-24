package console

import (
	"encoding/json"
	"net/http"

	"groundops/internal/audit"
	"groundops/internal/flight"
	"groundops/internal/ops"
	"groundops/internal/store"
	"groundops/internal/task"
)

func (a *API) ListFlights(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.state.Flights())
}

func (a *API) GetFlight(w http.ResponseWriter, r *http.Request) {
	id := chiURL(r, "id")
	f, ok := a.state.Flight(id)
	if !ok {
		writeErr(w, http.StatusNotFound, errNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"flight":  f,
		"servicable": flight.IsServicable(f),
		"terminal": flight.IsTerminal(f),
		"status":  flight.LiveStatus(f),
		"tasks":          a.state.TasksByFlight(id),
		"open_tasks":     task.OpenForFlight(a.state, id),
		"completed_tasks": task.CompletedCountForFlight(a.state, id),
	})
}

func (a *API) RegisterFlight(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FlightNo    string `json:"flight_no"`
		ScheduledAt string `json:"scheduled_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, errBadBody)
		return
	}
	at := storeNow()
	if body.ScheduledAt != "" {
		if parsed, err := parseTime(body.ScheduledAt); err == nil {
			at = parsed
		}
	}
	f, err := flight.Register(a.state, body.FlightNo, at)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	_ = audit.Record(a.state, "flight_create", f.ID, f.FlightNo, storeNow())
	writeJSON(w, http.StatusCreated, f)
}

func (a *API) UpdateFlight(w http.ResponseWriter, r *http.Request) {
	id := chiURL(r, "id")
	var body struct {
		Seq     int64  `json:"seq"`
		Status  string `json:"status"`
		EventAt string `json:"event_at"`
		Note    string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, errBadBody)
		return
	}
	at := storeNow()
	if body.EventAt != "" {
		if parsed, err := parseTime(body.EventAt); err == nil {
			at = parsed
		}
	}
	update := store.FlightUpdate{Seq: body.Seq, Status: body.Status, EventAt: at, Note: body.Note}
	if err := ops.ApplyUpdate(a.state, id, update); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	_ = audit.Record(a.state, "flight_update", id, update.Note, storeNow())
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *API) CancelFlight(w http.ResponseWriter, r *http.Request) {
	id := chiURL(r, "id")
	if err := ops.CancelFlight(a.state, id, storeNow()); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	_ = audit.Record(a.state, "flight_cancel", id, "flight cancelled", storeNow())
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}
