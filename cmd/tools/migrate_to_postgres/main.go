// migrate_to_postgres copies all data from a local SQLite vault into a
// Postgres database. Designed to be run ONCE before flipping a deployed
// instance from sqlite to postgres.
//
// Usage:
//
//	go run ./cmd/tools/migrate_to_postgres \
//	    --sqlite ./data/lambdavault.db \
//	    --postgres "postgres://user:pass@host:5432/lambdavault?sslmode=require"
//
// The destination Postgres can be empty — the tool will run AutoMigrate to
// create the tables. The migration is idempotent: re-running it skips rows
// that already exist (matched by primary key).
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/lambdavault/api/internal/domain/entity"
	dbpkg "github.com/lambdavault/api/internal/infrastructure/persistence/database"
)

func main() {
	sqlitePath := flag.String("sqlite", "./data/lambdavault.db", "path to source SQLite database file")
	pgDSN := flag.String("postgres", os.Getenv("DATABASE_URL"), "destination Postgres DSN (defaults to $DATABASE_URL)")
	dryRun := flag.Bool("dry-run", false, "count rows but do not write to Postgres")
	flag.Parse()

	if *pgDSN == "" {
		exit("--postgres (or DATABASE_URL env) is required")
	}
	if _, err := os.Stat(*sqlitePath); err != nil {
		exit("source sqlite file not found: %v", err)
	}

	src, err := gorm.Open(sqlite.Open(*sqlitePath), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		exit("failed to open sqlite: %v", err)
	}

	dst, err := gorm.Open(postgres.Open(*pgDSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		exit("failed to open postgres: %v", err)
	}

	log.Printf("→ ensuring postgres schema")
	if err := dst.AutoMigrate(
		&entity.User{},
		&entity.PasswordGroup{},
		&entity.GroupMember{},
		&entity.Password{},
		&entity.PasswordGroupShare{},
	); err != nil {
		exit("failed to migrate postgres schema: %v", err)
	}

	// Order matters: respect foreign keys. Some older SQLite vaults predate
	// shared groups, so password_groups / group_members may legitimately be
	// absent in the source DB and should be skipped.
	if err := copyTable[entity.User](src, dst, "users", *dryRun, true); err != nil {
		exit("%v", err)
	}
	if err := copyTable[entity.PasswordGroup](src, dst, "password_groups", *dryRun, false); err != nil {
		exit("%v", err)
	}
	if err := copyTable[entity.GroupMember](src, dst, "group_members", *dryRun, false); err != nil {
		exit("%v", err)
	}
	if err := copyTable[entity.Password](src, dst, "passwords", *dryRun, true); err != nil {
		exit("%v", err)
	}
	if err := copyTable[entity.PasswordGroupShare](src, dst, "password_group_shares", *dryRun, false); err != nil {
		exit("%v", err)
	}

	if !*dryRun {
		if err := dbpkg.BackfillPasswordGroupShares(dst); err != nil {
			exit("failed to backfill password-group shares: %v", err)
		}
	}

	if *dryRun {
		log.Printf("✅ dry-run complete (no changes written)")
	} else {
		log.Printf("✅ migration complete")
	}
}

// copyTable streams every row of the given entity from src to dst, inserting
// in batches and skipping primary-key conflicts so the operation is safe to
// retry.
func copyTable[T any](src, dst *gorm.DB, name string, dryRun bool, required bool) error {
	if !src.Migrator().HasTable(name) {
		if required {
			return fmt.Errorf("source sqlite table %s not found", name)
		}
		log.Printf("→ %s: source table missing, skipping (legacy schema)", name)
		return nil
	}

	var rows []T
	if err := src.Find(&rows).Error; err != nil {
		return fmt.Errorf("read %s from sqlite: %w", name, err)
	}
	log.Printf("→ %s: %d rows", name, len(rows))
	if dryRun || len(rows) == 0 {
		return nil
	}
	// Insert in batches of 200 with ON CONFLICT DO NOTHING semantics so the
	// command is idempotent.
	if err := dst.Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(rows, 200).Error; err != nil {
		return fmt.Errorf("write %s to postgres: %w", name, err)
	}
	return nil
}

func exit(format string, args ...any) {
	log.Printf("error: "+format, args...)
	os.Exit(1)
}
