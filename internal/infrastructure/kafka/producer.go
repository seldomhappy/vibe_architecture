package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/config"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/logger"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/metrics"
)

// Producer represents Kafka producer
type Producer struct {
	producer sarama.SyncProducer
	cfg      config.KafkaConfig
	log      logger.ILogger
	metrics  *metrics.Metrics
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg config.KafkaConfig, log logger.ILogger, m *metrics.Metrics) (*Producer, error) {
	if !cfg.Enabled || !cfg.Producer.Enabled {
		log.Info("Kafka producer disabled")
		return NewDisabledProducer(log), nil
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3

	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Kafka version: %w", err)
	}
	config.Version = version

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	log.Info("Kafka producer initialized")

	return &Producer{
		producer: producer,
		cfg:      cfg,
		log:      log,
		metrics:  m,
	}, nil
}

// NewDisabledProducer creates a disabled producer stub
func NewDisabledProducer(log logger.ILogger) *Producer {
	return &Producer{
		producer: nil,
		log:      log,
		metrics:  nil,
	}
}

// Start starts the producer (no-op)
func (p *Producer) Start(ctx context.Context) error {
	return nil
}

// Stop stops the producer
func (p *Producer) Stop(ctx context.Context) error {
	if p.producer != nil {
		if err := p.producer.Close(); err != nil {
			p.log.Error("Failed to close Kafka producer: %v", err)
			return err
		}
		p.log.Info("Kafka producer closed")
	}
	return nil
}

// SendTaskCreated sends task created event
func (p *Producer) SendTaskCreated(ctx context.Context, event interface{}) error {
	if p.producer == nil {
		p.log.Debug("Kafka producer disabled, skipping SendTaskCreated")
		return nil
	}
	return p.sendEvent(p.cfg.Topics.TaskCreated.Name, event)
}

// SendTaskUpdated sends task updated event
func (p *Producer) SendTaskUpdated(ctx context.Context, event interface{}) error {
	if p.producer == nil {
		p.log.Debug("Kafka producer disabled, skipping SendTaskUpdated")
		return nil
	}
	return p.sendEvent(p.cfg.Topics.TaskUpdated.Name, event)
}

// SendTaskCompleted sends task completed event
func (p *Producer) SendTaskCompleted(ctx context.Context, event interface{}) error {
	if p.producer == nil {
		p.log.Debug("Kafka producer disabled, skipping SendTaskCompleted")
		return nil
	}
	return p.sendEvent(p.cfg.Topics.TaskCompleted.Name, event)
}

// SendTaskDeleted sends task deleted event
func (p *Producer) SendTaskDeleted(ctx context.Context, event interface{}) error {
	if p.producer == nil {
		p.log.Debug("Kafka producer disabled, skipping SendTaskDeleted")
		return nil
	}
	return p.sendEvent(p.cfg.Topics.TaskDeleted.Name, event)
}

func (p *Producer) sendEvent(topic string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	p.log.Debug("Event sent to topic %s", topic)
	return nil
}
