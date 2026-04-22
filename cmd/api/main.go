package main

import (
	"log"
	"os"

	"github.com/lambdavault/api/internal/infrastructure/config"
	"github.com/lambdavault/api/internal/infrastructure/persistence/sqlite"
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

	if cfg.IsProduction() {
		if err := cfg.Validate(); err != nil {
			return err
		}
	}

	db, err := sqlite.NewDatabase(cfg.Database.Path, cfg.IsProduction())
	if err != nil {
		return err
	}
	defer db.Close()

	userRepo := sqlite.NewUserRepository(db.DB)
	passwordRepo := sqlite.NewPasswordRepository(db.DB)
	hasher := security.NewArgon2Hasher()
	jwtService := security.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiration)

	encryptor, err := security.NewAESEncryptor(cfg.Encryption.Key)
	if err != nil {
		return err
	}

	r := router.New(cfg, userRepo, passwordRepo, jwtService, hasher, encryptor)
	r.Setup()

	log.Printf("🔐 %s starting on port %s [%s]", cfg.App.Name, cfg.App.Port, cfg.App.Env)

	return r.Run()
}
