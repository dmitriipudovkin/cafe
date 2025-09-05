package main

import (
	"auth/internal/app"
	"auth/internal/config"
	"auth/internal/lib/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil && os.Getenv("CONFIG_PATH") == "" {
		panic("failed to load env")
	}

	cfg := config.MustLoad()

	logger := logger.NewLogger()
	logger.Info("Starting auth service")

	application := app.New(logger, cfg)

	go application.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	err = application.Stop()

	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}
	logger.Info("Stopping auth service")
}
