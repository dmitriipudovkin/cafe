package userStorage

import (
	"auth/internal/domain/models"
	"auth/internal/lib/hash"
	"auth/internal/lib/logger"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
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
	sessionStorage, err := New(dbOptions, logger, hasher)

	if err != nil {
		panic(err)
	}

	return sessionStorage
}

func New(dbOptions UserStorageOptions, logger *logger.Logger, hasher *hash.Hasher) (*UserStorage, error) {
	db, err := sql.Open("sqlite3", dbOptions.DBPath)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	// Create a table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT,
			password TEXT,
			is_admin BOOLEAN
		);
	`)

	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	} else {
		logger.Info("Table created or already exists")
	}

	res := &UserStorage{db: db, logger: logger, hasher: hasher}

	// Create admin if not exist
	var adminExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE is_admin = TRUE)").Scan(&adminExists)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	} else if !adminExists {
		logger.Info("Creating admin")
		err = res.CreateUser(db, dbOptions.AdminLogin, dbOptions.AdminPassword, true)

		if err != nil {
			logger.Fatal(err)
			os.Exit(1)
		}
	} else {
		logger.Info("Admin already exists")
	}

	return res, nil
}

func (UserStorage *UserStorage) Stop() {
	UserStorage.db.Close()
}

func (UserStorage *UserStorage) CreateUser(db *sql.DB, name string, password string, isAdmin bool) error {
	hashedPassword, _ := UserStorage.hasher.Hash(password)
	fmt.Println(password, hashedPassword)
	id := uuid.New().String()

	_, err := db.Exec("INSERT INTO users (id, name, password, is_admin) VALUES (?, ?, ?, ?)", id, name, hashedPassword, isAdmin)
	return err
}

func (UserStorage *UserStorage) CheckPassword(name string, password string) (bool, error) {
	_, err := UserStorage.GetUserByCredentials(name, password)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (UserStorage *UserStorage) GetUserByCredentials(name string, password string) (*models.User, error) {
	row := UserStorage.db.QueryRow("SELECT * FROM users WHERE name = ? AND password = ?", name, password)
	fmt.Println(row)
	var user models.User
	err := row.Scan(&user.ID, &user.Name, &user.Password, &user.IsAdmin)
	fmt.Println(user, err)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}
