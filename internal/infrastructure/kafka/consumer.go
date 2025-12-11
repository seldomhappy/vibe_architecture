package kafka

import (
	"context"

	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/config"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/logger"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/metrics"
)

// Consumer represents Kafka consumer
type Consumer struct {
	cfg     config.KafkaConfig
	log     logger.ILogger
	metrics *metrics.Metrics
	handler EventHandler
}

// EventHandler defines the interface for handling Kafka events
type EventHandler interface {
	HandleTaskEvent(ctx context.Context, event []byte) error
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg config.KafkaConfig, handler EventHandler, log logger.ILogger, m *metrics.Metrics) (*Consumer, error) {
	if !cfg.Enabled || !cfg.Consumer.Enabled {
		log.Info("Kafka consumer disabled")
		return &Consumer{
			cfg:     cfg,
			log:     log,
			metrics: m,
			handler: handler,
		}, nil
	}

	// TODO: Implement actual Kafka consumer with sarama
	log.Info("Kafka consumer initialized (stub)")

	return &Consumer{
		cfg:     cfg,
		log:     log,
		metrics: m,
		handler: handler,
	}, nil
}

// Start starts the consumer
func (c *Consumer) Start(ctx context.Context) error {
	if !c.cfg.Enabled || !c.cfg.Consumer.Enabled {
		return nil
	}
	c.log.Info("Kafka consumer started")
	return nil
}

// Stop stops the consumer
func (c *Consumer) Stop(ctx context.Context) error {
	if !c.cfg.Enabled || !c.cfg.Consumer.Enabled {
		return nil
	}
	c.log.Info("Kafka consumer stopped")
	return nil
}

// TaskEventHandler implements EventHandler
type TaskEventHandler struct {
	log logger.ILogger
}

// NewTaskEventHandler creates a new task event handler
func NewTaskEventHandler(log logger.ILogger) *TaskEventHandler {
	return &TaskEventHandler{log: log}
}

// HandleTaskEvent handles task events
func (h *TaskEventHandler) HandleTaskEvent(ctx context.Context, event []byte) error {
	h.log.Debug("Received task event: %s", string(event))
	return nil
}
