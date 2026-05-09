package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App          AppConfig
	Database     DatabaseConfig
	JWT          JWTConfig
	Encryption   EncryptionConfig
	RateLimit    RateLimitConfig
	CORS         CORSConfig
	Notification NotificationConfig
}

type AppConfig struct {
	Env  string
	Port string
	Name string
}

type DatabaseConfig struct {
	// Driver is "sqlite" (default) or "postgres".
	Driver string
	// Path is the SQLite file path. Used only when Driver == "sqlite".
	Path string
	// DSN is the Postgres connection string. Used only when Driver == "postgres".
	DSN string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type EncryptionConfig struct {
	Key string
}

type RateLimitConfig struct {
	RPS       float64
	Burst     int
	AuthRPS   float64
	AuthBurst int
}

type CORSConfig struct {
	// Origins is the allow-list. Empty list disables CORS in production. In
	// development we fall back to "*" when this is empty.
	Origins []string
}

type NotificationConfig struct {
	WhatsmiauAPIBaseURL string
	WhatsmiauAPIKey     string
	WhatsmiauInstance   string
	WhatsmiauPhone      string
	Timeout             time.Duration
}

// Insecure values that are explicitly forbidden for JWT_SECRET / ENCRYPTION_KEY.
// We compare against these to make sure operators rotate them before going
// to production.
var forbiddenSecrets = map[string]struct{}{
	"": {},
	"your-super-secret-jwt-key-change-in-production": {},
	"your-32-byte-encryption-key-here":               {},
	"changeme":                                       {},
	"secret":                                         {},
}

const (
	minJWTSecretLength    = 32
	requiredEncryptionLen = 32
)

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) && os.Getenv("APP_ENV") == "" {
			return nil, err
		}
	}

	jwtExpHours, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	rateRPS, _ := strconv.ParseFloat(getEnv("RATE_LIMIT_RPS", "10"), 64)
	rateBurst, _ := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "20"))
	authRPS, _ := strconv.ParseFloat(getEnv("RATE_LIMIT_AUTH_RPS", "1"), 64)
	authBurst, _ := strconv.Atoi(getEnv("RATE_LIMIT_AUTH_BURST", "5"))
	notificationTimeoutSeconds, _ := strconv.Atoi(getEnv("WHATSMIAU_TIMEOUT_SECONDS", "5"))

	cfg := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", getEnv("PORT", "8080")),
			Name: getEnv("APP_NAME", "lambdavault"),
		},
		Database: DatabaseConfig{
			Driver: strings.ToLower(getEnv("DB_DRIVER", "sqlite")),
			Path:   getEnv("DB_PATH", "./data/lambdavault.db"),
			DSN:    getEnv("DB_DSN", os.Getenv("DATABASE_URL")),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", ""),
			Expiration: time.Duration(jwtExpHours) * time.Hour,
		},
		Encryption: EncryptionConfig{
			Key: getEnv("ENCRYPTION_KEY", ""),
		},
		RateLimit: RateLimitConfig{
			RPS:       rateRPS,
			Burst:     rateBurst,
			AuthRPS:   authRPS,
			AuthBurst: authBurst,
		},
		CORS: CORSConfig{
			Origins: parseCSV(getEnv("CORS_ORIGINS", "")),
		},
		Notification: NotificationConfig{
			WhatsmiauAPIBaseURL: getEnv("WHATSMIAU_API_BASE_URL", "https://api.whatsmiau.dev/v2"),
			WhatsmiauAPIKey:     getEnv("WHATSMIAU_API_KEY", ""),
			WhatsmiauInstance:   getEnv("WHATSMIAU_INSTANCE", ""),
			WhatsmiauPhone:      getEnv("WHATSMIAU_PHONE", ""),
			Timeout:             time.Duration(notificationTimeoutSeconds) * time.Second,
		},
	}

	return cfg, nil
}

func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// Validate enforces the security invariants required to safely run the
// service. It is always executed at boot — even in development — so that
// dangerous defaults never make it into a deployed binary.
//
// In production it additionally requires explicit CORS_ORIGINS and a
// non-default postgres setup is encouraged (warned via a separate check).
func (c *Config) Validate() error {
	if _, forbidden := forbiddenSecrets[c.JWT.Secret]; forbidden {
		return ErrMissingJWTSecret
	}
	if len(c.JWT.Secret) < minJWTSecretLength {
		return fmt.Errorf("%w: got %d chars, need >= %d", ErrWeakJWTSecret, len(c.JWT.Secret), minJWTSecretLength)
	}

	if _, forbidden := forbiddenSecrets[c.Encryption.Key]; forbidden {
		return ErrMissingEncryptionKey
	}
	if len(c.Encryption.Key) != requiredEncryptionLen {
		return fmt.Errorf("%w: got %d bytes, need exactly %d", ErrInvalidEncryptionKey, len(c.Encryption.Key), requiredEncryptionLen)
	}

	switch c.Database.Driver {
	case "sqlite", "":
		if c.Database.Path == "" {
			return ErrMissingDBPath
		}
	case "postgres":
		if strings.TrimSpace(c.Database.DSN) == "" {
			return ErrMissingDBDSN
		}
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedDBDriver, c.Database.Driver)
	}

	if c.IsProduction() {
		if len(c.CORS.Origins) == 0 {
			return ErrMissingCORSOrigins
		}
		for _, o := range c.CORS.Origins {
			if o == "*" {
				return ErrWildcardCORSInProduction
			}
		}
		if c.Database.Driver == "sqlite" || c.Database.Driver == "" {
			// SQLite on container disks is not durable across redeploys.
			// Refuse to boot in production unless the operator opts in.
			if os.Getenv("ALLOW_SQLITE_IN_PRODUCTION") != "true" {
				return ErrSQLiteInProduction
			}
		}
	}

	return nil
}

func parseCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
