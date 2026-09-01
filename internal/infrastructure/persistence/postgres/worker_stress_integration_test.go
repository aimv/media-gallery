package postgres

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/aimv/media-gallery/internal/domain/service"
	"github.com/aimv/media-gallery/internal/infrastructure/config"
	"github.com/aimv/media-gallery/internal/usecase/media"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// --- Stubs for E2E test ---

// stressStubVideoProcessor — заглушка videoProcessor, возвращающая фиксированные метаданные.
type stressStubVideoProcessor struct{}

func (s *stressStubVideoProcessor) Validate(_ context.Context, _ string) (*service.VideoMetadata, error) {
	return &service.VideoMetadata{Width: 1920, Height: 1080, DurationMS: 10000, Codec: "h264"}, nil
}

func (s *stressStubVideoProcessor) ProcessToHLS(_ context.Context, _, _ string) error {
	return nil
}

func (s *stressStubVideoProcessor) ProbeMetadata(_ context.Context, _ string) (*service.VideoMetadata, error) {
	return &service.VideoMetadata{Width: 1920, Height: 1080, DurationMS: 10000, Codec: "h264"}, nil
}

// stressStubFileStorage — заглушка FileStorage.
type stressStubFileStorage struct{}

func (s *stressStubFileStorage) Save(_ context.Context, _ string, _ io.Reader) (string, error) {
	return "", nil
}

func (s *stressStubFileStorage) Delete(_ context.Context, _ string) error { return nil }

func (s *stressStubFileStorage) MoveDir(_ context.Context, _, _ string) error { return nil }

// --- Helpers ---

// setupStressTestPool подготавливает подключение к изолированной тестовой БД,
// применяет миграции и очищает рабочие таблицы.
func setupStressTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	root, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to locate project root: %v", err)
	}
	_ = godotenv.Load(filepath.Join(root, ".env"))

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	pool, err := NewPool(cfg.TestDBDSN)
	if err != nil {
		t.Skipf("test database not available, skipping integration test: %v", err)
	}

	migrationsPath := "file://" + filepath.Join(root, "internal/infrastructure/persistence/postgres/migrations")
	if err := RunMigrations(cfg.TestDBDSN, migrationsPath); err != nil {
		pool.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	ctx := context.Background()
	if _, err := pool.Exec(ctx, `DELETE FROM processing_jobs`); err != nil {
		pool.Close()
		t.Fatalf("failed to clean processing_jobs: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM media_assets`); err != nil {
		pool.Close()
		t.Fatalf("failed to clean media_assets: %v", err)
	}

	return pool
}

// insertQueuedJob вставляет медиаассет и связанную с ним задачу в статусе queued.
func insertQueuedJob(ctx context.Context, t *testing.T, pool *pgxpool.Pool, suffix string) (uuid.UUID, uuid.UUID) {
	t.Helper()

	assetID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO media_assets (
			id, original_filename, media_type, status, storage_path,
			size_bytes, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, now(), now())
	`, assetID, "stress-"+suffix+".mp4", entity.MediaTypeMP4, entity.StatusUploaded,
		"uploads/stress-"+suffix+".mp4", 1024, `{}`)
	if err != nil {
		t.Fatalf("failed to insert media asset: %v", err)
	}

	jobID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO processing_jobs (
			id, asset_id, status, attempt, max_attempts, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, now(), now())
	`, jobID, assetID, entity.JobStatusQueued, 0, 3)
	if err != nil {
		t.Fatalf("failed to insert processing job: %v", err)
	}

	return assetID, jobID
}

// --- Tests ---

// TestIntegration_Queue_RaceConditions проверяет, что конкурентный захват задач
// через FOR UPDATE SKIP LOCKED не выдаёт одну и ту же задачу двум воркерам.
func TestIntegration_Queue_RaceConditions(t *testing.T) {
	pool := setupStressTestPool(t)
	defer pool.Close()

	ctx := context.Background()

	const (
		totalJobs   = 10
		totalWorker = 5
	)

	for i := range totalJobs {
		insertQueuedJob(ctx, t, pool, fmt.Sprintf("%d", i))
	}

	queueRepo := NewQueueRepository(pool)

	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		seen  = make(map[uuid.UUID]struct{}, totalJobs)
		total int
	)

	for i := range totalWorker {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			workerID := fmt.Sprintf("stress-worker-%d", idx)

			for {
				select {
				case <-workerCtx.Done():
					return
				default:
				}

				job, err := queueRepo.ClaimNext(workerCtx, workerID, 5*time.Minute)
				if err != nil {
					t.Errorf("worker %s: claim error: %v", workerID, err)
					return
				}
				if job == nil {
					// Больше нет доступных задач — воркер завершает работу.
					return
				}

				mu.Lock()
				if _, dup := seen[job.ID]; dup {
					mu.Unlock()
					t.Errorf("duplicate job claimed: %s", job.ID)
					return
				}
				seen[job.ID] = struct{}{}
				total++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if total != totalJobs {
		t.Errorf("total claimed jobs = %d, want %d", total, totalJobs)
	}

	// Дополнительная проверка: все задачи в БД перешли в processing.
	var processingCount int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM processing_jobs WHERE status = $1`,
		entity.JobStatusProcessing,
	).Scan(&processingCount); err != nil {
		t.Fatalf("failed to count processing jobs: %v", err)
	}
	if processingCount != totalJobs {
		t.Errorf("processing jobs in DB = %d, want %d", processingCount, totalJobs)
	}
}

// TestIntegration_Worker_EndToEnd проверяет сквозной цикл обработки задачи:
// claim → execute use case → статусы в БД обновляются до success/ready.
func TestIntegration_Worker_EndToEnd(t *testing.T) {
	pool := setupStressTestPool(t)
	defer pool.Close()

	ctx := context.Background()

	assetID, jobID := insertQueuedJob(ctx, t, pool, "e2e")

	mediaRepo := NewMediaRepository(pool)
	queueRepo := NewQueueRepository(pool)

	// Захват задачи воркером.
	job, err := queueRepo.ClaimNext(ctx, "e2e-worker", 5*time.Minute)
	if err != nil {
		t.Fatalf("ClaimNext() unexpected error: %v", err)
	}
	if job == nil {
		t.Fatal("expected to claim a job, got nil")
	}
	if job.ID != jobID {
		t.Fatalf("claimed job ID = %s, want %s", job.ID, jobID)
	}

	processor := &stressStubVideoProcessor{}
	storage := &stressStubFileStorage{}

	uc := media.NewProcessVideoUseCase(mediaRepo, queueRepo, processor, storage, t.TempDir())

	if err := uc.Execute(ctx, job); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	// Проверяем статус задачи.
	var jobStatus string
	if err := pool.QueryRow(ctx,
		`SELECT status FROM processing_jobs WHERE id = $1`, jobID,
	).Scan(&jobStatus); err != nil {
		t.Fatalf("query job status: %v", err)
	}
	if jobStatus != string(entity.JobStatusSuccess) {
		t.Errorf("job status = %q, want %q", jobStatus, entity.JobStatusSuccess)
	}

	// Проверяем статус ассета и обновлённые метаданные.
	var (
		assetStatus string
		width       int
		height      int
		durationMS  int64
		codec       string
	)
	if err := pool.QueryRow(ctx, `
		SELECT status, width, height, duration_ms, codec
		FROM media_assets WHERE id = $1
	`, assetID).Scan(&assetStatus, &width, &height, &durationMS, &codec); err != nil {
		t.Fatalf("query asset status: %v", err)
	}
	if assetStatus != string(entity.StatusReady) {
		t.Errorf("asset status = %q, want %q", assetStatus, entity.StatusReady)
	}
	if width != 1920 || height != 1080 {
		t.Errorf("asset dimensions = %dx%d, want 1920x1080", width, height)
	}
	if durationMS != 10000 {
		t.Errorf("asset duration_ms = %d, want 10000", durationMS)
	}
	if codec != "h264" {
		t.Errorf("asset codec = %q, want h264", codec)
	}
}
