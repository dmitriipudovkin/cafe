/* Package sessionstorage */
package sessionstorage

// docker run -d -p 6379:6379 -e REDIS_PASSWORD=my_password redis

import (
	"context"
	"time"

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
	client := redis.NewClient(&options)

	err := client.Ping(context.Background()).Err()

	return &SessionStorage{
		client: client,
	}, err
}

func (s *SessionStorage) CreateSession(id string, accessToken string, refreshToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
	defer cancel()

	err := s.client.Set(ctx, id+AccessTokenPrefix, accessToken, SessionTTL).Err()

	if err != nil {
		return err
	}
	err = s.client.Set(ctx, id+RefreshTokenPrefix, refreshToken, SessionTTL).Err()

	return err
}

func (s *SessionStorage) GetAccessToken(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
	defer cancel()

	return s.client.Get(ctx, id+AccessTokenPrefix).Result()
}

func (s *SessionStorage) GetRefreshToken(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
	defer cancel()

	return s.client.Get(ctx, id+RefreshTokenPrefix).Result()
}

func (s *SessionStorage) GetSession(id string) (string, string, error) {
	accessToken, err := s.GetAccessToken(id)

	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.GetRefreshToken(id)

	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *SessionStorage) DeleteSession(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
	defer cancel()

	return s.client.Del(ctx, id+AccessTokenPrefix, id+RefreshTokenPrefix).Err()
}
