package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/logger"
	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/postgres"
)

// TxManager implements transaction manager
type TxManager struct {
	db  *postgres.DB
	log logger.ILogger
}

// NewTxManager creates a new transaction manager
func NewTxManager(db *postgres.DB, log logger.ILogger) *TxManager {
	return &TxManager{
		db:  db,
		log: log,
	}
}

type txKey struct{}

// WithinTransaction executes a function within a transaction
func (tm *TxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	ctx = context.WithValue(ctx, txKey{}, tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(ctx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			tm.log.Error("Failed to rollback transaction: %v", rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTx retrieves transaction from context
func GetTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}
