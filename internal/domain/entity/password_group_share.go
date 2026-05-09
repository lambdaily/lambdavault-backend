package entity

import (
	"time"

	"github.com/google/uuid"
)

// PasswordGroupShare grants a group access to a password owned by a user.
//
// The underlying password remains a single canonical record in `passwords`.
// Sharing/unsharing is modeled by adding/removing rows from this table.
type PasswordGroupShare struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	PasswordID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_password_group_share"`
	GroupID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_password_group_share"`
	SharedByID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time

	Password *Password      `gorm:"foreignKey:PasswordID;constraint:OnDelete:CASCADE"`
	Group    *PasswordGroup `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
	SharedBy *User          `gorm:"foreignKey:SharedByID;constraint:OnDelete:CASCADE"`
}

func NewPasswordGroupShare(passwordID, groupID, sharedByID uuid.UUID) *PasswordGroupShare {
	return &PasswordGroupShare{
		ID:         uuid.New(),
		PasswordID: passwordID,
		GroupID:    groupID,
		SharedByID: sharedByID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func (s *PasswordGroupShare) TableName() string {
	return "password_group_shares"
}
