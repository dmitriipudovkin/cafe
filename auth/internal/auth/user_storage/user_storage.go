package userStorage

import (
	"cafe_main/internal/auth/hash"
	"cafe_main/internal/logger"
	"database/sql"
	"errors"
	"os"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type UserStorage struct {
	db     *sql.DB
	logger logger.Logger
	hasher hash.HasherInterface
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

func (UserStorage *UserStorage) CreateUser(db *sql.DB, name string, password string, isAdmin bool) error {
	hashedPassword, _ := UserStorage.hasher.Hash(password)
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

func (UserStorage *UserStorage) GetUserByCredentials(name string, password string) (*User, error) {
	row := UserStorage.db.QueryRow("SELECT * FROM users WHERE name = ? AND password = ?", name, password)

	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Password, &user.IsAdmin)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

type UserStorageOptions struct {
	DBPath        string
	AdminLogin    string
	AdminPassword string
}

func InitUserStorage(dbOptions UserStorageOptions, logger logger.Logger, hasher hash.HasherInterface) (*UserStorage, error) {
	// TO DO путь считается от корня(
	db, err := sql.Open("sqlite3", dbOptions.DBPath)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}
	// defer db.Close()

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
