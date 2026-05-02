package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/lambdavault/api/internal/domain/entity"
)

// PasswordGroupRepository persists password groups and their memberships.
type PasswordGroupRepository interface {
	Create(ctx context.Context, group *entity.PasswordGroup) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.PasswordGroup, error)
	Update(ctx context.Context, group *entity.PasswordGroup) error
	Delete(ctx context.Context, id uuid.UUID) error

	// FindAccessibleByUserID returns all groups the user owns or is an active
	// member of.
	FindAccessibleByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.PasswordGroup, error)

	// AddMember inserts a new membership row.
	AddMember(ctx context.Context, member *entity.GroupMember) error
	UpdateMember(ctx context.Context, member *entity.GroupMember) error
	RemoveMember(ctx context.Context, memberID uuid.UUID) error

	FindMemberByID(ctx context.Context, memberID uuid.UUID) (*entity.GroupMember, error)
	FindMemberByGroupAndUser(ctx context.Context, groupID, userID uuid.UUID) (*entity.GroupMember, error)
	FindMemberByGroupAndEmail(ctx context.Context, groupID uuid.UUID, email string) (*entity.GroupMember, error)
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]*entity.GroupMember, error)

	// ActivatePendingForEmail links every pending invitation matching the email
	// to the newly registered user. Returns the number of memberships activated.
	ActivatePendingForEmail(ctx context.Context, email string, userID uuid.UUID) (int64, error)
}
