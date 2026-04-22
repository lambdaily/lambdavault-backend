package entity

import (
	"time"

	"github.com/google/uuid"
)

type Password struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID            uuid.UUID `gorm:"type:uuid;index;not null"`
	SiteName          string    `gorm:"not null"`
	SiteURL           string
	Username          string    `gorm:"not null"`
	EncryptedPassword string    `gorm:"not null"`
	IV                string    `gorm:"not null"`
	Notes             string
	Category          string
	CreatedAt         time.Time
	UpdatedAt         time.Time

	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
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
