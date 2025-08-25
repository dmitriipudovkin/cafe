package userStorage

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator struct {
	db *sql.DB
	m  *migrate.Migrate
}

func NewMigrator(db *sql.DB) (*Migrator, error) {
	// Инициализируем драйвер для миграций
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate driver: %w", err)
	}

	// Получаем абсолютный путь к миграциям
	migrationsPath, err := filepath.Abs("./internal/storage/user_storage/migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"sqlite",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{
		db: db,
		m:  m,
	}, nil
}

func (m *Migrator) Up() error {
	if err := m.m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	fmt.Println("Migrations applied successfully")
	return nil
}

func (m *Migrator) Down() error {
	if err := m.m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to revert migrations: %w", err)
	}
	fmt.Println("Migrations reverted successfully")
	return nil
}

func (m *Migrator) Close() error {
	if sourceErr, dbErr := m.m.Close(); sourceErr != nil || dbErr != nil {
		return fmt.Errorf("failed to close migrator with errors: source: %w, db: %w", sourceErr, dbErr)
	}
	return m.db.Close()
}
