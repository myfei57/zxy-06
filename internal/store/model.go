package store

import "time"

// State machine constants shared across the control plane.
const (
	FlightScheduled = "scheduled"
	FlightArrived   = "arrived"
	FlightServicing = "servicing"
	FlightReady     = "ready"
	FlightDeparted  = "departed"
	FlightCancelled = "cancelled"

	TaskPending    = "pending"
	TaskAssigned   = "assigned"
	TaskInProgress = "in_progress"
	TaskCompleted  = "completed"
	TaskFailed     = "failed"
	TaskCancelled  = "cancelled"

	ResourceIdle     = "idle"
	ResourceOccupied = "occupied"

	EscalationPending = "pending"
	EscalationDone    = "done"
)

// Flight is one scheduled or active flight.
type Flight struct {
	ID          string    `json:"id"`
	FlightNo    string    `json:"flight_no"`
	Status      string    `json:"status"`
	ScheduledAt time.Time `json:"scheduled_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedSeq  int64     `json:"updated_seq"`
}

// FlightUpdate is one dynamics message for a flight.
type FlightUpdate struct {
	Seq     int64     `json:"seq"`
	Status  string    `json:"status"`
	EventAt time.Time `json:"event_at"`
	Note    string    `json:"note"`
}

// Task is one ground service task.
type Task struct {
	ID          string    `json:"id"`
	FlightID    string    `json:"flight_id"`
	TaskType    string    `json:"task_type"`
	Status      string    `json:"status"`
	Owner       string    `json:"owner"`
	ShiftID     string    `json:"shift_id"`
	ResourceID  string    `json:"resource_id"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// TaskResult is the durable work result of a completed task.
type TaskResult struct {
	ID     string    `json:"id"`
	TaskID string    `json:"task_id"`
	Signer string    `json:"signer"`
	Note   string    `json:"note"`
	At     time.Time `json:"at"`
}

// Resource is a vehicle or a person that can be assigned to a task.
type Resource struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Status   string `json:"status"`
	TaskID   string `json:"task_id"`
}

// Shift is one working shift of a ground crew.
type Shift struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

// RosterEntry binds a personnel resource to a shift.
type RosterEntry struct {
	ID         string `json:"id"`
	ShiftID    string `json:"shift_id"`
	ResourceID string `json:"resource_id"`
}

// Escalation is the SLA timer state of one task.
type Escalation struct {
	TaskID  string    `json:"task_id"`
	Level   int       `json:"level"`
	DueAt   time.Time `json:"due_at"`
	State   string    `json:"state"`
}

// AuditEntry records one control-plane operation.
type AuditEntry struct {
	ID     string    `json:"id"`
	Action string    `json:"action"`
	Target string    `json:"target"`
	Detail string    `json:"detail"`
	At     time.Time `json:"at"`
}
