package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aimv/media-gallery/internal/infrastructure/config"
	"github.com/aimv/media-gallery/internal/infrastructure/delivery/http/v1"
	"github.com/aimv/media-gallery/internal/infrastructure/persistence/postgres"
	"github.com/aimv/media-gallery/internal/infrastructure/storage"
	"github.com/aimv/media-gallery/internal/pkg/logger"
	"github.com/aimv/media-gallery/internal/usecase/media"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	// Инициализация логгера
	logger := logger.New(cfg.LogLevel)
	logger.Info("starting api service")

	// Запуск миграций
	migrationsDir := "internal/infrastructure/persistence/postgres/migrations"
	if err := postgres.RunMigrations(cfg.DBDSN, migrationsDir); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	// Подключение к БД
	pool, err := pgxpool.New(context.Background(), cfg.DBDSN)
	if err != nil {
		logger.Error("db pool init failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Инициализация инфраструктуры
	mediaRepo := postgres.NewMediaRepository(pool)
	localStorage := storage.NewLocalStorage(cfg.StorageRoot)

	// Слой usecase
	mediaUsecase := media.NewMediaUsecase(localStorage, mediaRepo)

	// HTTP слой
	mediaHandler := v1.NewMediaHandler(mediaUsecase)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/media/upload", mediaHandler.Upload)

	// Healthcheck
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info("http server listening", "port", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", "error", err)
	}
	logger.Info("server stopped")
}
