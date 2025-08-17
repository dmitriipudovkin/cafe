package app

import (
	grpcapp "auth/internal/app/grpc"
	"auth/internal/lib/logger"
	"auth/internal/services/auth"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	logger *logger.Logger,
	grpcPort int,
) *App {
	// TODO: инициализировать хранилище

	// TODO: инициализировать auth сервис
	authService := auth.New()

	grpcapp := grpcapp.New(logger, grpcPort, authService)

	return &App{
		GRPCServer: grpcapp,
	}
}
