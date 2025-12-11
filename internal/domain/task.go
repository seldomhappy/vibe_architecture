package domain

import (
	"fmt"
	"strings"
	"time"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Priority represents the priority level of a task
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Task represents a task entity
type Task struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	Priority    Priority   `json:"priority"`
	AssignedTo  *int64     `json:"assigned_to,omitempty"`
	CreatedBy   int64      `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Validate validates the task entity
func (t *Task) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return ErrEmptyTaskName
	}
	if len(t.Name) > 255 {
		return ErrTaskNameTooLong
	}
	if !t.Status.IsValid() {
		return ErrInvalidInput
	}
	if !t.Priority.IsValid() {
		return ErrInvalidInput
	}
	if t.CreatedBy <= 0 {
		return ErrInvalidInput
	}
	return nil
}

// IsCompleted returns true if the task is completed
func (t *Task) IsCompleted() bool {
	return t.Status == TaskStatusCompleted
}

// CanBeAssigned returns true if the task can be assigned to someone
func (t *Task) CanBeAssigned() bool {
	return t.Status == TaskStatusPending || t.Status == TaskStatusInProgress
}

// Complete marks the task as completed
func (t *Task) Complete() error {
	if t.IsCompleted() {
		return fmt.Errorf("task is already completed")
	}
	if t.Status == TaskStatusCancelled {
		return fmt.Errorf("cannot complete a cancelled task")
	}
	t.Status = TaskStatusCompleted
	t.UpdatedAt = time.Now()
	return nil
}

// Assign assigns the task to a user
func (t *Task) Assign(userID int64) error {
	if !t.CanBeAssigned() {
		return fmt.Errorf("task cannot be assigned in its current status: %s", t.Status)
	}
	if userID <= 0 {
		return ErrUserNotFound
	}
	t.AssignedTo = &userID
	if t.Status == TaskStatusPending {
		t.Status = TaskStatusInProgress
	}
	t.UpdatedAt = time.Now()
	return nil
}

// Cancel marks the task as cancelled
func (t *Task) Cancel() error {
	if t.IsCompleted() {
		return fmt.Errorf("cannot cancel a completed task")
	}
	if t.Status == TaskStatusCancelled {
		return fmt.Errorf("task is already cancelled")
	}
	t.Status = TaskStatusCancelled
	t.UpdatedAt = time.Now()
	return nil
}

// IsValid returns true if the status is valid
func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusPending, TaskStatusInProgress, TaskStatusCompleted, TaskStatusCancelled:
		return true
	}
	return false
}

// IsValid returns true if the priority is valid
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	}
	return false
}
