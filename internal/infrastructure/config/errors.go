package config

import "errors"

var (
	ErrMissingJWTSecret         = errors.New("JWT_SECRET is required and must not be a default value")
	ErrWeakJWTSecret            = errors.New("JWT_SECRET is too short")
	ErrMissingEncryptionKey     = errors.New("ENCRYPTION_KEY is required and must not be a default value")
	ErrInvalidEncryptionKey     = errors.New("ENCRYPTION_KEY must be exactly 32 bytes")
	ErrMissingDBPath            = errors.New("DB_PATH is required when DB_DRIVER=sqlite")
	ErrMissingDBDSN             = errors.New("DB_DSN (or DATABASE_URL) is required when DB_DRIVER=postgres")
	ErrUnsupportedDBDriver      = errors.New("unsupported DB_DRIVER")
	ErrMissingCORSOrigins       = errors.New("CORS_ORIGINS must be set in production")
	ErrWildcardCORSInProduction = errors.New("CORS_ORIGINS cannot contain '*' in production")
	ErrSQLiteInProduction       = errors.New("SQLite is not durable in production: set DB_DRIVER=postgres or ALLOW_SQLITE_IN_PRODUCTION=true if you really know what you are doing")
)
