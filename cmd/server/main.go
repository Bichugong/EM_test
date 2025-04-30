package main

import (
	"em_test/internal/config"
	"em_test/internal/transport/rest"
	"em_test/pkg/logging"
	"em_test/pkg/postgres"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Инициализация логгера
	logger := logging.New(cfg.LogLevel)

	// Подключение к PostgreSQL
	db, err := postgres.New(cfg.DB)
	if err != nil {
		logger.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	// Запуск HTTP-сервера
	server := rest.NewServer(cfg, db, logger)
	if err := server.Run(); err != nil {
		logger.Fatal("Server failed", err)
	}
}