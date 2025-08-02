package sessionStorage

// docker run -d -p 6379:6379 -e REDIS_PASSWORD=my_password redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type SessionStorage struct {
	client *redis.Client
}

func NewSessionStorage() (*SessionStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis address
		Password: "",               // Redis password
		DB:       0,                // Redis database
	})

	err := client.Ping(context.Background()).Err()

	return &SessionStorage{
		client: client,
	}, err
}

func (s *SessionStorage) CreateSession(id string, accessToken string, refreshToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.client.Set(ctx, id+"_access_token", accessToken, time.Hour*72).Err()

	if err != nil {
		return err
	}
	err = s.client.Set(ctx, id+"_refresh_token", refreshToken, time.Hour*72).Err()

	return err
}

func (s *SessionStorage) GetAccessToken(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.client.Get(ctx, id+"_access_token").Result()
}

func (s *SessionStorage) GetRefreshToken(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.client.Get(ctx, id+"_refresh_token").Result()
}
