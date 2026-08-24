package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"groundops/internal/store"
)

// API exposes the ground operations control plane over HTTP.
type API struct {
	state *store.State
}

// NewAPI builds the HTTP handler for the control plane.
func NewAPI(state *store.State) http.Handler {
	api := &API{state: state}
	r := chi.NewRouter()
	r.Get("/", api.IndexPage)
	r.Get("/console/flights", api.FlightsPage)
	r.Get("/console/tasks", api.TasksPage)
	r.Get("/console/resources", api.ResourcesPage)
	r.Get("/console/audit", api.AuditPage)

	r.Get("/api/stats", api.Stats)
	r.Get("/api/flights", api.ListFlights)
	r.Get("/api/flights/{id}", api.GetFlight)
	r.Post("/api/flights", api.RegisterFlight)
	r.Post("/api/flights/{id}/update", api.UpdateFlight)
	r.Post("/api/flights/{id}/cancel", api.CancelFlight)

	r.Get("/api/tasks", api.ListTasks)
	r.Get("/api/tasks/{id}", api.GetTask)
	r.Post("/api/tasks", api.CreateTask)
	r.Post("/api/tasks/{id}/assign", api.AssignTask)
	r.Post("/api/tasks/{id}/start", api.StartTask)
	r.Post("/api/tasks/{id}/complete", api.CompleteTask)
	r.Post("/api/tasks/{id}/receipt", api.ReceiptTask)

	r.Get("/api/resources", api.ListResources)
	r.Get("/api/resources/{id}", api.GetResource)
	r.Post("/api/resources", api.RegisterResource)
	r.Post("/api/resources/{id}/release", api.ReleaseResource)

	r.Get("/api/shifts", api.ListShifts)
	r.Post("/api/shifts", api.CreateShift)
	r.Post("/api/shifts/{id}/close", api.CloseShift)
	r.Post("/api/shifts/handover", api.HandoverShift)

	r.Get("/api/roster", api.ListRoster)
	r.Post("/api/roster", api.AddRosterPersonnel)
	r.Get("/api/planner", api.PlannerPool)

	r.Post("/api/recover/sweep", api.RecoverSweep)
	r.Get("/api/escalations", api.ListEscalations)
	r.Post("/api/sla/tick", api.SlaTick)

	r.Get("/api/audit", api.ListAudit)
	r.Get("/api/audit/{id}", api.GetAuditEntry)
	return r
}
