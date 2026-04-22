package dto

import "github.com/google/uuid"

type RegisterRequest struct {
	Email          string `json:"email" validate:"required,email"`
	MasterPassword string `json:"master_password" validate:"required,min=8,max=128"`
}

type LoginRequest struct {
	Email          string `json:"email" validate:"required,email"`
	MasterPassword string `json:"master_password" validate:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}
