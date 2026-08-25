// Package main является точкой входа для фонового воркера обработки видео.
package main

import (
	"log/slog"
	"os"

	"github.com/aimv/media-gallery/internal/infrastructure/config"
	"github.com/aimv/media-gallery/internal/infrastructure/persistence/postgres"
	"github.com/aimv/media-gallery/internal/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.InitLogger(cfg.LogLevel)
	slog.Info("Starting worker service...")

	if err := postgres.RunMigrations(cfg.DBDSN); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// TODO: дальнейшая инициализация и запуск воркера.
}
