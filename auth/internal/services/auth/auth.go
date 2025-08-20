package auth

import (
	"auth/internal/domain/models"
	"auth/internal/grpc/auth"
	"auth/internal/lib/hash"
	"auth/internal/lib/logger"
	"auth/internal/lib/token"
	userStorage "auth/internal/storage/user_storage"
	"context"
	"errors"
	"maps"
	"time"
)

type AuthServiceUserStorage interface {
	GetUserByCredentials(name string, password string) (*models.User, error)
}

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
	userStorage    AuthServiceUserStorage
	sessionStorage AuthServiceSessionStorage
	tokenizer      Tokenizer
	hasher         *hash.Hasher
}

type UserStorage interface {
	GetUserByCredentials(name string, password string) (*models.User, error)
}

type SessionStorage interface {
	CreateSession(id string, accessToken string, refreshToken string) error
	GetSession(id string) (string, string, error)
	GetAccessToken(id string) (string, error)
	GetRefreshToken(id string) (string, error)
	DeleteSession(id string) error
}

func MustRun(userStorage UserStorage, sessionStorage SessionStorage, logger *logger.Logger, hasher *hash.Hasher, tokenizer *token.Tokenizer) *AuthService {
	authService := &AuthService{
		userStorage,
		sessionStorage,
		tokenizer,
		hasher,
	}

	return authService
}

var (
	ErrInvalidCredentials = userStorage.ErrInvalidCredentials
)

func (as *AuthService) Login(ctx context.Context, login string, password string) (auth.Token, error) {
	hashedPassword, err := as.hasher.Hash(password)

	if err != nil {
		return auth.Token{}, err
	}

	user, err := as.userStorage.GetUserByCredentials(login, hashedPassword)

	if err != nil {
		return auth.Token{}, err
	}

	tokens, err := as.GetToken(ctx, user.ID, map[string]any{
		"username": user.Name,
	})

	as.sessionStorage.CreateSession(user.ID, tokens.AccessToken, tokens.RefreshToken)

	return auth.Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, err
}

func (as *AuthService) Logout(ctx context.Context, token string) error {
	claims, err := as.tokenizer.VerifyToken(token)

	if err != nil {
		return err
	}

	userID := claims["sub"].(string)

	return as.sessionStorage.DeleteSession(userID)
}

func (as *AuthService) GetToken(ctx context.Context, sub string, restClaims map[string]any) (auth.Token, error) {
	now := time.Now()

	claim := map[string]any{
		"sub":     sub,
		"exp":     now.Add(time.Hour * 72).Unix(),
		"created": now,
	}

	maps.Copy(claim, restClaims)

	token, err := as.tokenizer.GetToken(claim)

	if err != nil {
		return auth.Token{}, err
	}

	refreshClaim := map[string]any{
		"sub":     sub,
		"exp":     now.Add(time.Hour * 72).Unix(),
		"created": now,
	}

	refreshToken, err := as.tokenizer.GetToken(refreshClaim)

	if err != nil {
		return auth.Token{}, err
	}

	return auth.Token{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}

func (as *AuthService) CheckToken(ctx context.Context, token string) (bool, error) {
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

func (as *AuthService) RefreshToken(ctx context.Context, passedRefreshToken string) (auth.Token, error) {
	claims, err := as.tokenizer.VerifyToken(passedRefreshToken)

	if err != nil {
		return auth.Token{}, err
	}

	userID := claims["sub"].(string)

	accessToken, refreshToken, err := as.sessionStorage.GetSession(userID)

	if err != nil {
		return auth.Token{}, err
	}

	if passedRefreshToken != refreshToken {
		return auth.Token{}, errors.New("invalid refresh token")
	}

	claims, err = as.tokenizer.VerifyToken(accessToken)

	if err != nil {
		return auth.Token{}, err
	}

	tokens, err := as.GetToken(ctx, userID, map[string]any{
		"username": claims["username"].(string),
	})

	as.sessionStorage.CreateSession(userID, tokens.AccessToken, tokens.RefreshToken)

	return auth.Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, err
}
