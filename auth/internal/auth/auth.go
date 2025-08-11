package auth

import (
	"errors"
	"maps"
	"time"

	"cafe_main/internal/auth/hash"
	sessionStorage "cafe_main/internal/auth/session_storage"
	"cafe_main/internal/auth/token"
	userStorage "cafe_main/internal/auth/user_storage"
	"cafe_main/internal/logger"
)

type AuthServiceUserStorage interface {
	GetUserByCredentials(name string, password string) (*userStorage.User, error)
}

// TO DO Добавить сессии
type AuthServiceSessionStorage interface {
	CreateSession(id string, accessToken string, refreshToken string) error
	GetSession(id string) (string, string, error)
	GetAccessToken(id string) (string, error)
	GetRefreshToken(id string) (string, error)
	DeleteSession(id string) error
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

type SessionStorageOptions = sessionStorage.SessionStorageOptions
type UserStorageOptions = userStorage.UserStorageOptions

type DBOptions struct {
	UserStorage    UserStorageOptions
	SessionStorage SessionStorageOptions
}

func NewAuthService(dbOptions *DBOptions, logger logger.Logger, secret string) (AuthService, error) {
	hasher := hash.NewHasher(secret)
	tokenizer := token.NewTokenizer(secret)

	authStorage, err := userStorage.InitUserStorage(dbOptions.UserStorage, logger, hasher)

	if err != nil {
		return AuthService{}, err
	}

	sessionStorage, err := sessionStorage.NewSessionStorage(dbOptions.SessionStorage)

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

	tokens, err := as.GetToken(user.ID, map[string]any{
		"username": user.Name,
	})

	as.sessionStorage.CreateSession(user.ID, tokens.AccessToken, tokens.RefreshToken)

	return Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, err
}

func (as *AuthService) Logout(token string) error {
	claims, err := as.tokenizer.VerifyToken(token)

	if err != nil {
		return err
	}

	userID := claims["sub"].(string)

	return as.sessionStorage.DeleteSession(userID)
}

func (as *AuthService) GetToken(sub string, restClaims map[string]any) (Token, error) {
	now := time.Now()

	claim := map[string]any{
		"sub":     sub,
		"exp":     now.Add(time.Hour * 72).Unix(),
		"created": now,
	}

	maps.Copy(claim, restClaims)

	token, err := as.tokenizer.GetToken(claim)

	if err != nil {
		return Token{}, err
	}

	refreshClaim := map[string]any{
		"sub":     sub,
		"exp":     now.Add(time.Hour * 72).Unix(),
		"created": now,
	}

	refreshToken, err := as.tokenizer.GetToken(refreshClaim)

	if err != nil {
		return Token{}, err
	}

	return Token{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}

func (as *AuthService) CheckToken(token string) (bool, error) {
	claims, err := as.tokenizer.VerifyToken(token)

	if err != nil {
		return false, err
	}

	userID := claims["sub"].(string)

	storedToken, err := as.sessionStorage.GetAccessToken(userID)

	if err != nil {
		return false, errors.New("session not found")
	}

	if claims["exp"].(float64) < float64(time.Now().Unix()) || storedToken != token {
		return false, errors.New("token expired")
	}

	return true, nil
}

func (as *AuthService) RefreshToken(passedRefreshToken string) (Token, error) {
	claims, err := as.tokenizer.VerifyToken(passedRefreshToken)

	if err != nil {
		return Token{}, err
	}

	userID := claims["sub"].(string)

	accessToken, refreshToken, err := as.sessionStorage.GetSession(userID)

	if err != nil {
		return Token{}, err
	}

	if passedRefreshToken != refreshToken {
		return Token{}, errors.New("invalid refresh token")
	}

	claims, err = as.tokenizer.VerifyToken(accessToken)

	if err != nil {
		return Token{}, err
	}

	tokens, err := as.GetToken(userID, map[string]any{
		"username": claims["username"].(string),
	})

	as.sessionStorage.CreateSession(userID, tokens.AccessToken, tokens.RefreshToken)

	return Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, err
}
