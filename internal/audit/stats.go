package audit

import "groundops/internal/store"

// CompletedResults counts durable task results.
func CompletedResults(state *store.State) int {
	return len(state.Results())
}

// EscalationFired counts audit entries of escalation actions.
func EscalationFired(state *store.State) int {
	count := 0
	for _, e := range state.AuditEntries() {
		if e.Action == "escalation" {
			count++
		}
	}
	return count
}
