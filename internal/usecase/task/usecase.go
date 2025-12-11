package task

import (
	"context"
	"fmt"
	"time"

	"github.com/seldomhappy/vibe_architecture/internal/domain"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/kafka"
	pkgcontext "github.com/seldomhappy/vibe_architecture/internal/pkg/context"
	"github.com/seldomhappy/vibe_architecture/internal/pkg/metrics"
	"github.com/seldomhappy/vibe_architecture/internal/pkg/tracing"
	"github.com/seldomhappy/vibe_architecture/internal/repository"
	"github.com/seldomhappy/vibe_architecture/logger"
	"go.opentelemetry.io/otel/attribute"
)

// TaskUseCase implements the UseCase interface
type TaskUseCase struct {
	repo     Repository
	producer *kafka.Producer
	logger   logger.ILogger
	metrics  *metrics.Metrics
}

// New creates a new task use case
func New(repo Repository, producer *kafka.Producer, log logger.ILogger, m *metrics.Metrics) UseCase {
	return &TaskUseCase{
		repo:     repo,
		producer: producer,
		logger:   log,
		metrics:  m,
	}
}

// CreateTask creates a new task
func (uc *TaskUseCase) CreateTask(ctx context.Context, input CreateTaskInput) (*domain.Task, error) {
	start := time.Now()
	ctx, span := tracing.StartSpan(ctx, "usecase", "create_task")
	defer span.End()

	requestID := pkgcontext.GetRequestID(ctx)
	traceID := pkgcontext.GetTraceID(ctx)

	span.SetAttributes(
		attribute.String("task.name", input.Name),
		attribute.String("task.priority", string(input.Priority)),
	)

	uc.logger.Info("[%s][trace:%s] Creating task: %s", requestID, traceID, input.Name)

	task := &domain.Task{
		Name:        input.Name,
		Description: input.Description,
		Status:      domain.TaskStatusPending,
		Priority:    input.Priority,
		CreatedBy:   input.CreatedBy,
	}

	if err := task.Validate(); err != nil {
		uc.logger.Error("[%s][trace:%s] Task validation failed: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		uc.metrics.RecordTaskFailed()
		return nil, err
	}

	if err := uc.repo.Create(ctx, task); err != nil {
		uc.logger.Error("[%s][trace:%s] Failed to create task: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		uc.metrics.RecordTaskFailed()
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Publish task created event
	event := domain.TaskCreatedEvent{
		TaskID:      task.ID,
		Name:        task.Name,
		Description: task.Description,
		Priority:    task.Priority,
		CreatedBy:   task.CreatedBy,
		CreatedAt:   task.CreatedAt,
	}

	if err := uc.producer.PublishTaskCreated(ctx, event); err != nil {
		uc.logger.Warn("[%s][trace:%s] Failed to publish task created event: %v", requestID, traceID, err)
	}

	uc.metrics.RecordTaskCreated()
	uc.metrics.RecordTaskProcessingDuration(time.Since(start))
	uc.logger.Info("[%s][trace:%s] Task created successfully: ID=%d", requestID, traceID, task.ID)

	return task, nil
}

// GetTask retrieves a task by ID
func (uc *TaskUseCase) GetTask(ctx context.Context, id int64) (*domain.Task, error) {
	ctx, span := tracing.StartSpan(ctx, "usecase", "get_task")
	defer span.End()

	requestID := pkgcontext.GetRequestID(ctx)
	traceID := pkgcontext.GetTraceID(ctx)

	span.SetAttributes(attribute.Int64("task.id", id))

	uc.logger.Debug("[%s][trace:%s] Getting task: ID=%d", requestID, traceID, id)

	task, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("[%s][trace:%s] Failed to get task: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		return nil, err
	}

	return task, nil
}

// ListTasks retrieves tasks with filters
func (uc *TaskUseCase) ListTasks(ctx context.Context, filter ListTasksFilter) ([]*domain.Task, error) {
	ctx, span := tracing.StartSpan(ctx, "usecase", "list_tasks")
	defer span.End()

	requestID := pkgcontext.GetRequestID(ctx)
	traceID := pkgcontext.GetTraceID(ctx)

	uc.logger.Debug("[%s][trace:%s] Listing tasks with filter", requestID, traceID)

	repoFilter := repository.TaskFilter{
		Status:     filter.Status,
		Priority:   filter.Priority,
		AssignedTo: filter.AssignedTo,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
	}

	tasks, err := uc.repo.GetAll(ctx, repoFilter)
	if err != nil {
		uc.logger.Error("[%s][trace:%s] Failed to list tasks: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))
	return tasks, nil
}

// UpdateTask updates an existing task
func (uc *TaskUseCase) UpdateTask(ctx context.Context, id int64, input UpdateTaskInput) (*domain.Task, error) {
	ctx, span := tracing.StartSpan(ctx, "usecase", "update_task")
	defer span.End()

	requestID := pkgcontext.GetRequestID(ctx)
	traceID := pkgcontext.GetTraceID(ctx)

	span.SetAttributes(attribute.Int64("task.id", id))

	uc.logger.Info("[%s][trace:%s] Updating task: ID=%d", requestID, traceID, id)

	task, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("[%s][trace:%s] Task not found: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		return nil, err
	}

	if input.Name != nil {
		task.Name = *input.Name
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Status != nil {
		task.Status = *input.Status
	}
	if input.Priority != nil {
		task.Priority = *input.Priority
	}
	task.UpdatedAt = time.Now()

	if err := task.Validate(); err != nil {
		uc.logger.Error("[%s][trace:%s] Task validation failed: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		uc.metrics.RecordTaskFailed()
		return nil, err
	}

	if err := uc.repo.Update(ctx, task); err != nil {
		uc.logger.Error("[%s][trace:%s] Failed to update task: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		uc.metrics.RecordTaskFailed()
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Publish task updated event
	event := domain.TaskUpdatedEvent{
		TaskID:      task.ID,
		Name:        task.Name,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		AssignedTo:  task.AssignedTo,
		UpdatedAt:   task.UpdatedAt,
	}

	if err := uc.producer.PublishTaskUpdated(ctx, event); err != nil {
		uc.logger.Warn("[%s][trace:%s] Failed to publish task updated event: %v", requestID, traceID, err)
	}

	uc.logger.Info("[%s][trace:%s] Task updated successfully: ID=%d", requestID, traceID, task.ID)

	return task, nil
}

// DeleteTask deletes a task
func (uc *TaskUseCase) DeleteTask(ctx context.Context, id int64) error {
	ctx, span := tracing.StartSpan(ctx, "usecase", "delete_task")
	defer span.End()

	requestID := pkgcontext.GetRequestID(ctx)
	traceID := pkgcontext.GetTraceID(ctx)

	span.SetAttributes(attribute.Int64("task.id", id))

	uc.logger.Info("[%s][trace:%s] Deleting task: ID=%d", requestID, traceID, id)

	if err := uc.repo.Delete(ctx, id); err != nil {
		uc.logger.Error("[%s][trace:%s] Failed to delete task: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		return err
	}

	// Publish task deleted event
	event := domain.TaskDeletedEvent{
		TaskID:    id,
		DeletedAt: time.Now(),
	}

	if err := uc.producer.PublishTaskDeleted(ctx, event); err != nil {
		uc.logger.Warn("[%s][trace:%s] Failed to publish task deleted event: %v", requestID, traceID, err)
	}

	uc.logger.Info("[%s][trace:%s] Task deleted successfully: ID=%d", requestID, traceID, id)

	return nil
}

// AssignTask assigns a task to a user
func (uc *TaskUseCase) AssignTask(ctx context.Context, taskID, userID int64) error {
	ctx, span := tracing.StartSpan(ctx, "usecase", "assign_task")
	defer span.End()

	requestID := pkgcontext.GetRequestID(ctx)
	traceID := pkgcontext.GetTraceID(ctx)

	span.SetAttributes(
		attribute.Int64("task.id", taskID),
		attribute.Int64("user.id", userID),
	)

	uc.logger.Info("[%s][trace:%s] Assigning task %d to user %d", requestID, traceID, taskID, userID)

	task, err := uc.repo.GetByID(ctx, taskID)
	if err != nil {
		uc.logger.Error("[%s][trace:%s] Task not found: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		return err
	}

	if err := task.Assign(userID); err != nil {
		uc.logger.Error("[%s][trace:%s] Failed to assign task: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		return err
	}

	if err := uc.repo.Update(ctx, task); err != nil {
		uc.logger.Error("[%s][trace:%s] Failed to save task: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		return fmt.Errorf("failed to save task: %w", err)
	}

	// Publish task updated event
	event := domain.TaskUpdatedEvent{
		TaskID:      task.ID,
		Name:        task.Name,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		AssignedTo:  task.AssignedTo,
		UpdatedAt:   task.UpdatedAt,
	}

	if err := uc.producer.PublishTaskUpdated(ctx, event); err != nil {
		uc.logger.Warn("[%s][trace:%s] Failed to publish task updated event: %v", requestID, traceID, err)
	}

	uc.logger.Info("[%s][trace:%s] Task assigned successfully", requestID, traceID)

	return nil
}

// CompleteTask marks a task as completed
func (uc *TaskUseCase) CompleteTask(ctx context.Context, id int64) error {
	start := time.Now()
	ctx, span := tracing.StartSpan(ctx, "usecase", "complete_task")
	defer span.End()

	requestID := pkgcontext.GetRequestID(ctx)
	traceID := pkgcontext.GetTraceID(ctx)

	span.SetAttributes(attribute.Int64("task.id", id))

	uc.logger.Info("[%s][trace:%s] Completing task: ID=%d", requestID, traceID, id)

	task, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("[%s][trace:%s] Task not found: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		return err
	}

	if err := task.Complete(); err != nil {
		uc.logger.Error("[%s][trace:%s] Failed to complete task: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		return err
	}

	if err := uc.repo.Update(ctx, task); err != nil {
		uc.logger.Error("[%s][trace:%s] Failed to save task: %v", requestID, traceID, err)
		tracing.RecordError(ctx, err)
		return fmt.Errorf("failed to save task: %w", err)
	}

	// Publish task completed event
	event := domain.TaskCompletedEvent{
		TaskID:      task.ID,
		CompletedAt: time.Now(),
	}

	if err := uc.producer.PublishTaskCompleted(ctx, event); err != nil {
		uc.logger.Warn("[%s][trace:%s] Failed to publish task completed event: %v", requestID, traceID, err)
	}

	uc.metrics.RecordTaskCompleted()
	uc.metrics.RecordTaskProcessingDuration(time.Since(start))
	uc.logger.Info("[%s][trace:%s] Task completed successfully: ID=%d", requestID, traceID, id)

	return nil
}
