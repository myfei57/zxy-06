package console

import (
	"encoding/json"
	"net/http"

	"groundops/internal/audit"
	"groundops/internal/sla"
	"groundops/internal/store"
	"groundops/internal/task"
)

func (a *API) ListTasks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.state.Tasks())
}

func (a *API) GetTask(w http.ResponseWriter, r *http.Request) {
	id := chiURL(r, "id")
	t, ok := a.state.Task(id)
	if !ok {
		writeErr(w, http.StatusNotFound, errNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"task":   t,
		"active": task.IsActive(t),
	})
}

func (a *API) CreateTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FlightID string `json:"flight_id"`
		TaskType string `json:"task_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, errBadBody)
		return
	}
	t, err := task.Create(a.state, body.FlightID, body.TaskType)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (a *API) AssignTask(w http.ResponseWriter, r *http.Request) {
	id := chiURL(r, "id")
	var body struct {
		Owner      string `json:"owner"`
		ResourceID string `json:"resource_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, errBadBody)
		return
	}
	if err := task.Assign(a.state, id, body.Owner, body.ResourceID, storeNow()); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	t, _ := a.state.Task(id)
	if err := sla.EnsureTimer(a.state, id, t.FlightID, storeNow()); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	_ = audit.Record(a.state, "task_assign", id, body.Owner, storeNow())
	writeJSON(w, http.StatusOK, map[string]string{"status": "assigned"})
}

func (a *API) StartTask(w http.ResponseWriter, r *http.Request) {
	id := chiURL(r, "id")
	if err := task.Start(a.state, id, storeNow()); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "in_progress"})
}

func (a *API) CompleteTask(w http.ResponseWriter, r *http.Request) {
	id := chiURL(r, "id")
	var body struct {
		Signer string `json:"signer"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, errBadBody)
		return
	}
	result := &store.TaskResult{Signer: body.Signer, Note: body.Note}
	if err := task.Complete(a.state, id, result, storeNow()); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	_ = audit.Record(a.state, "task_complete", id, body.Signer, storeNow())
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (a *API) ReceiptTask(w http.ResponseWriter, r *http.Request) {
	id := chiURL(r, "id")
	if err := task.Receipt(a.state, id, storeNow()); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	_ = audit.Record(a.state, "task_receipt", id, "receipt accepted", storeNow())
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}
