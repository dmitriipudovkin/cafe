package app

import (
	grpcapp "auth/internal/app/grpc"
	"auth/internal/logger"
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

	grpcapp := grpcapp.New(logger, grpcPort)

	return &App{
		GRPCServer: grpcapp,
	}
}
