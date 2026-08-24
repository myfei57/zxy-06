package store

import "sort"

// TasksByFlight returns tasks of a flight ordered by creation time.
func (s *State) TasksByFlight(flightID string) []*Task {
	out := make([]*Task, 0)
	for _, t := range s.Tasks() {
		if t.FlightID == flightID {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

// TasksByResource returns tasks bound to a resource ordered by creation time.
func (s *State) TasksByResource(resourceID string) []*Task {
	out := make([]*Task, 0)
	for _, t := range s.Tasks() {
		if t.ResourceID == resourceID {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

// RosterByShift returns roster entries of one shift.
func (s *State) RosterByShift(shiftID string) []*RosterEntry {
	out := make([]*RosterEntry, 0)
	for _, e := range s.RosterEntries() {
		if e.ShiftID == shiftID {
			out = append(out, e)
		}
	}
	return out
}

// InProgressTasks returns tasks currently in progress.
func (s *State) InProgressTasks() []*Task {
	out := make([]*Task, 0)
	for _, t := range s.Tasks() {
		if t.Status == TaskInProgress {
			out = append(out, t)
		}
	}
	return out
}

// PendingOrActiveTasks returns tasks that are not terminal.
func (s *State) PendingOrActiveTasks() []*Task {
	out := make([]*Task, 0)
	for _, t := range s.Tasks() {
		if t.Status == TaskPending || t.Status == TaskAssigned || t.Status == TaskInProgress {
			out = append(out, t)
		}
	}
	return out
}

// RecentAudit returns the most recent audit entries.
func (s *State) RecentAudit(n int) []*AuditEntry {
	all := s.AuditEntries()
	if n <= 0 || n > len(all) {
		n = len(all)
	}
	out := make([]*AuditEntry, n)
	for i := 0; i < n; i++ {
		out[i] = all[len(all)-1-i]
	}
	return out
}
