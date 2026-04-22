package sqlite

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lambdavault/api/internal/domain/entity"
	domainErrors "github.com/lambdavault/api/internal/domain/errors"
	"github.com/lambdavault/api/internal/domain/repository"
)

type passwordRepository struct {
	db *gorm.DB
}

func NewPasswordRepository(db *gorm.DB) repository.PasswordRepository {
	return &passwordRepository{db: db}
}

func (r *passwordRepository) Create(ctx context.Context, password *entity.Password) error {
	return r.db.WithContext(ctx).Create(password).Error
}

func (r *passwordRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Password, error) {
	var password entity.Password
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&password)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrPasswordNotFound
		}
		return nil, result.Error
	}
	return &password, nil
}

func (r *passwordRepository) FindByIDAndUserID(ctx context.Context, id, userID uuid.UUID) (*entity.Password, error) {
	var password entity.Password
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&password)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrPasswordNotFound
		}
		return nil, result.Error
	}
	return &password, nil
}

func (r *passwordRepository) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Password, error) {
	var passwords []*entity.Password
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&passwords)
	if result.Error != nil {
		return nil, result.Error
	}
	return passwords, nil
}

func (r *passwordRepository) FindByUserIDAndSiteName(ctx context.Context, userID uuid.UUID, siteName string) ([]*entity.Password, error) {
	var passwords []*entity.Password
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND site_name LIKE ?", userID, "%"+siteName+"%").
		Order("created_at DESC").
		Find(&passwords)
	if result.Error != nil {
		return nil, result.Error
	}
	return passwords, nil
}

func (r *passwordRepository) Update(ctx context.Context, password *entity.Password) error {
	return r.db.WithContext(ctx).Save(password).Error
}

func (r *passwordRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Password{}, "id = ?", id)
	return result.Error
}

func (r *passwordRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Password{}, "user_id = ?", userID)
	return result.Error
}

func (r *passwordRepository) SearchByUserID(ctx context.Context, userID uuid.UUID, query string) ([]*entity.Password, error) {
	var passwords []*entity.Password
	searchPattern := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("site_name LIKE ? OR username LIKE ? OR notes LIKE ? OR category LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Order("updated_at DESC").
		Find(&passwords).Error
	if err != nil {
		return nil, err
	}
	return passwords, nil
}
