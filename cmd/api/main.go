package main

import (
	"log"
	"os"

	"github.com/lambdavault/api/internal/infrastructure/config"
	"github.com/lambdavault/api/internal/infrastructure/persistence/database"
	"github.com/lambdavault/api/internal/infrastructure/security"
	"github.com/lambdavault/api/internal/interfaces/http/router"
)

func main() {
	if err := run(); err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Always validate. Insecure defaults must never reach a deployed binary,
	// not even by accident in a "development" environment.
	if err := cfg.Validate(); err != nil {
		return err
	}

	dsn := cfg.Database.Path
	if cfg.Database.Driver == "postgres" {
		dsn = cfg.Database.DSN
	}

	db, err := database.New(database.Config{
		Driver:       database.Driver(cfg.Database.Driver),
		DSN:          dsn,
		IsProduction: cfg.IsProduction(),
	})
	if err != nil {
		return err
	}
	defer db.Close()

	userRepo := database.NewUserRepository(db.DB)
	passwordRepo := database.NewPasswordRepository(db.DB)
	groupRepo := database.NewPasswordGroupRepository(db.DB)
	hasher := security.NewArgon2Hasher()
	jwtService := security.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiration)

	encryptor, err := security.NewAESEncryptor(cfg.Encryption.Key)
	if err != nil {
		return err
	}

	r := router.New(cfg, userRepo, passwordRepo, groupRepo, jwtService, hasher, encryptor, db)
	r.Setup()

	log.Printf("🔐 %s starting on port %s [%s] driver=%s", cfg.App.Name, cfg.App.Port, cfg.App.Env, cfg.Database.Driver)

	return r.Run()
}
