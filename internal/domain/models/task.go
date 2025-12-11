package models

import (
	"time"

	"github.com/google/uuid"
)

// Task represents a task entity
type Task struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// TaskPriority defines task priority levels
type TaskPriority string

const (
	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"
)

// TaskStatus defines task status
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusActive    TaskStatus = "active"
	StatusCompleted TaskStatus = "completed"
	StatusCancelled TaskStatus = "cancelled"
)

// CreateTaskRequest represents request to create a task
type CreateTaskRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

// UpdateTaskRequest represents request to update a task
type UpdateTaskRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	Status      *string `json:"status,omitempty"`
}
