package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/aimv/media-gallery/internal/infrastructure/config"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// findProjectRoot поднимается вверх от текущей директории до каталога с go.mod.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found")
}

func TestIntegration_MediaRepository_Lifecycle(t *testing.T) {
	// Загружаем .env из корня проекта, т.к. текущая рабочая директория — пакетная.
	root, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to locate project root: %v", err)
	}
	_ = godotenv.Load(filepath.Join(root, ".env")) // если файла нет, не критично

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	pool, err := NewPool(cfg.TestDBDSN)
	if err != nil {
		t.Skipf("database not available, skipping integration test: %v", err)
	}
	defer pool.Close()

	// Применяем миграции для гарантии наличия схемы.
	migrationsPath := "file://" + filepath.Join(root, "internal/infrastructure/persistence/postgres/migrations")
	if err := RunMigrations(cfg.TestDBDSN, migrationsPath); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewMediaRepository(pool)
	ctx := context.Background()

	assetID := uuid.New()
	asset := &entity.MediaAsset{
		ID:               assetID,
		OriginalFilename: "integration-test.png",
		MediaType:        entity.MediaTypePNG,
		Status:           entity.StatusUploaded,
		StoragePath:      "uploads/integration-test.png",
		SizeBytes:        12345,
		ChecksumSHA256:   "dummychecksum",
		Metadata:         make(map[string]any),
	}

	// Очистка тестовых данных после завершения.
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM media_assets WHERE id = $1`, assetID)
	})

	// 1. Save
	if err := repo.Save(ctx, asset); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	// 2. FindByID
	fetched, err := repo.FindByID(ctx, assetID)
	if err != nil {
		t.Fatalf("FindByID() unexpected error: %v", err)
	}
	if fetched == nil {
		t.Fatal("FindByID() returned nil asset")
	}
	if fetched.OriginalFilename != asset.OriginalFilename {
		t.Errorf("OriginalFilename = %q, want %q", fetched.OriginalFilename, asset.OriginalFilename)
	}
	if fetched.SizeBytes != asset.SizeBytes {
		t.Errorf("SizeBytes = %d, want %d", fetched.SizeBytes, asset.SizeBytes)
	}
	if fetched.Status != asset.Status {
		t.Errorf("Status = %q, want %q", fetched.Status, asset.Status)
	}
	if len(fetched.Metadata) != 0 {
		t.Errorf("Metadata should be empty map, got %v", fetched.Metadata)
	}

	// 3. UpdateStatus
	if err := repo.UpdateStatus(ctx, assetID, entity.StatusQueued); err != nil {
		t.Fatalf("UpdateStatus() unexpected error: %v", err)
	}
	fetched, err = repo.FindByID(ctx, assetID)
	if err != nil {
		t.Fatalf("FindByID after UpdateStatus() error: %v", err)
	}
	if fetched.Status != entity.StatusQueued {
		t.Errorf("Status after update = %q, want %q", fetched.Status, entity.StatusQueued)
	}

	// 4. UpdateMetadata
	width, height := 1920, 1080
	duration := int64(60000)
	codec := "h264"
	if err := repo.UpdateMetadata(ctx, assetID, width, height, duration, codec); err != nil {
		t.Fatalf("UpdateMetadata() unexpected error: %v", err)
	}
	fetched, err = repo.FindByID(ctx, assetID)
	if err != nil {
		t.Fatalf("FindByID after UpdateMetadata() error: %v", err)
	}
	if fetched.Width != width {
		t.Errorf("Width = %d, want %d", fetched.Width, width)
	}
	if fetched.Height != height {
		t.Errorf("Height = %d, want %d", fetched.Height, height)
	}
	if fetched.DurationMS != duration {
		t.Errorf("DurationMS = %d, want %d", fetched.DurationMS, duration)
	}
	if fetched.Codec != codec {
		t.Errorf("Codec = %q, want %q", fetched.Codec, codec)
	}
}
