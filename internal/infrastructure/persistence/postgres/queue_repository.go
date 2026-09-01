package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/aimv/media-gallery/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Проверка на этапе компиляции.
var _ repository.QueueRepository = (*QueueRepository)(nil)

// QueueRepository реализует repository.QueueRepository поверх PostgreSQL.
type QueueRepository struct {
	pool *pgxpool.Pool
}

// NewQueueRepository создаёт репозиторий очереди задач.
func NewQueueRepository(pool *pgxpool.Pool) *QueueRepository {
	return &QueueRepository{pool: pool}
}

// ClaimNext выбирает следующую доступную задачу с блокировкой SKIP LOCKED.
func (r *QueueRepository) ClaimNext(ctx context.Context, workerID string, leaseDuration time.Duration) (*entity.ProcessingJob, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row := tx.QueryRow(ctx, `
		SELECT id, asset_id, status, attempt, max_attempts,
		       lease_owner, lease_expires_at, error,
		       created_at, started_at, finished_at, updated_at
		FROM processing_jobs
		WHERE status = $1
		   OR (status = $2 AND lease_expires_at < now())
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`, entity.JobStatusQueued, entity.JobStatusProcessing)

	var job entity.ProcessingJob
	err = row.Scan(
		&job.ID,
		&job.AssetID,
		&job.Status,
		&job.Attempt,
		&job.MaxAttempts,
		&job.LeaseOwner,
		&job.LeaseExpiresAt,
		&job.Error,
		&job.CreatedAt,
		&job.StartedAt,
		&job.FinishedAt,
		&job.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select next job: %w", err)
	}

	now := time.Now()
	newLeaseExpires := now.Add(leaseDuration)

	_, err = tx.Exec(ctx, `
		UPDATE processing_jobs
		SET status = $2,
		    lease_owner = $3,
		    lease_expires_at = $4,
		    started_at = COALESCE(started_at, $5),
		    updated_at = $5
		WHERE id = $1
	`,
		job.ID,
		entity.JobStatusProcessing,
		workerID,
		newLeaseExpires,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("update job lease: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	job.Status = entity.JobStatusProcessing
	job.LeaseOwner = &workerID
	job.LeaseExpiresAt = &newLeaseExpires
	if job.StartedAt == nil {
		job.StartedAt = &now
	}
	job.UpdatedAt = now

	return &job, nil
}

// UpdateStatus изменяет статус задачи обработки.
func (r *QueueRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.JobStatus, errMsg *string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE processing_jobs
		SET status = $2::job_status,
		    error = $3,
		    updated_at = now(),
		    finished_at = CASE
		        WHEN $2::job_status IN ('success', 'failed') THEN COALESCE(finished_at, now())
		        ELSE finished_at
		    END
		WHERE id = $1
	`, id, status, errMsg)
	if err != nil {
		return fmt.Errorf("update job status: %w", err)
	}
	return nil
}

// ExtendLease продлевает аренду задачи, если она всё ещё находится в обработке.
func (r *QueueRepository) ExtendLease(ctx context.Context, id uuid.UUID, duration time.Duration) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE processing_jobs
		SET lease_expires_at = now() + make_interval(secs => $2),
		    updated_at = now()
		WHERE id = $1 AND status = $3
	`, id, duration.Seconds(), entity.JobStatusProcessing)
	if err != nil {
		return fmt.Errorf("extend lease: %w", err)
	}
	return nil
}
