package entity

import (
	"time"

	"github.com/google/uuid"
)

type Password struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID            uuid.UUID  `gorm:"type:uuid;index;not null"`
	GroupID           *uuid.UUID `gorm:"type:uuid;index"`
	SiteName          string     `gorm:"not null"`
	SiteURL           string
	Username          string `gorm:"not null"`
	EncryptedPassword string `gorm:"not null"`
	IV                string `gorm:"not null"`
	Notes             string
	Category          string
	CreatedAt         time.Time
	UpdatedAt         time.Time

	User  *User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Group *PasswordGroup `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
}

// IsShared reports whether the password belongs to a shared group rather than
// to a single user's private vault.
func (p *Password) IsShared() bool {
	return p.GroupID != nil
}

func NewPassword(userID uuid.UUID, siteName, siteURL, username, encryptedPassword, iv, notes, category string) *Password {
	return &Password{
		ID:                uuid.New(),
		UserID:            userID,
		SiteName:          siteName,
		SiteURL:           siteURL,
		Username:          username,
		EncryptedPassword: encryptedPassword,
		IV:                iv,
		Notes:             notes,
		Category:          category,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// NewGroupPassword creates a password that belongs to a shared group. UserID
// records the creator (for auditing); GroupID controls access.
func NewGroupPassword(creatorID, groupID uuid.UUID, siteName, siteURL, username, encryptedPassword, iv, notes, category string) *Password {
	p := NewPassword(creatorID, siteName, siteURL, username, encryptedPassword, iv, notes, category)
	p.GroupID = &groupID
	return p
}

func (p *Password) TableName() string {
	return "passwords"
}

func (p *Password) Update(siteName, siteURL, username, encryptedPassword, iv, notes, category string) {
	p.SiteName = siteName
	p.SiteURL = siteURL
	p.Username = username
	p.EncryptedPassword = encryptedPassword
	p.IV = iv
	p.Notes = notes
	p.Category = category
	p.UpdatedAt = time.Now()
}
