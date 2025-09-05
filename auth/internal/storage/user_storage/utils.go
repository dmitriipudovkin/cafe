package userStorage

import (
	errfmt "auth/internal/lib/errors"
	"auth/internal/lib/logger"
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func createParentDir(dbPath string) error {
	const op = "userStorage.createParentDir"
	formatter := errfmt.NewOpErrorFormatter(op)
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, FileMod); err != nil {
		return formatter.Format(err, "failed to create directory: %s", dir)
	}
	return nil
}

func runMigrations(db *sql.DB, logger *logger.Logger) error {
	const op = "userStorage.runMigrations"
	formatter := errfmt.NewOpErrorFormatter(op)

	migrator, err := NewMigrator(db, logger)
	if err != nil {
		return formatter.Format(err, "failed to create migrator")
	}

	// Apply migrations
	if err := migrator.Up(); err != nil {
		return formatter.Format(err, "failed to apply migrations")
	}

	// Close migrator after applying migrations
	if err := migrator.Close(); err != nil {
		return formatter.Format(err, "failed to close migrator")
	}

	return nil
}

func createAdminIfNotExists(us *UserStorage, dbOptions UserStorageOptions, logger *logger.Logger) error {
	const op = "userStorage.createAdminIfNotExists"
	formatter := errfmt.NewOpErrorFormatter(op)

	var adminExists bool
	err := us.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE is_admin = TRUE)").Scan(&adminExists)
	if err != nil {
		return formatter.Format(err)
	}

	if !adminExists {
		logger.Info("Creating admin")
		err = us.CreateUser(us.db, dbOptions.AdminLogin, dbOptions.AdminPassword, true)
		if err != nil {
			return formatter.Format(err)
		}
	} else {
		logger.Info("Admin already exists")
	}

	return nil
}
