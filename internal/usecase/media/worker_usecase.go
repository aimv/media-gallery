package media

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aimv/media-gallery/internal/domain/media"
)

type WorkerUsecase struct {
	jobQueue       media.JobQueue
	repo           media.MediaRepository
	videoProcessor media.VideoProcessor
	leaseDuration  time.Duration
	logger         *slog.Logger
	workerID       string
}

func NewWorkerUsecase(
	jobQueue media.JobQueue,
	repo media.MediaRepository,
	videoProcessor media.VideoProcessor,
	leaseDuration time.Duration,
	logger *slog.Logger,
) *WorkerUsecase {
	host, _ := os.Hostname()
	pid := os.Getpid()
	return &WorkerUsecase{
		jobQueue:       jobQueue,
		repo:           repo,
		videoProcessor: videoProcessor,
		leaseDuration:  leaseDuration,
		logger:         logger,
		workerID:       host + "-" + strconv.Itoa(pid),
	}
}

// Run запускает бесконечный цикл обработки задач.
func (w *WorkerUsecase) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	w.logger.Info("worker loop started", "worker_id", w.workerID)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker loop stopped")
			return
		case <-ticker.C:
			w.processNext(ctx)
		}
	}
}

func (w *WorkerUsecase) processNext(ctx context.Context) {
	job, err := w.jobQueue.Claim(ctx, w.workerID, w.leaseDuration)
	if err != nil {
		w.logger.Error("claim job failed", "error", err)
		return
	}
	if job == nil {
		return
	}

	w.logger.Info("job claimed", "job_id", job.ID, "asset_id", job.AssetID)

	// Получаем ассет
	asset, err := w.repo.GetByID(ctx, job.AssetID)
	if err != nil {
		w.logger.Error("get asset failed", "job_id", job.ID, "error", err)
		_ = w.jobQueue.MarkFailed(ctx, job.ID, w.workerID, err.Error())
		return
	}

	// Запускаем heartbeat, который продлевает lease во время обработки
	heartbeatCtx, cancelHeartbeat := context.WithCancel(ctx)
	defer cancelHeartbeat()
	go w.heartbeat(heartbeatCtx, job.ID)

	// Обработка
	outputDir := filepath.Join("hls", asset.ID)
	if err := w.videoProcessor.ConvertToHLS(ctx, asset.StoragePath, outputDir); err != nil {
		w.logger.Error("hls conversion failed", "job_id", job.ID, "asset_id", asset.ID, "error", err)
		_ = w.jobQueue.MarkFailed(ctx, job.ID, w.workerID, err.Error())
		_ = w.repo.UpdateStatus(ctx, asset.ID, media.StatusFailed)
		return
	}

	// Обновляем hls_path и статус
	hlsPath := filepath.Join(outputDir, "master.m3u8")
	if err := w.repo.UpdateHlsPath(ctx, asset.ID, &hlsPath); err != nil {
		w.logger.Error("update hls path failed", "job_id", job.ID, "error", err)
	}
	if err := w.repo.UpdateStatus(ctx, asset.ID, media.StatusReady); err != nil {
		w.logger.Error("update status ready failed", "job_id", job.ID, "error", err)
	}
	if err := w.jobQueue.MarkDone(ctx, job.ID, w.workerID); err != nil {
		w.logger.Error("mark job done failed", "job_id", job.ID, "error", err)
	}

	w.logger.Info("job completed", "job_id", job.ID, "asset_id", asset.ID)
}

func (w *WorkerUsecase) heartbeat(ctx context.Context, jobID string) {
	ticker := time.NewTicker(w.leaseDuration / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.jobQueue.Heartbeat(ctx, jobID, w.workerID, w.leaseDuration); err != nil {
				w.logger.Warn("heartbeat failed", "job_id", jobID, "error", err)
			}
		}
	}
}
