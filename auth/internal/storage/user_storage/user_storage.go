package userStorage

import (
	"auth/internal/domain/models"
	errfmt "auth/internal/lib/errors"
	"auth/internal/lib/hash"
	"auth/internal/lib/logger"
	"database/sql"
	"errors"
	"os"

	"github.com/google/uuid"
	// _ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

type UserStorage struct {
	db     *sql.DB
	logger *logger.Logger
	hasher *hash.Hasher
}

type UserStorageOptions struct {
	DBPath        string
	AdminLogin    string
	AdminPassword string
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func MustRun(dbOptions UserStorageOptions, logger *logger.Logger, hasher *hash.Hasher) *UserStorage {
	const op = "userStorage.MustRun"
	formatter := errfmt.NewOpErrorFormatter(op)
	sessionStorage, err := New(dbOptions, logger, hasher)

	if err != nil {
		panic(formatter.Format(err))
	}

	return sessionStorage
}

const FileMod = 0o755

func New(dbOptions UserStorageOptions, logger *logger.Logger, hasher *hash.Hasher) (*UserStorage, error) {
	const op = "userStorage.New"
	formatter := errfmt.NewOpErrorFormatter(op)

	// Ensure parent directory exists
	if err := createParentDir(dbOptions.DBPath); err != nil {
		logger.Fatal(formatter.Format(err))
		os.Exit(1)
	}

	db, err := sql.Open("sqlite", dbOptions.DBPath)
	if err != nil {
		logger.Fatal(formatter.Format(err, "failed to open database: %s", dbOptions.DBPath))
		os.Exit(1)
	}

	if err := runMigrations(db, logger); err != nil {
		logger.Fatal(formatter.Format(err))
		os.Exit(1)
	}

	// Close initial connection as it might have been closed by migrator
	if err := db.Close(); err != nil {
		logger.Fatal(formatter.Format(err, "failed to close database"))
		os.Exit(1)
	}

	// Open new connection for UserStorage
	db, err = sql.Open("sqlite", dbOptions.DBPath)
	if err != nil {
		logger.Fatal(formatter.Format(err, "failed to reopen database: %s", dbOptions.DBPath))
		os.Exit(1)
	}

	res := &UserStorage{db: db, logger: logger, hasher: hasher}

	if err := createAdminIfNotExists(res, dbOptions, logger); err != nil {
		logger.Fatal(formatter.Format(err))
		os.Exit(1)
	}

	return res, nil
}

func (us *UserStorage) Stop() error {
	us.logger.Info("Stopping user storage")
	return us.db.Close()
}

func (us *UserStorage) CreateUser(db *sql.DB, name string, password string, isAdmin bool) error {
	const op = "userStorage.CreateUser"
	formatter := errfmt.NewOpErrorFormatter(op)
	hashedPassword, _ := us.hasher.Hash(password)
	id := uuid.New().String()

	_, err := db.Exec(
		"INSERT INTO users (id, name, password, is_admin) VALUES (?, ?, ?, ?)",
		id,
		name,
		hashedPassword,
		isAdmin,
	)
	if err != nil {
		return formatter.Format(err)
	}
	return nil
}

func (us *UserStorage) CheckPassword(name string, password string) (bool, error) {
	const op = "userStorage.CheckPassword"
	formatter := errfmt.NewOpErrorFormatter(op)

	_, err := us.GetUserByCredentials(name, password)
	if err != nil {
		return false, formatter.Format(err)
	}
	return true, nil
}

func (us *UserStorage) GetUserByCredentials(name string, password string) (*models.User, error) {
	const op = "userStorage.GetUserByCredentials"
	formatter := errfmt.NewOpErrorFormatter(op)

	row := us.db.QueryRow("SELECT * FROM users WHERE name = ? AND password = ?", name, password)
	var user models.User
	err := row.Scan(&user.ID, &user.Name, &user.Password, &user.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, formatter.Format(ErrInvalidCredentials)
		}

		return nil, formatter.Format(err)
	}

	return &user, nil
}
