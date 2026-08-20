package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aimv/media-gallery/internal/domain/media"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaRepository struct {
	pool *pgxpool.Pool
}

func NewMediaRepository(pool *pgxpool.Pool) *MediaRepository {
	return &MediaRepository{pool: pool}
}

// Save сохраняет медиа-ассет. Если это видео (media_type = 'video/mp4'),
// то в рамках той же транзакции создаётся запись в processing_jobs.
func (r *MediaRepository) Save(ctx context.Context, asset *media.MediaAsset) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO media_assets (
			id, original_file_name, media_type, status,
			storage_path, hls_path, size_bytes, checksum_sha256,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		asset.ID,
		asset.OriginalFileName,
		asset.MediaType,
		asset.Status,
		asset.StoragePath,
		asset.HlsPath,
		asset.SizeBytes,
		asset.ChecksumSHA256,
		asset.Metadata,
		asset.CreatedAt,
		asset.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert media_asset: %w", err)
	}

	if asset.MediaType == media.MediaTypeMP4 {
		_, err = tx.Exec(ctx, `
			INSERT INTO processing_jobs (
				id, asset_id, status, attempt, max_attempts,
				created_at, updated_at
			) VALUES (gen_random_uuid(), $1, 'queued', 0, 3, now(), now())
		`, asset.ID)
		if err != nil {
			return fmt.Errorf("insert processing_job: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *MediaRepository) GetByID(ctx context.Context, id string) (*media.MediaAsset, error) {
	var asset media.MediaAsset
	err := r.pool.QueryRow(ctx, `
		SELECT id, original_file_name, media_type, status,
		       storage_path, hls_path, size_bytes, checksum_sha256,
		       metadata, created_at, updated_at
		FROM media_assets
		WHERE id = $1
	`, id).Scan(
		&asset.ID,
		&asset.OriginalFileName,
		&asset.MediaType,
		&asset.Status,
		&asset.StoragePath,
		&asset.HlsPath,
		&asset.SizeBytes,
		&asset.ChecksumSHA256,
		&asset.Metadata,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("asset not found: %w", err)
		}
		return nil, fmt.Errorf("select asset: %w", err)
	}
	return &asset, nil
}

func (r *MediaRepository) List(ctx context.Context, filter media.ListFilter) ([]*media.MediaAsset, error) {
	query := `SELECT id, original_file_name, media_type, status,
		       storage_path, hls_path, size_bytes, checksum_sha256,
		       metadata, created_at, updated_at
		FROM media_assets WHERE 1=1`
	args := []any{}
	argPos := 1

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filter.Status)
		argPos++
	}
	if filter.Type != nil {
		query += fmt.Sprintf(" AND media_type = $%d", argPos)
		args = append(args, *filter.Type)
		argPos++
	}
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
		argPos++
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}
	defer rows.Close()

	var assets []*media.MediaAsset
	for rows.Next() {
		var a media.MediaAsset
		err := rows.Scan(
			&a.ID,
			&a.OriginalFileName,
			&a.MediaType,
			&a.Status,
			&a.StoragePath,
			&a.HlsPath,
			&a.SizeBytes,
			&a.ChecksumSHA256,
			&a.Metadata,
			&a.CreatedAt,
			&a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		assets = append(assets, &a)
	}
	return assets, rows.Err()
}

func (r *MediaRepository) UpdateStatus(ctx context.Context, id string, status media.AssetStatus) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE media_assets SET status = $1, updated_at = now()
		WHERE id = $2
	`, status, id)
	if err != nil {
		return fmt.Errorf("update asset status: %w", err)
	}
	return nil
}

func (r *MediaRepository) UpdateHlsPath(ctx context.Context, id string, hlsPath *string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE media_assets SET hls_path = $1, updated_at = now()
		WHERE id = $2
	`, hlsPath, id)
	if err != nil {
		return fmt.Errorf("update hls path: %w", err)
	}
	return nil
}

func (r *MediaRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM media_assets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete asset: %w", err)
	}
	return nil
}

// ==================== JobQueue ====================

type JobQueue struct {
	pool *pgxpool.Pool
}

func NewJobQueue(pool *pgxpool.Pool) *JobQueue {
	return &JobQueue{pool: pool}
}

func (q *JobQueue) Enqueue(ctx context.Context, assetID string) (*media.ProcessingJob, error) {
	var job media.ProcessingJob
	err := q.pool.QueryRow(ctx, `
		INSERT INTO processing_jobs (
			id, asset_id, status, attempt, max_attempts, created_at, updated_at
		) VALUES (gen_random_uuid(), $1, 'queued', 0, 3, now(), now())
		RETURNING id, asset_id, status, attempt, max_attempts, lease_owner, lease_expires_at, last_error, created_at, started_at, finished_at
	`, assetID).Scan(
		&job.ID,
		&job.AssetID,
		&job.Status,
		&job.Attempt,
		&job.MaxAttempts,
		&job.LeaseOwner,
		&job.LeaseExpiresAt,
		&job.LastError,
		&job.CreatedAt,
		&job.StartedAt,
		&job.FinishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("enqueue job: %w", err)
	}
	return &job, nil
}

func (q *JobQueue) Claim(ctx context.Context, owner string, leaseDuration time.Duration) (*media.ProcessingJob, error) {
	tx, err := q.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var job media.ProcessingJob
	err = tx.QueryRow(ctx, `
		UPDATE processing_jobs
		SET status = 'processing',
		    lease_owner = $1,
		    lease_expires_at = now() + $2::interval,
		    started_at = COALESCE(started_at, now()),
		    updated_at = now()
		WHERE id = (
			SELECT id
			FROM processing_jobs
			WHERE status = 'queued'
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, asset_id, status, attempt, max_attempts, lease_owner, lease_expires_at, last_error, created_at, started_at, finished_at
	`, owner, leaseDuration).Scan(
		&job.ID,
		&job.AssetID,
		&job.Status,
		&job.Attempt,
		&job.MaxAttempts,
		&job.LeaseOwner,
		&job.LeaseExpiresAt,
		&job.LastError,
		&job.CreatedAt,
		&job.StartedAt,
		&job.FinishedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // no job available
		}
		return nil, fmt.Errorf("claim job: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return &job, nil
}

func (q *JobQueue) Heartbeat(ctx context.Context, jobID string, owner string, leaseDuration time.Duration) error {
	ct, err := q.pool.Exec(ctx, `
		UPDATE processing_jobs
		SET lease_expires_at = now() + $1::interval,
		    updated_at = now()
		WHERE id = $2 AND lease_owner = $3 AND status = 'processing'
	`, leaseDuration, jobID, owner)
	if err != nil {
		return fmt.Errorf("heartbeat: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("heartbeat: no rows affected")
	}
	return nil
}

func (q *JobQueue) MarkDone(ctx context.Context, jobID string, owner string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE processing_jobs
		SET status = 'done',
		    finished_at = now(),
		    updated_at = now()
		WHERE id = $1 AND lease_owner = $2 AND status = 'processing'
	`, jobID, owner)
	if err != nil {
		return fmt.Errorf("mark done: %w", err)
	}
	return nil
}

func (q *JobQueue) MarkFailed(ctx context.Context, jobID string, owner string, errMsg string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE processing_jobs
		SET status = 'failed',
		    last_error = $1,
		    finished_at = now(),
		    updated_at = now()
		WHERE id = $2 AND lease_owner = $3 AND status = 'processing'
	`, errMsg, jobID, owner)
	if err != nil {
		return fmt.Errorf("mark failed: %w", err)
	}
	return nil
}
