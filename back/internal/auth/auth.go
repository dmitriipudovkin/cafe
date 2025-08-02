package auth

import (
	"errors"
	"fmt"
	"time"

	"cafe_main/internal/auth/hash"
	sessionStorage "cafe_main/internal/auth/session_storage"
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
	CreateSession(id string, accessToken string, refreshToken string) error
	GetAccessToken(id string) (string, error)
	GetRefreshToken(id string) (string, error)
}

type AuthServiceInterface interface {
	Login(name string, password string)
}

type Tokenizer interface {
	GetToken(claims map[string]interface{}) (string, error)
	VerifyToken(tokenString string) (token.TokenClaims, error)
}

type AuthService struct {
	authStorage    AuthServiceUserStorage
	sessionStorage AuthServiceSessionStorage
	tokenizer      Tokenizer
	hasher         *hash.Hasher
}

func NewAuthService(dbPath string, logger logger.Logger, secret string) (AuthService, error) {
	hasher := hash.NewHasher(secret)
	tokenizer := token.NewTokenizer(secret)

	authStorage, err := userStorage.InitUserStorage(dbPath, logger, hasher)

	if err != nil {
		return AuthService{}, err
	}

	sessionStorage, err := sessionStorage.NewSessionStorage()

	authService := AuthService{
		authStorage,
		sessionStorage,
		tokenizer,
		hasher,
	}

	return authService, err
}

type Token struct {
	AccessToken  string
	RefreshToken string
}

func (as *AuthService) Login(name string, password string) (Token, error) {
	hashedPassword, err := as.hasher.Hash(password)
	if err != nil {
		return Token{}, err
	}

	user, err := as.authStorage.GetUserByCredentials(name, hashedPassword)

	if err != nil {
		return Token{}, err
	}

	now := time.Now()

	claim := map[string]any{
		"sub":      user.ID,
		"username": user.Name,
		"exp":      now.Add(time.Hour * 72).Unix(),
		"created":  now.Unix(),
	}

	token, err := as.tokenizer.GetToken(claim)

	if err != nil {
		return Token{}, err
	}

	refreshClaim := map[string]any{
		"exp":     now.Add(time.Hour * 72).Unix(),
		"created": now.Unix(),
	}

	refreshToken, err := as.tokenizer.GetToken(refreshClaim)

	as.sessionStorage.CreateSession(user.ID, token, refreshToken)

	return Token{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, err
}

func (as *AuthService) CheckToken(token string) (bool, error) {
	claims, err := as.tokenizer.VerifyToken(token)

	if err != nil {
		return false, err
	}

	if claims["exp"].(float64) < float64(time.Now().Unix()) {
		return false, errors.New("token expired")
	}

	userID := claims["sub"].(string)

	_, err = as.sessionStorage.GetAccessToken(userID)

	if err != nil {
		return false, errors.New("session not found")
	}

	return true, nil
}
