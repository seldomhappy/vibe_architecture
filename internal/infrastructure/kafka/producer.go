package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/seldomhappy/vibe_architecture/internal/domain"
	pkgcontext "github.com/seldomhappy/vibe_architecture/internal/pkg/context"
	"github.com/seldomhappy/vibe_architecture/logger"
)

// Producer represents a Kafka producer
type Producer struct {
	producer sarama.SyncProducer
	topic    string
	logger   logger.ILogger
}

// ProducerConfig holds producer configuration
type ProducerConfig struct {
	Brokers      []string
	Topic        string
	Compression  string
	RetryMax     int
	RetryBackoff time.Duration
	Idempotent   bool
	Timeout      time.Duration
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg ProducerConfig, log logger.ILogger) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = cfg.RetryMax
	config.Producer.Retry.Backoff = cfg.RetryBackoff
	config.Producer.Idempotent = cfg.Idempotent
	config.Producer.Timeout = cfg.Timeout

	switch cfg.Compression {
	case "snappy":
		config.Producer.Compression = sarama.CompressionSnappy
	case "gzip":
		config.Producer.Compression = sarama.CompressionGZIP
	case "lz4":
		config.Producer.Compression = sarama.CompressionLZ4
	default:
		config.Producer.Compression = sarama.CompressionNone
	}

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &Producer{
		producer: producer,
		topic:    cfg.Topic,
		logger:   log,
	}, nil
}

// Start initializes the producer
func (p *Producer) Start(ctx context.Context) error {
	p.logger.Info("Kafka producer started for topic: %s", p.topic)
	return nil
}

// Shutdown closes the producer
func (p *Producer) Shutdown(ctx context.Context) error {
	p.logger.Info("Shutting down Kafka producer")
	return p.producer.Close()
}

// SendMessage sends a message to Kafka
func (p *Producer) SendMessage(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("trace_id"),
				Value: []byte(pkgcontext.GetTraceID(ctx)),
			},
			{
				Key:   []byte("request_id"),
				Value: []byte(pkgcontext.GetRequestID(ctx)),
			},
		},
		Timestamp: time.Now(),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error("Failed to send message to Kafka: %v", err)
		return fmt.Errorf("failed to send message: %w", err)
	}

	p.logger.Debug("Message sent to partition %d at offset %d", partition, offset)
	return nil
}

// PublishTaskCreated publishes a task created event
func (p *Producer) PublishTaskCreated(ctx context.Context, event domain.TaskCreatedEvent) error {
	return p.SendMessage(ctx, fmt.Sprintf("task-%d", event.TaskID), map[string]interface{}{
		"event_type": domain.EventTypeTaskCreated,
		"payload":    event,
		"timestamp":  time.Now(),
	})
}

// PublishTaskUpdated publishes a task updated event
func (p *Producer) PublishTaskUpdated(ctx context.Context, event domain.TaskUpdatedEvent) error {
	return p.SendMessage(ctx, fmt.Sprintf("task-%d", event.TaskID), map[string]interface{}{
		"event_type": domain.EventTypeTaskUpdated,
		"payload":    event,
		"timestamp":  time.Now(),
	})
}

// PublishTaskCompleted publishes a task completed event
func (p *Producer) PublishTaskCompleted(ctx context.Context, event domain.TaskCompletedEvent) error {
	return p.SendMessage(ctx, fmt.Sprintf("task-%d", event.TaskID), map[string]interface{}{
		"event_type": domain.EventTypeTaskCompleted,
		"payload":    event,
		"timestamp":  time.Now(),
	})
}

// PublishTaskDeleted publishes a task deleted event
func (p *Producer) PublishTaskDeleted(ctx context.Context, event domain.TaskDeletedEvent) error {
	return p.SendMessage(ctx, fmt.Sprintf("task-%d", event.TaskID), map[string]interface{}{
		"event_type": domain.EventTypeTaskDeleted,
		"payload":    event,
		"timestamp":  time.Now(),
	})
}
