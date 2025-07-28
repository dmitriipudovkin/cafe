package auth

import (
	"fmt"
	"time"

	"cafe_main/internal/auth/hash"
	"cafe_main/internal/auth/token"
	userStorage "cafe_main/internal/auth/user_storage"
	"cafe_main/internal/logger"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	fmt.Println("Im a dummy!")

	c.Next()
}

type AuthServiceUserStorage interface {
	GetUserByCredentials(name string, password string) (*userStorage.User, error)
}

// TO DO Добавить сессии
type AuthServiceSessionStorage interface {
	CreateSession() error
	DeleteSession() error
}

type AuthServiceInterface interface {
	Login(name string, password string)
}

type Tokenizer interface {
	GetToken(claims map[string]interface{}) (string, error)
}

type AuthService struct {
	authStorage AuthServiceUserStorage
	tokenizer   Tokenizer
	hasher      *hash.Hasher
}

func NewAuthService(dbPath string, logger logger.Logger, secret string) (AuthService, error) {
	hasher := hash.NewHasher(secret)
	tokenizer := token.NewTokenizer(secret)

	authStorage, err := userStorage.InitUserStorage(dbPath, logger, hasher)

	authService := AuthService{
		authStorage,
		tokenizer,
		hasher,
	}

	return authService, err
}

func (as *AuthService) Login(name string, password string) (string, error) {
	hashedPassword, err := as.hasher.Hash(password)
	if err != nil {
		return "", err
	}

	user, err := as.authStorage.GetUserByCredentials(name, hashedPassword)

	if err != nil {
		return "", err
	}

	claim := map[string]any{
		"username": user.Name,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}

	token, err := as.tokenizer.GetToken(claim)

	if err != nil {
		return "", err
	}

	return token, err
}
