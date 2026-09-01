// Package media содержит use-case'ы для работы с медиафайлами.
package media

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/aimv/media-gallery/internal/domain/repository"
	"github.com/aimv/media-gallery/internal/domain/service"
)

// ProcessVideoUseCase реализует сценарий асинхронной обработки видео в HLS.
type ProcessVideoUseCase struct {
	mediaRepo         repository.MediaRepository
	queueRepo         repository.QueueRepository
	videoProcessor    service.VideoProcessor
	storage           repository.FileStorage
	storageBaseDir    string
	heartbeatInterval time.Duration
	leaseDuration     time.Duration
}

// NewProcessVideoUseCase создаёт use-case с указанными зависимостями.
// storageBaseDir используется для формирования физических путей к файлам,
// с которыми работают ffmpeg/ffprobe.
func NewProcessVideoUseCase(
	mediaRepo repository.MediaRepository,
	queueRepo repository.QueueRepository,
	videoProcessor service.VideoProcessor,
	storage repository.FileStorage,
	storageBaseDir string,
) *ProcessVideoUseCase {
	return &ProcessVideoUseCase{
		mediaRepo:         mediaRepo,
		queueRepo:         queueRepo,
		videoProcessor:    videoProcessor,
		storage:           storage,
		storageBaseDir:    storageBaseDir,
		heartbeatInterval: 30 * time.Second,
		leaseDuration:     5 * time.Minute,
	}
}

// Execute выполняет полный цикл обработки задачи.
func (u *ProcessVideoUseCase) Execute(ctx context.Context, job *entity.ProcessingJob) error {
	asset, err := u.mediaRepo.FindByID(ctx, job.AssetID)
	if err != nil {
		return fmt.Errorf("find asset: %w", err)
	}

	if err := u.mediaRepo.UpdateStatus(ctx, asset.ID, entity.StatusProcessing); err != nil {
		return fmt.Errorf("set asset status to processing: %w", err)
	}

	if err := u.queueRepo.UpdateStatus(ctx, job.ID, entity.JobStatusProcessing, nil); err != nil {
		return fmt.Errorf("set job status to processing: %w", err)
	}

	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(u.heartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := u.queueRepo.ExtendLease(ctx, job.ID, u.leaseDuration); err != nil {
					slog.Error("Failed to extend lease",
						"job_id", job.ID,
						"worker_id", job.LeaseOwner,
						"error", err,
					)
				}
			}
		}
	}()

	processErr := u.process(ctx, asset, job)

	close(stopCh)
	wg.Wait()

	if processErr != nil {
		return u.fail(ctx, asset, job, processErr)
	}
	return nil
}

// process выполняет основную последовательность действий по обработке видео.
func (u *ProcessVideoUseCase) process(ctx context.Context, asset *entity.MediaAsset, job *entity.ProcessingJob) error {
	inputPath := filepath.Join(u.storageBaseDir, asset.StoragePath)

	// Временная директория: сюда ffmpeg пишет плейлист и сегменты.
	outputDir := filepath.Join(u.storageBaseDir, "tmp", "hls", job.AssetID.String())

	// Финальная директория: сюда перемещается готовый HLS-поток.
	finalDir := filepath.Join(u.storageBaseDir, "hls", job.AssetID.String())

	if err := u.videoProcessor.ProcessToHLS(ctx, inputPath, outputDir); err != nil {
		return fmt.Errorf("process to hls: %w", err)
	}

	meta, err := u.videoProcessor.ProbeMetadata(ctx, inputPath)
	if err != nil {
		return fmt.Errorf("probe metadata: %w", err)
	}

	// Атомарно переносим полностью готовый HLS-поток в публичную директорию.
	if err := u.storage.MoveDir(ctx, outputDir, finalDir); err != nil {
		return fmt.Errorf("move hls dir: %w", err)
	}

	if err := u.mediaRepo.UpdateMetadata(ctx, asset.ID, meta.Width, meta.Height, meta.DurationMS, meta.Codec); err != nil {
		return fmt.Errorf("update metadata: %w", err)
	}

	if err := u.mediaRepo.UpdateStatus(ctx, asset.ID, entity.StatusReady); err != nil {
		return fmt.Errorf("set asset status to ready: %w", err)
	}

	if err := u.queueRepo.UpdateStatus(ctx, job.ID, entity.JobStatusSuccess, nil); err != nil {
		return fmt.Errorf("set job status to success: %w", err)
	}

	return nil
}

// fail выполняет компенсацию статусов и очистку временных файлов при ошибке.
func (u *ProcessVideoUseCase) fail(ctx context.Context, asset *entity.MediaAsset, job *entity.ProcessingJob, cause error) error {
	// Очищаем временную директорию, если она осталась после сбоя.
	tempDir := filepath.Join(u.storageBaseDir, "tmp", "hls", job.AssetID.String())
	if err := os.RemoveAll(tempDir); err != nil {
		slog.Error("Failed to cleanup temp hls dir",
			"path", tempDir,
			"error", err,
		)
	}

	errMsg := cause.Error()

	if err := u.mediaRepo.UpdateStatus(ctx, asset.ID, entity.StatusFailed); err != nil {
		slog.Error("Failed to compensate asset status",
			"asset_id", asset.ID,
			"error", err,
		)
	}

	if err := u.queueRepo.UpdateStatus(ctx, job.ID, entity.JobStatusFailed, &errMsg); err != nil {
		slog.Error("Failed to compensate job status",
			"job_id", job.ID,
			"error", err,
		)
	}

	return cause
}
