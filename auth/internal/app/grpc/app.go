package grpcapp

import (
	authgrpc "auth/internal/grpc/auth"
	"auth/internal/logger"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type App struct {
	logger     *logger.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(logger *logger.Logger, port int) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer)

	return &App{
		logger:     logger,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (g *App) MustRun() {
	if err := g.Run(); err != nil {
		panic(err)
	}
}

func (g *App) Run() error {
	const op = "grpcapp.App.Run"
	logger := g.logger.WithField("op", op)

	logger.Info("Start gRPC server on port", g.port)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := g.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (g *App) Stop() {
	const op = "grpcapp.App.Stop"
	logger := g.logger.WithField("op", op)

	logger.Info("Stop gRPC server")

	g.gRPCServer.GracefulStop()
}
