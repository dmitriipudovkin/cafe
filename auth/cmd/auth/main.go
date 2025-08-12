package main

import (
	"cafe_main/internal/config"
	"fmt"
)

func main() {
	// TODO: инит конфига
	cfg := config.MustLoad()

	fmt.Println(cfg)

	// TODO: инит логгера

	// TODO: инит приложения (app)

	// TODO: запуск grpc
}
