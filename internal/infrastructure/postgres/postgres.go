package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/config"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/logger"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/metrics"
)

// DB represents PostgreSQL database connection
type DB struct {
	Pool    *pgxpool.Pool
	cfg     config.DBConfig
	log     logger.ILogger
	metrics *metrics.Metrics
}

// NewDB creates a new database connection
func NewDB(cfg config.DBConfig, log logger.ILogger, m *metrics.Metrics) (*DB, error) {
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MinConnections)
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("PostgreSQL connection established")

	return &DB{
		Pool:    pool,
		cfg:     cfg,
		log:     log,
		metrics: m,
	}, nil
}

// Start starts the database (no-op, already started in constructor)
func (db *DB) Start(ctx context.Context) error {
	return nil
}

// Stop closes the database connection
func (db *DB) Stop(ctx context.Context) error {
	if db.Pool != nil {
		db.Pool.Close()
		db.log.Info("PostgreSQL connection closed")
	}
	return nil
}

// GetPool returns the connection pool
func (db *DB) GetPool() *pgxpool.Pool {
	return db.Pool
}
