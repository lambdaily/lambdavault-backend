package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email              string    `gorm:"uniqueIndex;not null"`
	Phone              string    `gorm:"index"`
	MasterPasswordHash string    `gorm:"not null"`
	Salt               string    `gorm:"not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewUser(email, phone, masterPasswordHash, salt string) *User {
	return &User{
		ID:                 uuid.New(),
		Email:              email,
		Phone:              phone,
		MasterPasswordHash: masterPasswordHash,
		Salt:               salt,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func (u *User) TableName() string {
	return "users"
}
