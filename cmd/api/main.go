package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aimv/media-gallery/internal/infrastructure/config"
	v1 "github.com/aimv/media-gallery/internal/infrastructure/delivery/http/v1"
	"github.com/aimv/media-gallery/internal/infrastructure/persistence/postgres"
	"github.com/aimv/media-gallery/internal/infrastructure/storage"
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
	logger.Info("starting api service")

	// Миграции
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
	//jobQueue := postgres.NewJobQueue(pool)
	cmsRepo := postgres.NewCMSRepository(pool)
	localStorage := storage.NewLocalStorage(cfg.StorageRoot)

	// Usecases
	mediaUsecase := media.NewMediaUsecase(localStorage, mediaRepo)

	// Handlers
	mediaHandler := v1.NewMediaHandler(mediaUsecase, cfg.MaxUploadSize)
	cmsHandler := v1.NewCMSHandler(cmsRepo)

	// Роутер (Go 1.22+ паттерны)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("POST /api/media/upload", mediaHandler.Upload)
	mux.HandleFunc("POST /api/cms/blocks", cmsHandler.SaveBlock)
	mux.HandleFunc("GET /api/cms/blocks/{id}", cmsHandler.GetBlock)

	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		logger.Info("http server listening", "port", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
