package database

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lambdavault/api/internal/domain/entity"
	domainErrors "github.com/lambdavault/api/internal/domain/errors"
	"github.com/lambdavault/api/internal/domain/repository"
)

type passwordGroupRepository struct {
	db *gorm.DB
}

func NewPasswordGroupRepository(db *gorm.DB) repository.PasswordGroupRepository {
	return &passwordGroupRepository{db: db}
}

func (r *passwordGroupRepository) Create(ctx context.Context, group *entity.PasswordGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *passwordGroupRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.PasswordGroup, error) {
	var group entity.PasswordGroup
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&group)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrGroupNotFound
		}
		return nil, result.Error
	}
	return &group, nil
}

func (r *passwordGroupRepository) Update(ctx context.Context, group *entity.PasswordGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

func (r *passwordGroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.PasswordGroup{}, "id = ?", id).Error
}

func (r *passwordGroupRepository) FindAccessibleByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.PasswordGroup, error) {
	var groups []*entity.PasswordGroup
	err := r.db.WithContext(ctx).
		Where(`id IN (
			SELECT id FROM password_groups WHERE owner_id = ?
			UNION
			SELECT group_id FROM group_members WHERE user_id = ? AND status = ?
		)`, userID, userID, entity.GroupMemberStatusActive).
		Order("created_at DESC").
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *passwordGroupRepository) AddMember(ctx context.Context, member *entity.GroupMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *passwordGroupRepository) UpdateMember(ctx context.Context, member *entity.GroupMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *passwordGroupRepository) RemoveMember(ctx context.Context, memberID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.GroupMember{}, "id = ?", memberID).Error
}

func (r *passwordGroupRepository) FindMemberByID(ctx context.Context, memberID uuid.UUID) (*entity.GroupMember, error) {
	var member entity.GroupMember
	result := r.db.WithContext(ctx).Where("id = ?", memberID).First(&member)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrGroupMemberNotFound
		}
		return nil, result.Error
	}
	return &member, nil
}

func (r *passwordGroupRepository) FindMemberByGroupAndUser(ctx context.Context, groupID, userID uuid.UUID) (*entity.GroupMember, error) {
	var member entity.GroupMember
	result := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrGroupMemberNotFound
		}
		return nil, result.Error
	}
	return &member, nil
}

func (r *passwordGroupRepository) FindMemberByGroupAndEmail(ctx context.Context, groupID uuid.UUID, email string) (*entity.GroupMember, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var member entity.GroupMember
	result := r.db.WithContext(ctx).
		Where("group_id = ? AND email = ?", groupID, email).
		First(&member)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrGroupMemberNotFound
		}
		return nil, result.Error
	}
	return &member, nil
}

func (r *passwordGroupRepository) ListMembers(ctx context.Context, groupID uuid.UUID) ([]*entity.GroupMember, error) {
	var members []*entity.GroupMember
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("created_at ASC").
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (r *passwordGroupRepository) ActivatePendingForEmail(ctx context.Context, email string, userID uuid.UUID) (int64, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	result := r.db.WithContext(ctx).
		Model(&entity.GroupMember{}).
		Where("email = ? AND status = ? AND (user_id IS NULL OR user_id = ?)", email, entity.GroupMemberStatusPending, userID).
		Updates(map[string]any{
			"user_id": userID,
			"status":  entity.GroupMemberStatusActive,
		})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
