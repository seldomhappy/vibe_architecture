package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"github.com/seldomhappy/vibe_architecture/logger"
)

// Consumer represents a Kafka consumer
type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	topics        []string
	handler       *TaskEventHandler
	logger        logger.ILogger
	workers       int
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// ConsumerConfig holds consumer configuration
type ConsumerConfig struct {
	Brokers          []string
	GroupID          string
	Topics           []string
	Workers          int
	SessionTimeout   string
	RebalanceTimeout string
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg ConsumerConfig, handler *TaskEventHandler, log logger.ILogger) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		consumerGroup: consumerGroup,
		topics:        cfg.Topics,
		handler:       handler,
		logger:        log,
		workers:       cfg.Workers,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Start starts the consumer
func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("Starting Kafka consumer for topics: %v with %d workers", c.topics, c.workers)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.consumerGroup.Consume(c.ctx, c.topics, c.handler); err != nil {
				c.logger.Error("Error from consumer: %v", err)
			}
			if c.ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the consumer
func (c *Consumer) Shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down Kafka consumer")
	c.cancel()
	c.wg.Wait()
	return c.consumerGroup.Close()
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler *TaskEventHandler
	logger  logger.ILogger
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.handler.HandleMessage(session.Context(), message)
		session.MarkMessage(message, "")
	}
	return nil
}
