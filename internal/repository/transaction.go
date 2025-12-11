package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/postgres"
	"github.com/seldomhappy/vibe_architecture/logger"
)

// TxManager manages database transactions
type TxManager struct {
	db     *postgres.DB
	logger logger.ILogger
}

// NewTxManager creates a new transaction manager
func NewTxManager(db *postgres.DB, log logger.ILogger) *TxManager {
	return &TxManager{
		db:     db,
		logger: log,
	}
}

// WithTransaction executes a function within a transaction
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := tm.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	err = fn(ctx, tx)
	return err
}
