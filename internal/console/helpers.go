package console

import (
	"encoding/json"
	"net/http"

	"groundops/internal/audit"
	"groundops/internal/flight"
	"groundops/internal/recover"
	"groundops/internal/resource"
	"groundops/internal/roster"
	"groundops/internal/shift"
	"groundops/internal/sla"
	"groundops/internal/task"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func (a *API) Stats(w http.ResponseWriter, r *http.Request) {
	totalF, activeF, terminalF := flight.Counts(a.state)
	totalT, activeT, completedT := task.Counts(a.state)
	totalR, idleR, occupiedR := resource.Counts(a.state)
	shiftName := ""
	if sh, ok := shift.ActiveShift(a.state); ok {
		shiftName = sh.Name
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"flights":        totalF,
		"active_flights": activeF,
		"terminal_flights": terminalF,
		"tasks":          totalT,
		"active_tasks":   activeT,
		"completed_tasks": completedT,
		"resources":      totalR,
		"idle_resources": idleR,
		"occupied_resources": occupiedR,
		"delayed_flights": len(flight.DelayedFlights(a.state)),
		"failed_tasks":    task.FailedCount(a.state),
		"flight_statuses": flight.StatusCounts(a.state),
		"shift_name":      shiftName,
		"completed_results": audit.CompletedResults(a.state),
		"escalations_fired": audit.EscalationFired(a.state),
		"current_shift":   roster.CurrentShiftID(a.state),
		"sla_minutes":     sla.LevelMinutes,
		"max_level":       sla.MaxLevel,
	})
}

func (a *API) RecoverSweep(w http.ResponseWriter, r *http.Request) {
	count, err := recover.Sweep(a.state, storeNow())
	if err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"redispatched": count})
}

func (a *API) ListShifts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.state.Shifts())
}

func (a *API) CreateShift(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, errBadBody)
		return
	}
	sh, err := shift.Create(a.state, body.Name, storeNow())
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, sh)
}

func (a *API) CloseShift(w http.ResponseWriter, r *http.Request) {
	id := chiURL(r, "id")
	if err := shift.Close(a.state, id, storeNow()); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}

func (a *API) HandoverShift(w http.ResponseWriter, r *http.Request) {
	var body struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, errBadBody)
		return
	}
	if err := shift.Handover(a.state, body.From, body.To); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "handover"})
}

func (a *API) ListRoster(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.state.RosterEntries())
}

func (a *API) AddRosterPersonnel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ShiftID string `json:"shift_id"`
		Name    string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, errBadBody)
		return
	}
	e, err := roster.AddPersonnel(a.state, body.ShiftID, body.Name)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (a *API) PlannerPool(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"pool": task.PlannerPool(a.state)})
}

func (a *API) ListEscalations(w http.ResponseWriter, r *http.Request) {
	out := make([]map[string]any, 0)
	now := storeNow()
	for _, e := range a.state.Escalations() {
		out = append(out, map[string]any{
			"escalation": e,
			"remaining":  sla.Remaining(e, now).String(),
			"level":      sla.LevelLabel(e.Level),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) SlaTick(w http.ResponseWriter, r *http.Request) {
	fired, err := sla.Tick(a.state, storeNow())
	if err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"fired": fired})
}

func (a *API) ListAudit(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.state.AuditEntries())
}

func (a *API) GetAuditEntry(w http.ResponseWriter, r *http.Request) {
	id := chiURL(r, "id")
	for _, e := range a.state.AuditEntries() {
		if e.ID == id {
			writeJSON(w, http.StatusOK, e)
			return
		}
	}
	writeErr(w, http.StatusNotFound, errNotFound)
}
