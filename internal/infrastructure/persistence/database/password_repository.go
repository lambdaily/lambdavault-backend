package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

// FindByIDAndUserID looks up a password owned by a user. Owned passwords stay
// in the owner's personal vault even when they are shared with groups.
func (r *passwordRepository) FindByIDAndUserID(ctx context.Context, id, userID uuid.UUID) (*entity.Password, error) {
	var password entity.Password
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&password)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrPasswordNotFound
		}
		return nil, result.Error
	}
	return &password, nil
}

func (r *passwordRepository) FindByIDAndGroupID(ctx context.Context, id, groupID uuid.UUID) (*entity.Password, error) {
	var password entity.Password
	result := r.sharedPasswordQuery(ctx, groupID).
		Where("passwords.id = ?", id).
		First(&password)
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
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&passwords)
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

func (r *passwordRepository) ShareWithGroup(ctx context.Context, passwordID, groupID, sharedByID uuid.UUID) error {
	share := entity.NewPasswordGroupShare(passwordID, groupID, sharedByID)
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(share).Error
}

func (r *passwordRepository) UnshareFromGroup(ctx context.Context, passwordID, groupID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&entity.PasswordGroupShare{}, "password_id = ? AND group_id = ?", passwordID, groupID).Error
}

func (r *passwordRepository) FindAllByGroupID(ctx context.Context, groupID uuid.UUID) ([]*entity.Password, error) {
	var passwords []*entity.Password
	result := r.sharedPasswordQuery(ctx, groupID).
		Order("passwords.created_at DESC").
		Find(&passwords)
	if result.Error != nil {
		return nil, result.Error
	}
	return passwords, nil
}

func (r *passwordRepository) SearchByGroupID(ctx context.Context, groupID uuid.UUID, query string) ([]*entity.Password, error) {
	var passwords []*entity.Password
	searchPattern := "%" + query + "%"
	err := r.sharedPasswordQuery(ctx, groupID).
		Where("passwords.site_name LIKE ? OR passwords.username LIKE ? OR passwords.notes LIKE ? OR passwords.category LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Order("passwords.updated_at DESC").
		Find(&passwords).Error
	if err != nil {
		return nil, err
	}
	return passwords, nil
}

func (r *passwordRepository) sharedPasswordQuery(ctx context.Context, groupID uuid.UUID) *gorm.DB {
	return r.db.WithContext(ctx).
		Model(&entity.Password{}).
		Joins("JOIN password_group_shares ON password_group_shares.password_id = passwords.id").
		Where("password_group_shares.group_id = ?", groupID).
		Distinct()
}
