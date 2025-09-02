/* Package sessionstorage */
package sessionstorage

// docker run -d -p 6379:6379 -e REDIS_PASSWORD=my_password redis

import (
	"context"
	"time"

	errfmt "auth/internal/lib/errors"

	"github.com/go-redis/redis/v8"
)

type SessionStorage struct {
	client *redis.Client
}

type SessionStorageOptions = redis.Options

const AccessTokenPrefix = "_access_token"
const RefreshTokenPrefix = "_refresh_token"

const SessionTTL = time.Hour * 72
const OperationTimeout = 5 * time.Second

func MustRun(options SessionStorageOptions) *SessionStorage {
	sessionStorage, err := New(options)

	if err != nil {
		panic(err)
	}

	return sessionStorage
}

func New(options SessionStorageOptions) (*SessionStorage, error) {
	const op = "sessionstorage.New"
	formatter := errfmt.NewOpErrorFormatter(op)
	client := redis.NewClient(&options)

	err := client.Ping(context.Background()).Err()

	if err != nil {
		return nil, formatter.Format(err)
	}

	return &SessionStorage{
		client: client,
	}, nil
}

func (s *SessionStorage) CreateSession(id string, accessToken string, refreshToken string) error {
	const op = "sessionstorage.CreateSession"
	formatter := errfmt.NewOpErrorFormatter(op)
	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
	defer cancel()

	err := s.client.Set(ctx, id+AccessTokenPrefix, accessToken, SessionTTL).Err()

	if err != nil {
		return formatter.Format(err)
	}
	err = s.client.Set(ctx, id+RefreshTokenPrefix, refreshToken, SessionTTL).Err()

	if err != nil {
		return formatter.Format(err)
	}

	return nil
}

func (s *SessionStorage) GetAccessToken(id string) (string, error) {
	const op = "sessionstorage.GetAccessToken"
	formatter := errfmt.NewOpErrorFormatter(op)
	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
	defer cancel()

	result, err := s.client.Get(ctx, id+AccessTokenPrefix).Result()
	if err != nil {
		return "", formatter.Format(err)
	}

	return result, nil
}

func (s *SessionStorage) GetRefreshToken(id string) (string, error) {
	const op = "sessionstorage.GetRefreshToken"
	formatter := errfmt.NewOpErrorFormatter(op)
	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
	defer cancel()

	result, err := s.client.Get(ctx, id+RefreshTokenPrefix).Result()
	if err != nil {
		return "", formatter.Format(err)
	}

	return result, nil
}

func (s *SessionStorage) GetSession(id string) (string, string, error) {
	const op = "sessionstorage.GetSession"
	formatter := errfmt.NewOpErrorFormatter(op)
	accessToken, err := s.GetAccessToken(id)

	if err != nil {
		return "", "", formatter.Format(err)
	}

	refreshToken, err := s.GetRefreshToken(id)

	if err != nil {
		return "", "", formatter.Format(err)
	}

	return accessToken, refreshToken, nil
}

func (s *SessionStorage) DeleteSession(id string) error {
	const op = "sessionstorage.DeleteSession"
	formatter := errfmt.NewOpErrorFormatter(op)
	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
	defer cancel()

	err := s.client.Del(ctx, id+AccessTokenPrefix, id+RefreshTokenPrefix).Err()
	if err != nil {
		return formatter.Format(err)
	}

	return nil
}
