// Package main является точкой входа для фонового воркера обработки видео.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/aimv/media-gallery/internal/infrastructure/config"
	"github.com/aimv/media-gallery/internal/infrastructure/persistence/postgres"
	"github.com/aimv/media-gallery/internal/pkg/logger"
)

const (
	leaseDuration   = 5 * time.Minute
	pollInterval    = 2 * time.Second
	mockProcessTime = 500 * time.Millisecond
	migrationsPath  = "file://internal/infrastructure/persistence/postgres/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.InitLogger(cfg.LogLevel)
	slog.Info("Starting worker service...", "pool_size", cfg.WorkerPoolSize)

	if err := postgres.RunMigrations(cfg.DBDSN, migrationsPath); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	pool, err := postgres.NewPool(cfg.DBDSN)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queueRepo := postgres.NewQueueRepository(pool)

	// Корневой контекст с отменой по сигналу.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		slog.Info("Received shutdown signal", "signal", sig.String())
		cancel()
	}()

	var wg sync.WaitGroup
	for i := 1; i <= cfg.WorkerPoolSize; i++ {
		workerID := fmt.Sprintf("worker-%d", i)
		wg.Add(1)
		go runWorker(ctx, &wg, workerID, queueRepo)
	}

	wg.Wait()
	slog.Info("Worker service stopped gracefully")
}

// runWorker запускает цикл обработки задач для одного воркера.
func runWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	workerID string,
	queueRepo *postgres.QueueRepository,
) {
	defer wg.Done()
	slog.Info("Worker started", "worker_id", workerID)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Worker stopping", "worker_id", workerID)
			return
		default:
		}

		job, err := queueRepo.ClaimNext(ctx, workerID, leaseDuration)
		if err != nil {
			slog.Error("Failed to claim job", "worker_id", workerID, "error", err)
			time.Sleep(pollInterval)
			continue
		}

		if job == nil {
			// Пустая очередь — ждём перед следующей попыткой.
			time.Sleep(pollInterval)
			continue
		}

		slog.Info("Job claimed successfully",
			"worker_id", workerID,
			"job_id", job.ID,
			"asset_id", job.AssetID,
		)

		// TODO(T4.6): здесь будет вызов usecase обработки видео.
		select {
		case <-ctx.Done():
			slog.Info("Shutdown during processing", "worker_id", workerID, "job_id", job.ID)
			return
		case <-time.After(mockProcessTime):
		}
	}
}
