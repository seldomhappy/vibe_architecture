package task

import (
	"context"

	"github.com/seldomhappy/vibe_architecture/internal/domain"
	"github.com/seldomhappy/vibe_architecture/internal/repository"
)

// Repository defines the task repository interface
type Repository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id int64) (*domain.Task, error)
	GetAll(ctx context.Context, filter repository.TaskFilter) ([]*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id int64) error
}

// UseCase defines the task use case interface
type UseCase interface {
	CreateTask(ctx context.Context, input CreateTaskInput) (*domain.Task, error)
	GetTask(ctx context.Context, id int64) (*domain.Task, error)
	ListTasks(ctx context.Context, filter ListTasksFilter) ([]*domain.Task, error)
	UpdateTask(ctx context.Context, id int64, input UpdateTaskInput) (*domain.Task, error)
	DeleteTask(ctx context.Context, id int64) error
	AssignTask(ctx context.Context, taskID, userID int64) error
	CompleteTask(ctx context.Context, id int64) error
}

// CreateTaskInput represents input for creating a task
type CreateTaskInput struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Priority    domain.Priority `json:"priority"`
	CreatedBy   int64           `json:"created_by"`
}

// UpdateTaskInput represents input for updating a task
type UpdateTaskInput struct {
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Status      *domain.TaskStatus `json:"status,omitempty"`
	Priority    *domain.Priority   `json:"priority,omitempty"`
}

// ListTasksFilter represents filters for listing tasks
type ListTasksFilter struct {
	Status     *domain.TaskStatus
	Priority   *domain.Priority
	AssignedTo *int64
	Limit      int
	Offset     int
}
