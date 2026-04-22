package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/lambdavault/api/internal/domain/entity"
)

type PasswordRepository interface {
	Create(ctx context.Context, password *entity.Password) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Password, error)
	FindByIDAndUserID(ctx context.Context, id, userID uuid.UUID) (*entity.Password, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Password, error)
	FindByUserIDAndSiteName(ctx context.Context, userID uuid.UUID, siteName string) ([]*entity.Password, error)
	Update(ctx context.Context, password *entity.Password) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
	SearchByUserID(ctx context.Context, userID uuid.UUID, query string) ([]*entity.Password, error)
}
