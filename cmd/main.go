package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/seldomhappy/vibe_architecture/config"
	httpdelivery "github.com/seldomhappy/vibe_architecture/internal/delivery/http"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/kafka"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/postgres"
	"github.com/seldomhappy/vibe_architecture/internal/pkg/lifecycle"
	"github.com/seldomhappy/vibe_architecture/internal/pkg/metrics"
	"github.com/seldomhappy/vibe_architecture/internal/pkg/tracing"
	"github.com/seldomhappy/vibe_architecture/internal/repository"
	"github.com/seldomhappy/vibe_architecture/internal/usecase/task"
	"github.com/seldomhappy/vibe_architecture/logger"
)

func main() {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	log := logger.New(cfg.App.Name)
	log.Info("Starting %s v%s in %s mode", cfg.App.Name, cfg.App.Version, cfg.App.Environment)

	// Run migrations if requested
	if os.Getenv("RUN_MIGRATIONS") == "true" {
		log.Info("Running database migrations...")
		if err := postgres.RunMigrations(cfg.DB.DSN(), log); err != nil {
			log.Fatal("Failed to run migrations: %v", err)
		}
		log.Info("Migrations completed successfully")
		return
	}

	// Initialize application
	app, err := initApp(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize application: %v", err)
	}

	// Start all services
	ctx := context.Background()
	if err := app.lifecycle.StartAll(ctx); err != nil {
		log.Fatal("Failed to start services: %v", err)
	}

	// Print startup information
	printStartupInfo(cfg, log)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down gracefully...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := app.lifecycle.ShutdownAll(shutdownCtx); err != nil {
		log.Error("Error during shutdown: %v", err)
	}

	log.Info("Server stopped")
}

type application struct {
	lifecycle *lifecycle.Manager
	logger    logger.ILogger
}

func loadConfig() (*config.Config, error) {
	var cfg config.Config

	// Try to load from config file
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		env := os.Getenv("APP_ENVIRONMENT")
		if env == "production" {
			configPath = "config/config.production.yaml"
		} else {
			configPath = "config/config.yaml"
		}
	}

	if _, err := os.Stat(configPath); err == nil {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		// Load from environment variables only
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("failed to read environment: %w", err)
		}
	}

	return &cfg, nil
}

func initApp(cfg *config.Config, log logger.ILogger) (*application, error) {
	lm := lifecycle.New()

	// 1. Initialize Metrics
	log.Info("Initializing metrics...")
	m := metrics.New(cfg.App.Name, cfg.App.Version, cfg.Metrics.Port, cfg.Metrics.Enabled)
	lm.Register("metrics", m)

	// 2. Initialize Tracing
	log.Info("Initializing tracing...")
	tracer, err := tracing.New(
		cfg.Tracing.ServiceName,
		cfg.Tracing.JaegerEndpoint,
		cfg.Tracing.SamplingRate,
		cfg.Tracing.Enabled,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}
	lm.Register("tracing", tracer)

	// 3. Initialize Database
	log.Info("Initializing database...")
	dbConfig := postgres.Config{
		DSN:             cfg.DB.DSN(),
		MaxOpenConns:    int32(cfg.DB.MaxOpenConns),
		MaxIdleConns:    int32(cfg.DB.MaxIdleConns),
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.DB.ConnMaxIdleTime,
	}
	
	dbTracer := tracing.GetTracer("postgres")
	db, err := postgres.New(dbConfig, log, m, dbTracer)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	lm.Register("database", db)

	// 4. Initialize Kafka Producer
	log.Info("Initializing Kafka producer...")
	producerConfig := kafka.ProducerConfig{
		Brokers:      cfg.Kafka.Brokers,
		Topic:        cfg.Kafka.Topics.TaskEvents,
		Compression:  cfg.Kafka.Producer.Compression,
		RetryMax:     cfg.Kafka.Producer.RetryMax,
		RetryBackoff: cfg.Kafka.Producer.RetryBackoff,
		Idempotent:   cfg.Kafka.Producer.Idempotent,
		Timeout:      cfg.Kafka.Producer.Timeout,
	}
	producer, err := kafka.NewProducer(producerConfig, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kafka producer: %w", err)
	}
	lm.Register("kafka-producer", producer)

	// 5. Initialize Repositories
	log.Info("Initializing repositories...")
	taskRepo := repository.NewTaskRepository(db, log)
	txManager := repository.NewTxManager(db, log)
	_ = txManager // For future use with transactions

	// 6. Initialize Use Cases
	log.Info("Initializing use cases...")
	taskUC := task.New(taskRepo, producer, log, m)

	// 7. Initialize Kafka Consumer
	log.Info("Initializing Kafka consumer...")
	eventHandler := kafka.NewTaskEventHandler(log)
	consumerConfig := kafka.ConsumerConfig{
		Brokers:          cfg.Kafka.Brokers,
		GroupID:          cfg.Kafka.ConsumerGroupID,
		Topics:           []string{cfg.Kafka.Topics.TaskEvents},
		Workers:          cfg.Kafka.Consumer.Workers,
		SessionTimeout:   cfg.Kafka.Consumer.SessionTimeout.String(),
		RebalanceTimeout: cfg.Kafka.Consumer.RebalanceTimeout.String(),
	}
	consumer, err := kafka.NewConsumer(consumerConfig, eventHandler, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kafka consumer: %w", err)
	}
	lm.Register("kafka-consumer", consumer)

	// 8. Initialize HTTP Server
	log.Info("Initializing HTTP server...")
	serverConfig := httpdelivery.Config{
		Host:            cfg.Server.Host,
		Port:            cfg.Server.Port,
		ReadTimeout:     cfg.Server.ReadTimeout,
		WriteTimeout:    cfg.Server.WriteTimeout,
		ShutdownTimeout: cfg.Server.ShutdownTimeout,
	}
	httpServer := httpdelivery.New(serverConfig, taskUC, m, log)
	lm.Register("http-server", httpServer)

	return &application{
		lifecycle: lm,
		logger:    log,
	}, nil
}

func printStartupInfo(cfg *config.Config, log logger.ILogger) {
	log.Info("===========================================")
	log.Info("  %s v%s", cfg.App.Name, cfg.App.Version)
	log.Info("===========================================")
	log.Info("HTTP Server:   http://%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("Health Check:  http://%s:%d/health", cfg.Server.Host, cfg.Server.Port)
	if cfg.Metrics.Enabled {
		log.Info("Metrics:       http://localhost:%d%s", cfg.Metrics.Port, cfg.Metrics.Path)
	}
	if cfg.Tracing.Enabled {
		log.Info("Tracing:       %s", cfg.Tracing.JaegerEndpoint)
		log.Info("Jaeger UI:     http://localhost:16686")
	}
	log.Info("===========================================")
	log.Info("Environment:   %s", cfg.App.Environment)
	log.Info("Debug Mode:    %v", cfg.App.Debug)
	log.Info("===========================================")
	log.Info("Application started successfully!")
}
