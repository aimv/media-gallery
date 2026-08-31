package postgres

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/aimv/media-gallery/internal/infrastructure/config"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func TestIntegration_QueueRepository_Lifecycle(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to locate project root: %v", err)
	}
	_ = godotenv.Load(filepath.Join(root, ".env"))

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	pool, err := NewPool(cfg.DBDSN)
	if err != nil {
		t.Skipf("database not available, skipping integration test: %v", err)
	}
	defer pool.Close()

	// Применяем миграции для гарантии наличия схемы.
	migrationsPath := "file://" + filepath.Join(root, "internal/infrastructure/persistence/postgres/migrations")
	if err := RunMigrations(cfg.DBDSN, migrationsPath); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewQueueRepository(pool)
	ctx := context.Background()

	// Очищаем таблицы перед тестом, чтобы предыдущие неудачные запуски не мешали.
	_, err = pool.Exec(ctx, `DELETE FROM processing_jobs`)
	if err != nil {
		t.Fatalf("failed to clean processing_jobs: %v", err)
	}
	_, err = pool.Exec(ctx, `DELETE FROM media_assets`)
	if err != nil {
		t.Fatalf("failed to clean media_assets: %v", err)
	}

	// Списки идентификаторов для очистки после теста.
	var cleanupAssetIDs []uuid.UUID
	var cleanupJobIDs []uuid.UUID

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		for _, jobID := range cleanupJobIDs {
			_, _ = pool.Exec(cleanupCtx, `DELETE FROM processing_jobs WHERE id = $1`, jobID)
		}
		for _, assetID := range cleanupAssetIDs {
			_, _ = pool.Exec(cleanupCtx, `DELETE FROM media_assets WHERE id = $1`, assetID)
		}
	})

	// Сценарий 1: Пустая очередь.
	t.Run("empty queue returns nil", func(t *testing.T) {
		job, err := repo.ClaimNext(ctx, "worker-1", 5*time.Minute)
		if err != nil {
			t.Fatalf("ClaimNext() unexpected error: %v", err)
		}
		if job != nil {
			t.Fatalf("expected nil job from empty queue, got %+v", job)
		}
	})

	// Helper для вставки минимального media_asset.
	insertAsset := func(t *testing.T) uuid.UUID {
		t.Helper()
		assetID := uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO media_assets (
				id, original_filename, media_type, status, storage_path,
				size_bytes, metadata, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, now(), now())
		`, assetID, "test.mp4", entity.MediaTypeMP4, entity.StatusUploaded, "uploads/test.mp4", 1024, `{}`)
		if err != nil {
			t.Fatalf("failed to insert test media asset: %v", err)
		}
		cleanupAssetIDs = append(cleanupAssetIDs, assetID)
		return assetID
	}

	// Сценарий 2: Успешный захват pending-задачи.
	t.Run("claim queued job", func(t *testing.T) {
		assetID := insertAsset(t)
		jobID := uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO processing_jobs (
				id, asset_id, status, attempt, max_attempts, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, now(), now())
		`, jobID, assetID, entity.JobStatusQueued, 0, 3)
		if err != nil {
			t.Fatalf("failed to insert test processing job: %v", err)
		}
		cleanupJobIDs = append(cleanupJobIDs, jobID)

		job, err := repo.ClaimNext(ctx, "worker-1", 5*time.Minute)
		if err != nil {
			t.Fatalf("ClaimNext() unexpected error: %v", err)
		}
		if job == nil {
			t.Fatal("expected job, got nil")
		}
		if job.ID != jobID {
			t.Errorf("job ID = %v, want %v", job.ID, jobID)
		}
		if job.Status != entity.JobStatusProcessing {
			t.Errorf("status = %q, want %q", job.Status, entity.JobStatusProcessing)
		}
		if job.LeaseOwner == nil || *job.LeaseOwner != "worker-1" {
			t.Errorf("lease_owner = %v, want worker-1", job.LeaseOwner)
		}
		if job.LeaseExpiresAt == nil {
			t.Error("lease_expires_at should not be nil")
		}

		// Проверяем обновление в БД.
		var status string
		var leaseOwner *string
		err = pool.QueryRow(ctx, `
			SELECT status, lease_owner FROM processing_jobs WHERE id = $1
		`, jobID).Scan(&status, &leaseOwner)
		if err != nil {
			t.Fatalf("failed to query job after claim: %v", err)
		}
		if status != string(entity.JobStatusProcessing) {
			t.Errorf("db status = %q, want processing", status)
		}
		if leaseOwner == nil || *leaseOwner != "worker-1" {
			t.Errorf("db lease_owner = %v, want worker-1", leaseOwner)
		}
	})

	// Сценарий 3: Перехват просроченной processing-задачи.
	t.Run("claim expired processing job", func(t *testing.T) {
		assetID := insertAsset(t)
		jobID := uuid.New()
		expiredTime := time.Now().Add(-10 * time.Minute)
		_, err := pool.Exec(ctx, `
			INSERT INTO processing_jobs (
				id, asset_id, status, attempt, max_attempts,
				lease_owner, lease_expires_at, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
		`, jobID, assetID, entity.JobStatusProcessing, 0, 3, "old-worker", expiredTime)
		if err != nil {
			t.Fatalf("failed to insert test processing job: %v", err)
		}
		cleanupJobIDs = append(cleanupJobIDs, jobID)

		job, err := repo.ClaimNext(ctx, "worker-2", 5*time.Minute)
		if err != nil {
			t.Fatalf("ClaimNext() unexpected error: %v", err)
		}
		if job == nil {
			t.Fatal("expected job, got nil")
		}
		if job.ID != jobID {
			t.Errorf("job ID = %v, want %v", job.ID, jobID)
		}
		if job.Status != entity.JobStatusProcessing {
			t.Errorf("status = %q, want %q", job.Status, entity.JobStatusProcessing)
		}
		if job.LeaseOwner == nil || *job.LeaseOwner != "worker-2" {
			t.Errorf("lease_owner = %v, want worker-2", job.LeaseOwner)
		}
		if job.LeaseExpiresAt == nil || !job.LeaseExpiresAt.After(time.Now()) {
			t.Error("lease_expires_at should be in the future")
		}
	})
}
