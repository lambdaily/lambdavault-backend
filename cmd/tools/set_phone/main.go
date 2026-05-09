package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/lambdavault/api/internal/domain/entity"
)

func main() {
	email := flag.String("email", "", "user email")
	phone := flag.String("phone", "", "phone number to store, e.g. +595976511816")
	pgDSN := flag.String("postgres", os.Getenv("DATABASE_URL"), "destination Postgres DSN (defaults to $DATABASE_URL)")
	sqlitePath := flag.String("sqlite", "", "optional SQLite path instead of Postgres")
	flag.Parse()

	*email = strings.ToLower(strings.TrimSpace(*email))
	*phone = strings.TrimSpace(*phone)
	if *email == "" {
		exit("--email is required")
	}
	if *phone == "" {
		exit("--phone is required")
	}

	db, err := openDB(*pgDSN, *sqlitePath)
	if err != nil {
		exit("failed to open database: %v", err)
	}

	if err := db.AutoMigrate(&entity.User{}); err != nil {
		exit("failed to migrate users table: %v", err)
	}

	var user entity.User
	if err := db.Where("email = ?", *email).First(&user).Error; err != nil {
		exit("failed to find user: %v", err)
	}
	user.Phone = *phone
	if err := db.Save(&user).Error; err != nil {
		exit("failed to update phone: %v", err)
	}

	log.Printf("✅ updated phone for %s -> %s", user.Email, user.Phone)
}

func openDB(pgDSN, sqlitePath string) (*gorm.DB, error) {
	gormCfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}
	if strings.TrimSpace(sqlitePath) != "" {
		return gorm.Open(sqlite.Open(sqlitePath), gormCfg)
	}
	if strings.TrimSpace(pgDSN) == "" {
		return nil, os.ErrInvalid
	}
	return gorm.Open(postgres.Open(pgDSN), gormCfg)
}

func exit(format string, args ...any) {
	log.Printf("error: "+format, args...)
	os.Exit(1)
}
