// recover_sqlite_passwords imports personal vault passwords from a legacy
// SQLite database into an existing Postgres account matched by email.
//
// This is meant for the recovery case where:
//  1. a user already existed in the old SQLite DB,
//  2. a new account with the same email was created in Postgres, and
//  3. we want to reattach the old personal passwords to the new user ID.
//
// Usage:
//
//	go run ./cmd/tools/recover_sqlite_passwords \
//	    --sqlite ./data/lambdavault.db \
//	    --email you@example.com \
//	    --postgres "postgres://user:pass@host:5432/lambdavault?sslmode=require"
package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/lambdavault/api/internal/domain/entity"
)

func main() {
	sqlitePath := flag.String("sqlite", "./data/lambdavault.db", "path to source SQLite database file")
	email := flag.String("email", "", "email to recover passwords for")
	pgDSN := flag.String("postgres", os.Getenv("DATABASE_URL"), "destination Postgres DSN (defaults to $DATABASE_URL)")
	dryRun := flag.Bool("dry-run", false, "count rows but do not write to Postgres")
	flag.Parse()

	*email = strings.ToLower(strings.TrimSpace(*email))
	if *email == "" {
		exit("--email is required")
	}
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

	if err := dst.AutoMigrate(&entity.User{}, &entity.Password{}); err != nil {
		exit("failed to ensure postgres schema: %v", err)
	}

	srcUser, err := findUserByEmail(src, *email)
	if err != nil {
		exit("source sqlite user lookup failed: %v", err)
	}
	if srcUser == nil {
		exit("user %s not found in sqlite", *email)
	}

	dstUser, err := findUserByEmail(dst, *email)
	if err != nil {
		exit("destination postgres user lookup failed: %v", err)
	}
	if dstUser == nil {
		exit("user %s not found in postgres; create the account first", *email)
	}

	var passwords []entity.Password
	if err := src.Where("user_id = ?", srcUser.ID).Order("created_at ASC").Find(&passwords).Error; err != nil {
		exit("failed to read sqlite passwords: %v", err)
	}

	log.Printf("→ sqlite user id: %s", srcUser.ID)
	log.Printf("→ postgres user id: %s", dstUser.ID)
	log.Printf("→ passwords to recover: %d", len(passwords))

	for i := range passwords {
		passwords[i].UserID = dstUser.ID
	}

	if *dryRun || len(passwords) == 0 {
		if *dryRun {
			log.Printf("✅ dry-run complete (no changes written)")
		} else {
			log.Printf("✅ nothing to import")
		}
		return
	}

	if err := dst.Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(passwords, 200).Error
	}); err != nil {
		exit("failed to import passwords: %v", err)
	}

	log.Printf("✅ password recovery complete")
	log.Printf("note: recovered passwords are decryptable only if Postgres is using the same ENCRYPTION_KEY as the old SQLite-backed app")
}

func findUserByEmail(db *gorm.DB, email string) (*entity.User, error) {
	var user entity.User
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func exit(format string, args ...any) {
	log.Printf("error: "+format, args...)
	os.Exit(1)
}
