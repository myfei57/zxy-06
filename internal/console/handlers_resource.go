package console

import (
	"encoding/json"
	"net/http"

	"groundops/internal/resource"
)

func (a *API) ListResources(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"resources":         a.state.Resources(),
		"available_vehicles": resource.Available(a.state, "vehicle"),
		"available_personnel": resource.Available(a.state, "personnel"),
	})
}

func (a *API) GetResource(w http.ResponseWriter, r *http.Request) {
	id := chiURL(r, "id")
	rs, ok := a.state.Resource(id)
	if !ok {
		writeErr(w, http.StatusNotFound, errNotFound)
		return
	}
	taskID, bound := resource.BoundTask(a.state, id)
	writeJSON(w, http.StatusOK, map[string]any{"resource": rs, "bound_task": taskID, "bound": bound})
}

func (a *API) RegisterResource(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
		Kind string `json:"kind"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, errBadBody)
		return
	}
	rs, err := resource.Register(a.state, body.Name, body.Kind)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, rs)
}

func (a *API) ReleaseResource(w http.ResponseWriter, r *http.Request) {
	id := chiURL(r, "id")
	if err := resource.Release(a.state, id); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "idle"})
}

