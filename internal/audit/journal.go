package audit

import (
	"time"

	"github.com/google/uuid"

	"groundops/internal/store"
)

// Record durably appends one audit entry.
func Record(state *store.State, action, target, detail string, at time.Time) error {
	e := &store.AuditEntry{
		ID:     uuid.NewString(),
		Action: action,
		Target: target,
		Detail: detail,
		At:     at,
	}
	return state.AppendAudit(e)
}

// RecordResult durably stores a completed task's work result.
func RecordResult(state *store.State, result *store.TaskResult) error {
	return state.PutResult(result)
}
