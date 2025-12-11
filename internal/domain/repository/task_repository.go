package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/seldomhappy/vibe_architecture/internal/domain/models"
)

// TaskRepository defines task repository interface
type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	List(ctx context.Context, limit, offset int) ([]*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TxManager defines transaction manager interface
type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
