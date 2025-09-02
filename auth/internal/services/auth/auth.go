package auth

import (
	"auth/internal/domain/models"
	errfmt "auth/internal/lib/errors"
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
	Stop() error
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
	Stop() error
}

type SessionStorage interface {
	CreateSession(id string, accessToken string, refreshToken string) error
	GetSession(id string) (string, string, error)
	GetAccessToken(id string) (string, error)
	GetRefreshToken(id string) (string, error)
	DeleteSession(id string) error
}

type Token struct {
	AccessToken  string
	RefreshToken string
}

func MustRun(
	userStorage UserStorage,
	sessionStorage SessionStorage,
	logger *logger.Logger,
	hasher *hash.Hasher,
	tokenizer *token.Tokenizer,
) *AuthService {
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
	ErrInvalidToken       = token.ErrInvalidToken
	ErrSessionNotFound    = errors.New("session not found")
	ErrTokenExpired       = errors.New("token expired")
)

const TokenTTL = time.Hour * 72

func (as *AuthService) Login(ctx context.Context, login string, password string) (*models.Token, error) {
	const op = "auth.Login"
	errFmt := errfmt.NewOpErrorFormatter(op)

	hashedPassword, err := as.hasher.Hash(password)
	if err != nil {
		return &models.Token{}, errFmt.Format(err)
	}

	user, err := as.userStorage.GetUserByCredentials(login, hashedPassword)
	if err != nil {
		return &models.Token{}, errFmt.Format(err)
	}

	tokens, err := as.GetToken(ctx, user.ID, map[string]any{
		"username": user.Name,
	})

	if err != nil {
		return &models.Token{}, errFmt.Format(err)
	}

	err = as.sessionStorage.CreateSession(user.ID, tokens.AccessToken, tokens.RefreshToken)
	if err != nil {
		return &models.Token{}, errFmt.Format(err)
	}

	return &models.Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, err
}

func (as *AuthService) Logout(ctx context.Context, token string) error {
	const op = "auth.Logout"
	errFmt := errfmt.NewOpErrorFormatter(op)

	claims, err := as.tokenizer.VerifyToken(token)
	if err != nil {
		return errFmt.Format(err)
	}

	userID := claims["sub"].(string)

	return as.sessionStorage.DeleteSession(userID)
}

func (as *AuthService) GetToken(ctx context.Context, sub string, restClaims map[string]any) (*models.Token, error) {
	const op = "auth.GetToken"
	errFmt := errfmt.NewOpErrorFormatter(op)

	now := time.Now()

	claim := map[string]any{
		"sub":     sub,
		"exp":     now.Add(TokenTTL).Unix(),
		"created": now,
	}

	maps.Copy(claim, restClaims)

	token, err := as.tokenizer.GetToken(claim)
	if err != nil {
		return &models.Token{}, errFmt.Format(err)
	}

	refreshClaim := map[string]any{
		"sub":     sub,
		"exp":     now.Add(TokenTTL).Unix(),
		"created": now,
	}

	refreshToken, err := as.tokenizer.GetToken(refreshClaim)
	if err != nil {
		return &models.Token{}, errFmt.Format(err)
	}

	return &models.Token{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}

func (as *AuthService) CheckToken(ctx context.Context, token string) (bool, error) {
	const op = "auth.CheckToken"
	errFmt := errfmt.NewOpErrorFormatter(op)

	claims, err := as.tokenizer.VerifyToken(token)
	if err != nil {
		return false, errFmt.Format(err)
	}

	userID := claims["sub"].(string)

	storedToken, err := as.sessionStorage.GetAccessToken(userID)

	if err != nil {
		return false, ErrSessionNotFound
	}

	if claims["exp"].(float64) < float64(time.Now().Unix()) || storedToken != token {
		return false, ErrTokenExpired
	}

	return true, nil
}

func (as *AuthService) RefreshToken(ctx context.Context, passedRefreshToken string) (*models.Token, error) {
	const op = "auth.RefreshToken"
	errFmt := errfmt.NewOpErrorFormatter(op)

	claims, err := as.tokenizer.VerifyToken(passedRefreshToken)
	if err != nil {
		return &models.Token{}, errFmt.Format(err)
	}

	userID := claims["sub"].(string)

	accessToken, refreshToken, err := as.sessionStorage.GetSession(userID)
	if err != nil {
		return &models.Token{}, errFmt.Format(err)
	}

	if passedRefreshToken != refreshToken {
		return &models.Token{}, errFmt.Format(ErrInvalidToken)
	}

	claims, err = as.tokenizer.VerifyToken(accessToken)
	if err != nil {
		return &models.Token{}, errFmt.Format(err)
	}

	tokens, err := as.GetToken(ctx, userID, map[string]any{
		"username": claims["username"].(string),
	})

	if err != nil {
		return &models.Token{}, errFmt.Format(err)
	}

	err = as.sessionStorage.CreateSession(userID, tokens.AccessToken, tokens.RefreshToken)
	if err != nil {
		return &models.Token{}, errFmt.Format(err)
	}

	return &models.Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (as *AuthService) Stop() error {
	return as.userStorage.Stop()
}
