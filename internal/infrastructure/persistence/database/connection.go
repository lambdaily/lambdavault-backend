package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/lambdavault/api/internal/domain/entity"
)

// Driver represents a supported database backend.
type Driver string

const (
	DriverSQLite   Driver = "sqlite"
	DriverPostgres Driver = "postgres"
)

// Config encapsulates everything required to open a database connection.
//
// For SQLite only DSN (the file path) is required. For Postgres DSN must be
// a libpq-compatible connection string, e.g.:
//
//	postgres://user:pass@host:5432/dbname?sslmode=require
type Config struct {
	Driver       Driver
	DSN          string
	IsProduction bool
}

type Database struct {
	DB     *gorm.DB
	driver Driver
}

// New opens a database connection using the given config and runs the GORM
// auto-migrations for all domain entities.
func New(cfg Config) (*Database, error) {
	logLevel := logger.Info
	if cfg.IsProduction {
		logLevel = logger.Warn
	}

	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	var dialector gorm.Dialector
	switch cfg.Driver {
	case DriverSQLite, "":
		if err := ensureDir(cfg.DSN); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
		dialector = sqlite.Open(cfg.DSN)
	case DriverPostgres:
		if strings.TrimSpace(cfg.DSN) == "" {
			return nil, fmt.Errorf("postgres DSN is required when DB_DRIVER=postgres")
		}
		dialector = postgres.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}

	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Database{DB: db, driver: cfg.Driver}, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.User{},
		&entity.PasswordGroup{},
		&entity.GroupMember{},
		&entity.Password{},
	)
}

func ensureDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping verifies the database is reachable. Useful for /ready healthchecks.
func (d *Database) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
