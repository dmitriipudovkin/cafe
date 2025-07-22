package storage

import (
	"cafe_main/internal/auth/crypto"
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// To Do Переписать на struct
// To Do вынести в конфиг путь до бд

type AuthStorage struct {
	db     *sql.DB
	logger *logrus.Logger
}

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

func CreateUser(db *sql.DB, name string, password string, isAdmin bool) error {
	hashedPassword, _ := crypto.GetHash(password)

	_, err := db.Exec("INSERT INTO users (name, password, is_admin) VALUES (?, ?, ?)", name, hashedPassword, isAdmin)
	return err
}

func (AuthStorage *AuthStorage) GetUserByCredentials(name string, password string) (*User, error) {
	row := AuthStorage.db.QueryRow("SELECT * FROM users WHERE name = ? AND password = ?", name, password)

	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Password, &user.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func NewAuthStorage(dbPath string, logger *logrus.Logger) (*AuthStorage, error) {
	// TO DO путь считается от корня(
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}
	defer db.Close()

	// Create a table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
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

	// Create admin if not exist
	var adminExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE is_admin = TRUE)").Scan(&adminExists)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	} else if !adminExists {
		logger.Info("Creating admin")
		err = CreateUser(db, "admin", "admin", true)

		if err != nil {
			logger.Fatal(err)
			os.Exit(1)
		}
	} else {
		logger.Info("Admin already exists")
	}

	return &AuthStorage{db: db, logger: logger}, nil
}
