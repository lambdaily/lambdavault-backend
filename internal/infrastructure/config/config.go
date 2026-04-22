package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	Encryption EncryptionConfig
	RateLimit  RateLimitConfig
}

type AppConfig struct {
	Env  string
	Port string
	Name string
}

type DatabaseConfig struct {
	Path string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type EncryptionConfig struct {
	Key string
}

type RateLimitConfig struct {
	RPS   float64
	Burst int
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		if os.Getenv("APP_ENV") == "" {
			return nil, err
		}
	}

	jwtExpHours, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	rateRPS, _ := strconv.ParseFloat(getEnv("RATE_LIMIT_RPS", "10"), 64)
	rateBurst, _ := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "20"))

	return &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", getEnv("PORT", "8080")),
			Name: getEnv("APP_NAME", "lambdavault"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./data/lambdavault.db"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", ""),
			Expiration: time.Duration(jwtExpHours) * time.Hour,
		},
		Encryption: EncryptionConfig{
			Key: getEnv("ENCRYPTION_KEY", ""),
		},
		RateLimit: RateLimitConfig{
			RPS:   rateRPS,
			Burst: rateBurst,
		},
	}, nil
}

func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return ErrMissingJWTSecret
	}
	if len(c.Encryption.Key) != 32 {
		return ErrInvalidEncryptionKey
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
