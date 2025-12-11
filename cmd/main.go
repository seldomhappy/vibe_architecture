package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seldomhappy/vibe_architecture/internal/delivery/http"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/config"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/kafka"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/lifecycle"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/logger"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/metrics"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/postgres"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/tracing"
	"github.com/seldomhappy/vibe_architecture/internal/repository"
	taskUC "github.com/seldomhappy/vibe_architecture/internal/usecase/task"
)

// App represents the application
type App struct {
	lifecycle *lifecycle.Manager
	log       logger.ILogger
}

func main() {
	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize logger
	log := logger.NewLogger(cfg.Logger)
	log.Info("=== Starting %s v%s ===", cfg.App.Name, cfg.App.Version)

	// Initialize application
	app, err := initApp(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize application: %v", err)
	}

	// Start application
	ctx := context.Background()
	if err := app.lifecycle.Start(ctx); err != nil {
		log.Fatal("Failed to start application: %v", err)
	}

	log.Info("=== Application started successfully ===")

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("=== Shutting down application ===")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.lifecycle.Stop(shutdownCtx); err != nil {
		log.Error("Error during shutdown: %v", err)
	}

	log.Info("=== Application stopped ===")
}

func initApp(cfg *config.Config, log logger.ILogger) (*App, error) {
	lm := lifecycle.NewManager()

	// 1. Metrics
	var m *metrics.Metrics
	if cfg.Metrics.Enabled {
		m = metrics.NewMetrics(cfg.Metrics, log)
		lm.Register(m)
		log.Info("✓ Metrics initialized")
	}

	// 2. Tracing (опционально)
	if cfg.Tracing.Enabled {
		tracer, err := tracing.NewTracer(cfg.Tracing, log)
		if err != nil {
			return nil, err
		}
		lm.Register(tracer)
		log.Info("✓ Tracing initialized")
	}

	// 3. Database (pgx)
	db, err := postgres.NewDB(cfg.DB, log, m)
	if err != nil {
		return nil, err
	}
	lm.Register(db)
	log.Info("✓ PostgreSQL (pgx) initialized")

	// 4. Kafka Producer (опционально)
	var kafkaProducer *kafka.Producer
	if cfg.Kafka.Enabled && cfg.Kafka.Producer.Enabled {
		kafkaProducer, err = kafka.NewProducer(cfg.Kafka, log, m)
		if err != nil {
			log.Warn("Kafka producer initialization failed (disabled): %v", err)
			kafkaProducer = nil
		} else {
			lm.Register(kafkaProducer)
			log.Info("✓ Kafka producer initialized")
		}
	} else {
		log.Info("⊘ Kafka producer disabled")
		kafkaProducer = kafka.NewDisabledProducer(log)
	}

	// 5. Repositories
	taskRepo := repository.NewTaskRepository(db, log, m)
	txManager := repository.NewTxManager(db, log)

	// 6. Use Cases
	taskUseCase := taskUC.NewUseCase(taskRepo, txManager, log, m, kafkaProducer)

	// 7. Kafka Consumer (опционально)
	if cfg.Kafka.Enabled && cfg.Kafka.Consumer.Enabled {
		kafkaHandler := kafka.NewTaskEventHandler(log)
		kafkaConsumer, err := kafka.NewConsumer(cfg.Kafka, kafkaHandler, log, m)
		if err != nil {
			log.Warn("Kafka consumer initialization failed (disabled): %v", err)
		} else {
			lm.Register(kafkaConsumer)
			log.Info("✓ Kafka consumer initialized")
		}
	} else {
		log.Info("⊘ Kafka consumer disabled")
	}

	// 8. HTTP Server
	httpServer := http.NewServer(cfg, log, taskUseCase, m)
	lm.Register(httpServer)
	log.Info("✓ HTTP server initialized")

	return &App{
		lifecycle: lm,
		log:       log,
	}, nil
}
