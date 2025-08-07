package auth

import (
	"cafe_main/internal/logger"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v ./... -run TestAuthService

const testDBPath string = "./auth_test.db"

var BDOptions = DBOptions{
	UserStorage: UserStorageOptions{
		DBPath:        testDBPath,
		AdminLogin:    os.Getenv("ADMIN_LOGIN"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
	},
	SessionStorage: SessionStorageOptions{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	},
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
		Error    string
	}{
		{"successful login", BDOptions.UserStorage.AdminLogin, BDOptions.UserStorage.AdminPassword, false, ""},
		{"failed login due to invalid credentials", BDOptions.UserStorage.AdminLogin, BDOptions.UserStorage.AdminPassword + "123", true, "invalid credentials"},
	}

	logger := logger.GetLogger()

	authService, err := NewAuthService(&BDOptions, logger, os.Getenv("SECRET_KEY"))
	assert.NoError(t, err)

	defer os.Remove(testDBPath)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := authService.Login(tt.username, tt.password)
			if err != nil {
				if tt.wantErr && tt.Error != err.Error() {
					t.Errorf("Login() error = %v, wantErr %v, error = %v", err, tt.wantErr, tt.Error)
					return
				}
			}

			if token.AccessToken != "" {
				_, err := authService.CheckToken(token.AccessToken)
				assert.NoError(t, err)
			}
		})
	}

	// clear
}

func TestAuthService_RefreshToken(t *testing.T) {
	logger := logger.GetLogger()

	authService, err := NewAuthService(&BDOptions, logger, os.Getenv("SECRET_KEY"))
	assert.NoError(t, err)

	defer os.Remove(testDBPath)

	t.Run("successful refresh token", func(t *testing.T) {
		prevToken, err := authService.Login(BDOptions.UserStorage.AdminLogin, BDOptions.UserStorage.AdminPassword)
		assert.NoError(t, err)

		// check token is valid
		_, err = authService.CheckToken(prevToken.AccessToken)
		assert.NoError(t, err)

		// refresh token
		newToken, err := authService.RefreshToken(prevToken.RefreshToken)
		assert.NoError(t, err)

		// check if old token isn't valid
		_, err = authService.CheckToken(prevToken.AccessToken)
		assert.Error(t, err)

		// check if new token is valid
		_, err = authService.CheckToken(newToken.AccessToken)
		assert.NoError(t, err)

		// logout
		err = authService.Logout(prevToken.AccessToken)
		assert.NoError(t, err)

		// old token isn't valid after logout
		_, err = authService.CheckToken(prevToken.AccessToken)
		assert.Error(t, err)
	})
}
