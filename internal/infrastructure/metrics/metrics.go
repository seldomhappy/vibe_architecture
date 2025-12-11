package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/config"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/logger"
)

// Metrics represents metrics server
type Metrics struct {
	cfg    config.MetricsConfig
	log    logger.ILogger
	server *http.Server
}

// NewMetrics creates a new metrics instance
func NewMetrics(cfg config.MetricsConfig, log logger.ILogger) *Metrics {
	return &Metrics{
		cfg: cfg,
		log: log,
	}
}

// Start starts the metrics server
func (m *Metrics) Start(ctx context.Context) error {
	if !m.cfg.Enabled {
		m.log.Info("Metrics disabled")
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle(m.cfg.Path, promhttp.Handler())

	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.cfg.Port),
		Handler: mux,
	}

	go func() {
		m.log.Info("Metrics server listening on :%d%s", m.cfg.Port, m.cfg.Path)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.log.Error("Metrics server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the metrics server
func (m *Metrics) Stop(ctx context.Context) error {
	if m.server != nil {
		return m.server.Shutdown(ctx)
	}
	return nil
}
