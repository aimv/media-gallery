package main

import (
	"fmt"
	"os"

	"github.com/aimv/media-gallery/internal/infrastructure/config"
	"github.com/aimv/media-gallery/internal/pkg/logger"
)

func main() {
	// 1. Инициализируем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	// 2. Инициализируем структурированный логгер
	log := logger.New(cfg.LogLevel)

	log.Info("Сервер успешно инициализирован",
		"http_port", cfg.HTTPPort,
		"log_level", cfg.LogLevel,
		"storage_root", cfg.StorageRoot,
	)
}
