package config

import "errors"

var (
	ErrMissingJWTSecret     = errors.New("JWT_SECRET is required")
	ErrInvalidEncryptionKey = errors.New("ENCRYPTION_KEY must be exactly 32 bytes")
)
