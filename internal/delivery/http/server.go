package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/seldomhappy/vibe_architecture/internal/delivery/http/handlers"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/config"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/logger"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/metrics"
	"github.com/seldomhappy/vibe_architecture/internal/usecase/task"
)

// Server represents HTTP server
type Server struct {
	cfg         *config.Config
	log         logger.ILogger
	taskUseCase *task.UseCase
	metrics     *metrics.Metrics
	server      *http.Server
}

// NewServer creates a new HTTP server
func NewServer(
	cfg *config.Config,
	log logger.ILogger,
	taskUseCase *task.UseCase,
	m *metrics.Metrics,
) *Server {
	return &Server{
		cfg:         cfg,
		log:         log,
		taskUseCase: taskUseCase,
		metrics:     m,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.healthHandler)

	// Task handlers
	taskHandler := handlers.NewTaskHandler(s.taskUseCase, s.log)
	mux.HandleFunc("/tasks", taskHandler.HandleTasks)
	mux.HandleFunc("/tasks/", taskHandler.HandleTaskByID)

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port),
		Handler:      s.loggingMiddleware(mux),
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
		IdleTimeout:  s.cfg.Server.IdleTimeout,
	}

	go func() {
		s.log.Info("HTTP server listening on %s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		s.log.Info("Shutting down HTTP server...")
		return s.server.Shutdown(ctx)
	}
	return nil
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": s.cfg.App.Name,
		"version": s.cfg.App.Version,
	})
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.log.Debug("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
