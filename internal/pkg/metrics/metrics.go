package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal      *prometheus.CounterVec
	HTTPRequestDuration    *prometheus.HistogramVec
	HTTPRequestsInFlight   prometheus.Gauge

	// Business metrics
	TasksCreatedTotal      prometheus.Counter
	TasksCompletedTotal    prometheus.Counter
	TasksFailedTotal       prometheus.Counter
	TasksByStatus          *prometheus.GaugeVec
	TaskProcessingDuration prometheus.Histogram

	// DB metrics
	DBConnectionsOpen      prometheus.Gauge
	DBConnectionsIdle      prometheus.Gauge
	DBQueryDuration        *prometheus.HistogramVec
	DBQueriesTotal         *prometheus.CounterVec

	// System metrics
	AppInfo                *prometheus.GaugeVec
	AppUptime              prometheus.Counter

	server  *http.Server
	enabled bool
	startTime time.Time
}

// New creates a new metrics instance
func New(serviceName, version string, port int, enabled bool) *Metrics {
	if !enabled {
		return &Metrics{enabled: false}
	}

	m := &Metrics{
		enabled:   true,
		startTime: time.Now(),

		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),

		// Business metrics
		TasksCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "tasks_created_total",
				Help: "Total number of tasks created",
			},
		),
		TasksCompletedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "tasks_completed_total",
				Help: "Total number of tasks completed",
			},
		),
		TasksFailedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "tasks_failed_total",
				Help: "Total number of failed task operations",
			},
		),
		TasksByStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tasks_by_status",
				Help: "Number of tasks by status",
			},
			[]string{"status"},
		),
		TaskProcessingDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "task_processing_duration_seconds",
				Help:    "Task processing duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
		),

		// DB metrics
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_open",
				Help: "Number of open database connections",
			},
		),
		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_idle",
				Help: "Number of idle database connections",
			},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"query"},
		),
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"query", "status"},
		),

		// System metrics
		AppInfo: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "app_info",
				Help: "Application information",
			},
			[]string{"service", "version"},
		),
		AppUptime: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "app_uptime_seconds",
				Help: "Application uptime in seconds",
			},
		),
	}

	m.AppInfo.WithLabelValues(serviceName, version).Set(1)

	// Create HTTP server for metrics endpoint
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return m
}

// Start starts the metrics HTTP server
func (m *Metrics) Start(ctx context.Context) error {
	if !m.enabled {
		return nil
	}

	// Start uptime counter goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.AppUptime.Add(1)
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error but don't stop the application
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the metrics server
func (m *Metrics) Shutdown(ctx context.Context) error {
	if !m.enabled || m.server == nil {
		return nil
	}
	return m.server.Shutdown(ctx)
}

// RecordHTTPRequest records an HTTP request metric
func (m *Metrics) RecordHTTPRequest(method, path, status string, duration time.Duration) {
	if !m.enabled {
		return
	}
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// IncHTTPRequestsInFlight increments the in-flight requests gauge
func (m *Metrics) IncHTTPRequestsInFlight() {
	if !m.enabled {
		return
	}
	m.HTTPRequestsInFlight.Inc()
}

// DecHTTPRequestsInFlight decrements the in-flight requests gauge
func (m *Metrics) DecHTTPRequestsInFlight() {
	if !m.enabled {
		return
	}
	m.HTTPRequestsInFlight.Dec()
}

// RecordTaskCreated records a task creation
func (m *Metrics) RecordTaskCreated() {
	if !m.enabled {
		return
	}
	m.TasksCreatedTotal.Inc()
}

// RecordTaskCompleted records a task completion
func (m *Metrics) RecordTaskCompleted() {
	if !m.enabled {
		return
	}
	m.TasksCompletedTotal.Inc()
}

// RecordTaskFailed records a failed task operation
func (m *Metrics) RecordTaskFailed() {
	if !m.enabled {
		return
	}
	m.TasksFailedTotal.Inc()
}

// SetTasksByStatus sets the number of tasks for a given status
func (m *Metrics) SetTasksByStatus(status string, count float64) {
	if !m.enabled {
		return
	}
	m.TasksByStatus.WithLabelValues(status).Set(count)
}

// RecordTaskProcessingDuration records task processing duration
func (m *Metrics) RecordTaskProcessingDuration(duration time.Duration) {
	if !m.enabled {
		return
	}
	m.TaskProcessingDuration.Observe(duration.Seconds())
}

// RecordDBQuery records a database query
func (m *Metrics) RecordDBQuery(query, status string, duration time.Duration) {
	if !m.enabled {
		return
	}
	m.DBQueriesTotal.WithLabelValues(query, status).Inc()
	m.DBQueryDuration.WithLabelValues(query).Observe(duration.Seconds())
}

// SetDBConnections sets database connection metrics
func (m *Metrics) SetDBConnections(open, idle int32) {
	if !m.enabled {
		return
	}
	m.DBConnectionsOpen.Set(float64(open))
	m.DBConnectionsIdle.Set(float64(idle))
}
