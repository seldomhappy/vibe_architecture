package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/seldomhappy/vibe_architecture/internal/pkg/metrics"
	"github.com/seldomhappy/vibe_architecture/internal/usecase/task"
	"github.com/seldomhappy/vibe_architecture/logger"
)

// Server represents the HTTP server
type Server struct {
	server  *http.Server
	handler *TaskHandler
	logger  logger.ILogger
}

// Config holds server configuration
type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// New creates a new HTTP server
func New(cfg Config, taskUC task.UseCase, m *metrics.Metrics, log logger.ILogger) *Server {
	handler := NewTaskHandler(taskUC, log)

	mux := http.NewServeMux()
	
	// Health check
	mux.HandleFunc("/health", handler.Health)
	
	// Task routes
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.ListTasks(w, r)
		case http.MethodPost:
			handler.CreateTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	mux.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		// Check if it's an action endpoint
		if contains(r.URL.Path, "/assign") {
			if r.Method == http.MethodPost {
				handler.AssignTask(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		
		if contains(r.URL.Path, "/complete") {
			if r.Method == http.MethodPost {
				handler.CompleteTask(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		
		// Regular CRUD operations
		switch r.Method {
		case http.MethodGet:
			handler.GetTask(w, r)
		case http.MethodPut:
			handler.UpdateTask(w, r)
		case http.MethodDelete:
			handler.DeleteTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Apply middleware chain in correct order
	finalHandler := RecoveryMiddleware(log)(
		RequestIDMiddleware()(
			TracingMiddleware()(
				LoggingMiddleware(log)(
					MetricsMiddleware(m)(
						TimeoutMiddleware(30*time.Second)(mux),
					),
				),
			),
		),
	)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      finalHandler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &Server{
		server:  server,
		handler: handler,
		logger:  log,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting HTTP server on %s", s.server.Addr)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
