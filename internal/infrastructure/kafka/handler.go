package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/seldomhappy/vibe_architecture/internal/domain"
	pkgcontext "github.com/seldomhappy/vibe_architecture/internal/pkg/context"
	"github.com/seldomhappy/vibe_architecture/internal/pkg/tracing"
	"github.com/seldomhappy/vibe_architecture/logger"
	"go.opentelemetry.io/otel/attribute"
)

// TaskEventHandler handles task events from Kafka
type TaskEventHandler struct {
	logger logger.ILogger
}

// NewTaskEventHandler creates a new task event handler
func NewTaskEventHandler(log logger.ILogger) *TaskEventHandler {
	return &TaskEventHandler{
		logger: log,
	}
}

// Setup implements sarama.ConsumerGroupHandler
func (h *TaskEventHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup implements sarama.ConsumerGroupHandler
func (h *TaskEventHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim implements sarama.ConsumerGroupHandler
func (h *TaskEventHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.HandleMessage(session.Context(), message)
		session.MarkMessage(message, "")
	}
	return nil
}

// HandleMessage handles a single Kafka message
func (h *TaskEventHandler) HandleMessage(ctx context.Context, message *sarama.ConsumerMessage) {
	// Extract trace_id from headers to continue the trace
	var traceID string
	for _, header := range message.Headers {
		if string(header.Key) == "trace_id" {
			traceID = string(header.Value)
			break
		}
	}

	// Start a new span for message processing
	ctx, span := tracing.StartSpan(ctx, "kafka-consumer", "process_message")
	defer span.End()

	if traceID != "" {
		span.SetAttributes(attribute.String("trace_id", traceID))
	}

	span.SetAttributes(
		attribute.String("kafka.topic", message.Topic),
		attribute.Int64("kafka.partition", int64(message.Partition)),
		attribute.Int64("kafka.offset", message.Offset),
	)

	var event map[string]interface{}
	if err := json.Unmarshal(message.Value, &event); err != nil {
		h.logger.Error("[trace:%s] Failed to unmarshal message: %v", traceID, err)
		return
	}

	eventType, ok := event["event_type"].(string)
	if !ok {
		h.logger.Error("[trace:%s] Event type not found in message", traceID)
		return
	}

	h.logger.Info("[trace:%s] Processing event: %s", traceID, eventType)

	switch domain.EventType(eventType) {
	case domain.EventTypeTaskCreated:
		h.handleTaskCreated(ctx, event)
	case domain.EventTypeTaskUpdated:
		h.handleTaskUpdated(ctx, event)
	case domain.EventTypeTaskCompleted:
		h.handleTaskCompleted(ctx, event)
	case domain.EventTypeTaskDeleted:
		h.handleTaskDeleted(ctx, event)
	default:
		h.logger.Warn("[trace:%s] Unknown event type: %s", traceID, eventType)
	}
}

func (h *TaskEventHandler) handleTaskCreated(ctx context.Context, event map[string]interface{}) {
	traceID := pkgcontext.GetTraceID(ctx)
	h.logger.Info("[trace:%s] Task created event received: %+v", traceID, event["payload"])
	// Add business logic here (e.g., send notification, update cache, etc.)
}

func (h *TaskEventHandler) handleTaskUpdated(ctx context.Context, event map[string]interface{}) {
	traceID := pkgcontext.GetTraceID(ctx)
	h.logger.Info("[trace:%s] Task updated event received: %+v", traceID, event["payload"])
	// Add business logic here
}

func (h *TaskEventHandler) handleTaskCompleted(ctx context.Context, event map[string]interface{}) {
	traceID := pkgcontext.GetTraceID(ctx)
	h.logger.Info("[trace:%s] Task completed event received: %+v", traceID, event["payload"])
	// Add business logic here (e.g., send completion notification)
}

func (h *TaskEventHandler) handleTaskDeleted(ctx context.Context, event map[string]interface{}) {
	traceID := pkgcontext.GetTraceID(ctx)
	h.logger.Info("[trace:%s] Task deleted event received: %+v", traceID, event["payload"])
	// Add business logic here
}

// HandleTaskCreated handles a task created event (alternative method for direct calls)
func (h *TaskEventHandler) HandleTaskCreated(ctx context.Context, event domain.TaskCreatedEvent) error {
	h.logger.Info("Handling task created: %d - %s", event.TaskID, event.Name)
	// Add your business logic here
	return nil
}

// HandleTaskUpdated handles a task updated event
func (h *TaskEventHandler) HandleTaskUpdated(ctx context.Context, event domain.TaskUpdatedEvent) error {
	h.logger.Info("Handling task updated: %d - %s", event.TaskID, event.Name)
	// Add your business logic here
	return nil
}

// HandleTaskCompleted handles a task completed event
func (h *TaskEventHandler) HandleTaskCompleted(ctx context.Context, event domain.TaskCompletedEvent) error {
	h.logger.Info("Handling task completed: %d", event.TaskID)
	// Add your business logic here
	return nil
}

// HandleTaskDeleted handles a task deleted event
func (h *TaskEventHandler) HandleTaskDeleted(ctx context.Context, event domain.TaskDeletedEvent) error {
	h.logger.Info("Handling task deleted: %d", event.TaskID)
	// Add your business logic here
	return nil
}

// LogError logs an error with trace context
func (h *TaskEventHandler) LogError(ctx context.Context, format string, args ...interface{}) {
	traceID := pkgcontext.GetTraceID(ctx)
	msg := fmt.Sprintf(format, args...)
	h.logger.Error("[trace:%s] %s", traceID, msg)
}
