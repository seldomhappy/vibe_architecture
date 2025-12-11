package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/seldomhappy/vibe_architecture/internal/domain"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/postgres"
	"github.com/seldomhappy/vibe_architecture/internal/pkg/tracing"
	"github.com/seldomhappy/vibe_architecture/logger"
	"go.opentelemetry.io/otel/attribute"
)

// TaskRepository implements task data access
type TaskRepository struct {
	db     *postgres.DB
	logger logger.ILogger
}

// TaskFilter represents filters for listing tasks
type TaskFilter struct {
	Status     *domain.TaskStatus
	Priority   *domain.Priority
	AssignedTo *int64
	Limit      int
	Offset     int
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *postgres.DB, log logger.ILogger) *TaskRepository {
	return &TaskRepository{
		db:     db,
		logger: log,
	}
}

// Create creates a new task
func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	ctx, span := tracing.StartSpan(ctx, "repository", "create_task")
	defer span.End()

	span.SetAttributes(
		attribute.String("task.name", task.Name),
		attribute.String("task.priority", string(task.Priority)),
	)

	query := `
		INSERT INTO tasks (name, description, status, priority, assigned_to, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		task.Name,
		task.Description,
		task.Status,
		task.Priority,
		task.AssignedTo,
		task.CreatedBy,
		now,
		now,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		r.logger.Error("Failed to create task: %v", err)
		tracing.RecordError(ctx, err)
		return fmt.Errorf("failed to create task: %w", err)
	}

	r.logger.Debug("Task created with ID: %d", task.ID)
	return nil
}

// GetByID retrieves a task by ID
func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	ctx, span := tracing.StartSpan(ctx, "repository", "get_task_by_id")
	defer span.End()

	span.SetAttributes(attribute.Int64("task.id", id))

	query := `
		SELECT id, name, description, status, priority, assigned_to, created_by, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	task := &domain.Task{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.Name,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.AssignedTo,
		&task.CreatedBy,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		r.logger.Error("Failed to get task by ID: %v", err)
		tracing.RecordError(ctx, err)
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// GetAll retrieves all tasks with optional filters
func (r *TaskRepository) GetAll(ctx context.Context, filter TaskFilter) ([]*domain.Task, error) {
	ctx, span := tracing.StartSpan(ctx, "repository", "get_all_tasks")
	defer span.End()

	query := `
		SELECT id, name, description, status, priority, assigned_to, created_by, created_at, updated_at
		FROM tasks
		WHERE 1=1
	`
	args := make([]interface{}, 0)
	argCount := 1

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filter.Status)
		argCount++
	}

	if filter.Priority != nil {
		query += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, *filter.Priority)
		argCount++
	}

	if filter.AssignedTo != nil {
		query += fmt.Sprintf(" AND assigned_to = $%d", argCount)
		args = append(args, *filter.AssignedTo)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to get all tasks: %v", err)
		tracing.RecordError(ctx, err)
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		task := &domain.Task{}
		err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.AssignedTo,
			&task.CreatedBy,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan task: %v", err)
			continue
		}
		tasks = append(tasks, task)
	}

	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))
	return tasks, nil
}

// Update updates an existing task
func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	ctx, span := tracing.StartSpan(ctx, "repository", "update_task")
	defer span.End()

	span.SetAttributes(attribute.Int64("task.id", task.ID))

	query := `
		UPDATE tasks
		SET name = $1, description = $2, status = $3, priority = $4, assigned_to = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := r.db.Pool().Exec(ctx, query,
		task.Name,
		task.Description,
		task.Status,
		task.Priority,
		task.AssignedTo,
		time.Now(),
		task.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update task: %v", err)
		tracing.RecordError(ctx, err)
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

// Delete deletes a task
func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	ctx, span := tracing.StartSpan(ctx, "repository", "delete_task")
	defer span.End()

	span.SetAttributes(attribute.Int64("task.id", id))

	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.Pool().Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete task: %v", err)
		tracing.RecordError(ctx, err)
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}
