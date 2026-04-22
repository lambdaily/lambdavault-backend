package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/lambdavault/api/internal/infrastructure/security"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	email := flag.String("email", "", "user email")
	dbPath := flag.String("db", "./data/lambdavault.db", "sqlite database path")
	password := flag.String("password", "", "new master password")
	passwordEnv := flag.String("password-env", "", "env var name containing new master password")
	flag.Parse()

	if *email == "" {
		exitErr("--email is required")
	}

	if *password == "" && *passwordEnv != "" {
		*password = os.Getenv(*passwordEnv)
	}

	if *password == "" {
		exitErr("new password is required")
	}

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		exitErr("failed to open db: %v", err)
	}
	defer db.Close()

	var exists int
	if err := db.QueryRow("SELECT COUNT(1) FROM users WHERE email = ?", *email).Scan(&exists); err != nil {
		exitErr("failed to check user: %v", err)
	}
	if exists == 0 {
		exitErr("user not found: %s", *email)
	}

	hasher := security.NewArgon2Hasher()
	hash, salt, err := hasher.Hash(*password)
	if err != nil {
		exitErr("failed to hash password: %v", err)
	}

	res, err := db.Exec(
		"UPDATE users SET master_password_hash = ?, salt = ?, updated_at = CURRENT_TIMESTAMP WHERE email = ?",
		hash,
		salt,
		*email,
	)
	if err != nil {
		exitErr("failed to update password: %v", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		exitErr("failed to confirm update: %v", err)
	}
	if rows == 0 {
		exitErr("no rows updated for user: %s", *email)
	}

	fmt.Printf("Password updated successfully for %s\n", *email)
}

func exitErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
