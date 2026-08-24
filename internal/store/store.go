package store

import (
	"path/filepath"
	"sort"
	"sync"
)

// State is the central in-memory plus file-backed registry.
type State struct {
	mu          sync.RWMutex
	root        string
	flights     map[string]*Flight
	tasks       map[string]*Task
	results     map[string]*TaskResult
	resources   map[string]*Resource
	shifts      map[string]*Shift
	roster      map[string]*RosterEntry
	escalations map[string]*Escalation
	audit       map[string]*AuditEntry
}

// NewState creates a State rooted at dir.
func NewState(dir string) *State {
	return &State{
		root:        dir,
		flights:     make(map[string]*Flight),
		tasks:       make(map[string]*Task),
		results:     make(map[string]*TaskResult),
		resources:   make(map[string]*Resource),
		shifts:      make(map[string]*Shift),
		roster:      make(map[string]*RosterEntry),
		escalations: make(map[string]*Escalation),
		audit:       make(map[string]*AuditEntry),
	}
}

// Root returns the state root directory.
func (s *State) Root() string {
	return s.root
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (s *State) PutFlight(f *Flight) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flights[f.ID] = f
	return SaveJSON(filepath.Join(s.root, "flights", f.ID+".json"), f)
}

func (s *State) Flight(id string) (*Flight, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flights[id]
	return f, ok
}

func (s *State) Flights() []*Flight {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Flight, 0, len(s.flights))
	for _, k := range sortedKeys(s.flights) {
		out = append(out, s.flights[k])
	}
	return out
}

func (s *State) PutTask(t *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.ID] = t
	return SaveJSON(filepath.Join(s.root, "tasks", t.ID+".json"), t)
}

func (s *State) Task(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *State) Tasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Task, 0, len(s.tasks))
	for _, k := range sortedKeys(s.tasks) {
		out = append(out, s.tasks[k])
	}
	return out
}

func (s *State) PutResult(r *TaskResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[r.ID] = r
	return SaveJSON(filepath.Join(s.root, "results", r.ID+".json"), r)
}

func (s *State) Result(id string) (*TaskResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.results[id]
	return r, ok
}

func (s *State) Results() []*TaskResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*TaskResult, 0, len(s.results))
	for _, k := range sortedKeys(s.results) {
		out = append(out, s.results[k])
	}
	return out
}

func (s *State) PutResource(r *Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[r.ID] = r
	return SaveJSON(filepath.Join(s.root, "resources", r.ID+".json"), r)
}

func (s *State) Resource(id string) (*Resource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.resources[id]
	return r, ok
}

func (s *State) Resources() []*Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Resource, 0, len(s.resources))
	for _, k := range sortedKeys(s.resources) {
		out = append(out, s.resources[k])
	}
	return out
}

func (s *State) PutShift(sh *Shift) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shifts[sh.ID] = sh
	return SaveJSON(filepath.Join(s.root, "shifts", sh.ID+".json"), sh)
}

func (s *State) Shift(id string) (*Shift, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sh, ok := s.shifts[id]
	return sh, ok
}

func (s *State) Shifts() []*Shift {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Shift, 0, len(s.shifts))
	for _, k := range sortedKeys(s.shifts) {
		out = append(out, s.shifts[k])
	}
	return out
}

func (s *State) PutRoster(e *RosterEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.roster[e.ID] = e
	return SaveJSON(filepath.Join(s.root, "roster", e.ID+".json"), e)
}

func (s *State) RosterEntries() []*RosterEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*RosterEntry, 0, len(s.roster))
	for _, k := range sortedKeys(s.roster) {
		out = append(out, s.roster[k])
	}
	return out
}

func (s *State) PutEscalation(e *Escalation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.escalations[e.TaskID] = e
	return SaveJSON(filepath.Join(s.root, "escalations", e.TaskID+".json"), e)
}

func (s *State) Escalation(taskID string) (*Escalation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.escalations[taskID]
	return e, ok
}

func (s *State) Escalations() []*Escalation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Escalation, 0, len(s.escalations))
	for _, k := range sortedKeys(s.escalations) {
		out = append(out, s.escalations[k])
	}
	return out
}

func (s *State) AppendAudit(e *AuditEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audit[e.ID] = e
	return SaveJSON(filepath.Join(s.root, "audit", e.ID+".json"), e)
}

func (s *State) AuditEntries() []*AuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*AuditEntry, 0, len(s.audit))
	for _, k := range sortedKeys(s.audit) {
		out = append(out, s.audit[k])
	}
	return out
}
