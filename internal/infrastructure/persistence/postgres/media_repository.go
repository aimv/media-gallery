package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/aimv/media-gallery/internal/domain/repository"
	"github.com/aimv/media-gallery/internal/pkg/apperror"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Проверка на этапе компиляции, что MediaRepository реализует доменный интерфейс.
var _ repository.MediaRepository = (*MediaRepository)(nil)

// MediaRepository реализует repository.MediaRepository на базе PostgreSQL.
type MediaRepository struct {
	pool *pgxpool.Pool
}

// NewMediaRepository создаёт репозиторий с указанным пулом соединений.
func NewMediaRepository(pool *pgxpool.Pool) *MediaRepository {
	return &MediaRepository{pool: pool}
}

// Save вставляет новую запись медиафайла в таблицу media_assets.
func (r *MediaRepository) Save(ctx context.Context, asset *entity.MediaAsset) error {
	metadataJSON, err := json.Marshal(asset.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	now := time.Now()
	asset.CreatedAt = now
	asset.UpdatedAt = now

	_, err = r.pool.Exec(ctx, `
		INSERT INTO media_assets (
			id, original_filename, media_type, status, storage_path, hls_path,
			size_bytes, checksum_sha256, width, height, duration_ms, codec,
			metadata, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`,
		asset.ID,
		asset.OriginalFilename,
		asset.MediaType,
		asset.Status,
		asset.StoragePath,
		asset.HlsPath,
		asset.SizeBytes,
		asset.ChecksumSHA256,
		asset.Width,
		asset.Height,
		asset.DurationMS,
		asset.Codec,
		metadataJSON,
		asset.CreatedAt,
		asset.UpdatedAt,
		asset.DeletedAt,
	)
	if err != nil {
		return fmt.Errorf("insert media asset: %w", err)
	}
	return nil
}

// FindByID возвращает медиафайл по ID, исключая мягко удалённые.
func (r *MediaRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.MediaAsset, error) {
	var asset entity.MediaAsset
	var metadataJSON []byte

	err := r.pool.QueryRow(ctx, `
		SELECT id, original_filename, media_type, status, storage_path, hls_path,
		       size_bytes, checksum_sha256, width, height, duration_ms, codec,
		       metadata, created_at, updated_at, deleted_at
		FROM media_assets
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&asset.ID,
		&asset.OriginalFilename,
		&asset.MediaType,
		&asset.Status,
		&asset.StoragePath,
		&asset.HlsPath,
		&asset.SizeBytes,
		&asset.ChecksumSHA256,
		&asset.Width,
		&asset.Height,
		&asset.DurationMS,
		&asset.Codec,
		&metadataJSON,
		&asset.CreatedAt,
		&asset.UpdatedAt,
		&asset.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select media asset: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &asset.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	return &asset, nil
}

// UpdateStatus обновляет статус медиафайла, проставляя время изменения.
func (r *MediaRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.MediaStatus) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE media_assets
		SET status = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, id, status)
	if err != nil {
		return fmt.Errorf("update media status: %w", err)
	}
	return nil
}

// UpdateMetadata обновляет технические метаданные медиафайла.
func (r *MediaRepository) UpdateMetadata(ctx context.Context, id uuid.UUID, width, height int, duration int64, codec string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE media_assets
		SET width = $2, height = $3, duration_ms = $4, codec = $5, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, id, width, height, duration, codec)
	if err != nil {
		return fmt.Errorf("update media metadata: %w", err)
	}
	return nil
}
