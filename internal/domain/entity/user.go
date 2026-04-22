package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email              string    `gorm:"uniqueIndex;not null"`
	MasterPasswordHash string    `gorm:"not null"`
	Salt               string    `gorm:"not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewUser(email, masterPasswordHash, salt string) *User {
	return &User{
		ID:                 uuid.New(),
		Email:              email,
		MasterPasswordHash: masterPasswordHash,
		Salt:               salt,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func (u *User) TableName() string {
	return "users"
}
