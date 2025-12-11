package postgres

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
	"github.com/seldomhappy/vibe_architecture/logger"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// RunMigrations runs database migrations using tern
func RunMigrations(dsn string, log logger.ILogger) error {
	log.Info("Running database migrations...")

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	migrator, err := migrate.NewMigrator(ctx, conn, "schema_version")
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	migrator.OnStart = func(sequence int32, name, direction, sql string) {
		log.Info("Executing migration: %s (%s)", name, direction)
	}

	// Load migrations from embedded file system
	if err := migrator.LoadMigrations(migrationFiles); err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	if err := migrator.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("Database migrations completed successfully")
	return nil
}
