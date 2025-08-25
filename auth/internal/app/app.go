package app

import (
	grpcapp "auth/internal/app/grpc"
	"auth/internal/config"
	"auth/internal/lib/hash"
	"auth/internal/lib/logger"
	"auth/internal/lib/token"
	"auth/internal/services/auth"
	sessionStorage "auth/internal/storage/session_storage"
	userStorage "auth/internal/storage/user_storage"
)

type App struct {
	GRPCServer  *grpcapp.App
	authService *auth.AuthService
}

func New(
	logger *logger.Logger,
	config *config.Config,
) *App {
	hasher := hash.NewHasher(config.SecretKey)
	tokenizer := token.NewTokenizer(config.SecretKey)

	userStorage := userStorage.MustRun(userStorage.UserStorageOptions{
		DBPath:        config.UserStorage.DBPath,
		AdminLogin:    config.UserStorage.AdminLogin,
		AdminPassword: config.UserStorage.AdminPassword,
	}, logger, hasher)

	sessionStorage := sessionStorage.MustRun(sessionStorage.SessionStorageOptions{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	authService := auth.MustRun(userStorage, sessionStorage, logger, hasher, tokenizer)

	grpcapp := grpcapp.New(logger, config.Grpc.Port, authService)

	return &App{
		GRPCServer:  grpcapp,
		authService: authService,
	}
}

func (app *App) Run() {
	app.GRPCServer.MustRun()
}

func (app *App) Stop() error {
	app.GRPCServer.Stop()
	err := app.authService.Stop()

	if err != nil {
		return err
	}

	return nil
}
