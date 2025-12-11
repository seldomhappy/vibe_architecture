package domain

import "time"

// EventType represents the type of domain event
type EventType string

const (
	EventTypeTaskCreated   EventType = "task.created"
	EventTypeTaskUpdated   EventType = "task.updated"
	EventTypeTaskCompleted EventType = "task.completed"
	EventTypeTaskDeleted   EventType = "task.deleted"
)

// TaskCreatedEvent is published when a task is created
type TaskCreatedEvent struct {
	TaskID      int64      `json:"task_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Priority    Priority   `json:"priority"`
	CreatedBy   int64      `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
}

// TaskUpdatedEvent is published when a task is updated
type TaskUpdatedEvent struct {
	TaskID      int64      `json:"task_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	Priority    Priority   `json:"priority"`
	AssignedTo  *int64     `json:"assigned_to,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TaskCompletedEvent is published when a task is completed
type TaskCompletedEvent struct {
	TaskID      int64     `json:"task_id"`
	CompletedAt time.Time `json:"completed_at"`
}

// TaskDeletedEvent is published when a task is deleted
type TaskDeletedEvent struct {
	TaskID    int64     `json:"task_id"`
	DeletedAt time.Time `json:"deleted_at"`
}
