package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreatePasswordRequest struct {
	SiteName string `json:"site_name" validate:"required,min=1,max=255"`
	SiteURL  string `json:"site_url" validate:"omitempty,max=500"`
	Username string `json:"username" validate:"required,min=1,max=255"`
	Password string `json:"password" validate:"required,min=1,max=500"`
	Notes    string `json:"notes" validate:"omitempty,max=1000"`
	Category string `json:"category" validate:"omitempty,max=100"`
}

type UpdatePasswordRequest struct {
	SiteName string `json:"site_name" validate:"required,min=1,max=255"`
	SiteURL  string `json:"site_url" validate:"omitempty,max=500"`
	Username string `json:"username" validate:"required,min=1,max=255"`
	Password string `json:"password" validate:"required,min=1,max=500"`
	Notes    string `json:"notes" validate:"omitempty,max=1000"`
	Category string `json:"category" validate:"omitempty,max=100"`
}

type PasswordResponse struct {
	ID        uuid.UUID `json:"id"`
	SiteName  string    `json:"site_name"`
	SiteURL   string    `json:"site_url"`
	Username  string    `json:"username"`
	Notes     string    `json:"notes"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PasswordWithSecretResponse struct {
	PasswordResponse
	Password string `json:"password"`
}

type PasswordListResponse struct {
	Passwords []PasswordResponse `json:"passwords"`
	Total     int                `json:"total"`
}
