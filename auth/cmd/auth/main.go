package main

import (
	"auth/internal/app"
	"auth/internal/config"
	"auth/internal/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	cfg := config.MustLoad()

	logger := logger.GetLogger()
	logger.Info("Starting auth service")

	application := app.New(logger, cfg.Grpc.Port)

	go application.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.GRPCServer.Stop()
}
