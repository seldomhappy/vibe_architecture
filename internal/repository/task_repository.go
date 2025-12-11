package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/seldomhappy/vibe_architecture/internal/domain/models"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/logger"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/metrics"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/postgres"
)

// TaskRepository implements domain task repository
type TaskRepository struct {
	db      *postgres.DB
	log     logger.ILogger
	metrics *metrics.Metrics
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *postgres.DB, log logger.ILogger, m *metrics.Metrics) *TaskRepository {
	return &TaskRepository{
		db:      db,
		log:     log,
		metrics: m,
	}
}

// Create creates a new task
func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tasks (id, name, description, priority, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		task.ID,
		task.Name,
		task.Description,
		task.Priority,
		task.Status,
		task.CreatedAt,
		task.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	r.log.Debug("Task created: %s", task.ID)
	return nil
}

// GetByID retrieves a task by ID
func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	query := `
		SELECT id, name, description, priority, status, created_at, updated_at, completed_at
		FROM tasks
		WHERE id = $1
	`

	var task models.Task
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.Name,
		&task.Description,
		&task.Priority,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.CompletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

// List retrieves a list of tasks
func (r *TaskRepository) List(ctx context.Context, limit, offset int) ([]*models.Task, error) {
	query := `
		SELECT id, name, description, priority, status, created_at, updated_at, completed_at
		FROM tasks
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.Description,
			&task.Priority,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	return tasks, nil
}

// Update updates a task
func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	query := `
		UPDATE tasks
		SET name = $2, description = $3, priority = $4, status = $5, updated_at = $6, completed_at = $7
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		task.ID,
		task.Name,
		task.Description,
		task.Priority,
		task.Status,
		task.UpdatedAt,
		task.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}

	r.log.Debug("Task updated: %s", task.ID)
	return nil
}

// Delete deletes a task
func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}

	r.log.Debug("Task deleted: %s", id)
	return nil
}
