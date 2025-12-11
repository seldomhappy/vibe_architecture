package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seldomhappy/vibe_architecture/internal/pkg/metrics"
	"github.com/seldomhappy/vibe_architecture/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DB wraps pgxpool.Pool with additional functionality
type DB struct {
	pool    *pgxpool.Pool
	logger  logger.ILogger
	metrics *metrics.Metrics
	tracer  trace.Tracer
}

// Config holds database configuration
type Config struct {
	DSN             string
	MaxOpenConns    int32
	MaxIdleConns    int32
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// New creates a new DB instance
func New(cfg Config, log logger.ILogger, m *metrics.Metrics, tracer trace.Tracer) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxOpenConns
	poolConfig.MinConns = cfg.MaxIdleConns
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	db := &DB{
		pool:    pool,
		logger:  log,
		metrics: m,
		tracer:  tracer,
	}

	return db, nil
}

// Start initializes the database connection
func (db *DB) Start(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	db.logger.Info("Database connection established")

	// Start monitoring pool stats
	go db.monitorPoolStats(ctx)

	return nil
}

// Shutdown closes the database connection
func (db *DB) Shutdown(ctx context.Context) error {
	db.logger.Info("Shutting down database connection")
	db.pool.Close()
	return nil
}

// Exec executes a query without returning any rows
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) error {
	start := time.Now()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.statement", query),
	)

	_, err := db.pool.Exec(ctx, query, args...)
	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
	}

	db.metrics.RecordDBQuery("exec", status, duration)
	return err
}

// Query executes a query that returns rows
func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	start := time.Now()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.statement", query),
	)

	rows, err := db.pool.Query(ctx, query, args...)
	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
	}

	db.metrics.RecordDBQuery("query", status, duration)
	return rows, err
}

// QueryRow executes a query that returns at most one row
func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	start := time.Now()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.statement", query),
	)

	row := db.pool.QueryRow(ctx, query, args...)
	duration := time.Since(start)

	db.metrics.RecordDBQuery("query_row", "success", duration)
	return row
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

// Pool returns the underlying connection pool
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// monitorPoolStats monitors and reports pool statistics
func (db *DB) monitorPoolStats(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stat := db.pool.Stat()
			db.metrics.SetDBConnections(stat.TotalConns(), stat.IdleConns())
			db.logger.Debug("Pool stats - Total: %d, Idle: %d, Acquired: %d",
				stat.TotalConns(), stat.IdleConns(), stat.AcquiredConns())
		case <-ctx.Done():
			return
		}
	}
}
