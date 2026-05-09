package entity

import (
	"time"

	"github.com/google/uuid"
)

type Password struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;index;not null"`
	// GroupID is a deprecated legacy field from the old model where a password
	// could belong to exactly one group. New code uses PasswordGroupShare.
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

	User   *User                 `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Group  *PasswordGroup        `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
	Shares []*PasswordGroupShare `gorm:"foreignKey:PasswordID;constraint:OnDelete:CASCADE"`
}

// IsShared reports whether the password is shared with at least one group.
//
// In the new model that information lives in PasswordGroupShare rows. We also
// check the deprecated GroupID field so older rows behave correctly before the
// one-time migration clears it.
func (p *Password) IsShared() bool {
	return p.GroupID != nil || len(p.Shares) > 0
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

// Shared passwords are modeled via PasswordGroupShare. To create one,
// create a regular Password and then create one or more shares for it.

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
