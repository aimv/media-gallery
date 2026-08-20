package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aimv/media-gallery/internal/infrastructure/config"
	"github.com/aimv/media-gallery/internal/infrastructure/persistence/postgres"
	"github.com/aimv/media-gallery/internal/infrastructure/worker"
	"github.com/aimv/media-gallery/internal/pkg/logger"
	"github.com/aimv/media-gallery/internal/usecase/media"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}
	logger := logger.New(cfg.LogLevel)
	logger.Info("starting worker service")

	// Подключение к БД
	pool, err := pgxpool.New(context.Background(), cfg.DBDSN)
	if err != nil {
		logger.Error("db pool init failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Инициализация инфраструктуры
	mediaRepo := postgres.NewMediaRepository(pool)
	jobQueue := postgres.NewJobQueue(pool)
	videoProcessor := worker.NewFFmpegProcessor(cfg.FFmpegPath, cfg.FFprobePath, cfg.StorageRoot)

	// Слой usecase
	workerUsecase := media.NewWorkerUsecase(jobQueue, mediaRepo, videoProcessor, cfg.LeaseDuration, logger)

	// Контекст с graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("worker started")
	workerUsecase.Run(ctx)
	logger.Info("worker stopped")
}
