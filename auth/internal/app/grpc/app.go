package grpcapp

import (
	authgrpc "auth/internal/grpc/auth"
	"auth/internal/lib/logger"
	"context"
	"fmt"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	logger     *logger.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(logger *logger.Logger, port int, authService authgrpc.Auth) *App {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			logger.Error("Recovered from panic", p)

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(InterceptorLogger(logger), loggingOpts...),
	))

	authgrpc.Register(gRPCServer, authService)

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

func InterceptorLogger(l *logger.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		fmt.Println(lvl, msg)
		logLevel := logrus.InfoLevel
		if lvl == logging.LevelError {
			logLevel = logrus.ErrorLevel
		}
		l.Log(logLevel, msg)
	})
}
