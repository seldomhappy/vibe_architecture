package task

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/seldomhappy/vibe_architecture/internal/domain/models"
	"github.com/seldomhappy/vibe_architecture/internal/domain/repository"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/kafka"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/logger"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/metrics"
)

// UseCase represents task use case
type UseCase struct {
	taskRepo      repository.TaskRepository
	txManager     repository.TxManager
	log           logger.ILogger
	metrics       *metrics.Metrics
	kafkaProducer *kafka.Producer
}

// NewUseCase creates a new task use case
func NewUseCase(
	taskRepo repository.TaskRepository,
	txManager repository.TxManager,
	log logger.ILogger,
	m *metrics.Metrics,
	kafkaProducer *kafka.Producer,
) *UseCase {
	return &UseCase{
		taskRepo:      taskRepo,
		txManager:     txManager,
		log:           log,
		metrics:       m,
		kafkaProducer: kafkaProducer,
	}
}

// CreateTask creates a new task
func (uc *UseCase) CreateTask(ctx context.Context, req models.CreateTaskRequest) (*models.Task, error) {
	task := &models.Task{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      string(models.StatusPending),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := uc.taskRepo.Create(ctx, task); err != nil {
			return err
		}

		// Send Kafka event (safe to call even if producer is disabled)
		if uc.kafkaProducer != nil {
			if err := uc.kafkaProducer.SendTaskCreated(ctx, task); err != nil {
				uc.log.Warn("Failed to send task created event: %v", err)
				// Don't fail the operation if Kafka is unavailable
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	uc.log.Info("Task created: %s", task.ID)
	return task, nil
}

// GetTask retrieves a task by ID
func (uc *UseCase) GetTask(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	task, err := uc.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// ListTasks retrieves a list of tasks
func (uc *UseCase) ListTasks(ctx context.Context, limit, offset int) ([]*models.Task, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	tasks, err := uc.taskRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	return tasks, nil
}

// UpdateTask updates a task
func (uc *UseCase) UpdateTask(ctx context.Context, id uuid.UUID, req models.UpdateTaskRequest) (*models.Task, error) {
	task, err := uc.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		task.Name = *req.Name
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Status != nil {
		task.Status = *req.Status
		if *req.Status == string(models.StatusCompleted) {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
	task.UpdatedAt = time.Now()

	err = uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := uc.taskRepo.Update(ctx, task); err != nil {
			return err
		}

		// Send Kafka event (safe to call even if producer is disabled)
		if uc.kafkaProducer != nil {
			if err := uc.kafkaProducer.SendTaskUpdated(ctx, task); err != nil {
				uc.log.Warn("Failed to send task updated event: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	uc.log.Info("Task updated: %s", task.ID)
	return task, nil
}

// DeleteTask deletes a task
func (uc *UseCase) DeleteTask(ctx context.Context, id uuid.UUID) error {
	err := uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := uc.taskRepo.Delete(ctx, id); err != nil {
			return err
		}

		// Send Kafka event (safe to call even if producer is disabled)
		if uc.kafkaProducer != nil {
			event := map[string]interface{}{"id": id.String()}
			if err := uc.kafkaProducer.SendTaskDeleted(ctx, event); err != nil {
				uc.log.Warn("Failed to send task deleted event: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	uc.log.Info("Task deleted: %s", id)
	return nil
}

// CompleteTask marks a task as completed
func (uc *UseCase) CompleteTask(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	status := string(models.StatusCompleted)
	return uc.UpdateTask(ctx, id, models.UpdateTaskRequest{
		Status: &status,
	})
}
