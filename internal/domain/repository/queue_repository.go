package repository

import (
	"context"
	"time"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/google/uuid"
)

// QueueRepository описывает контракт для доступа к очереди задач обработки.
type QueueRepository interface {
	// ClaimNext атомарно выбирает следующую доступную задачу,
	// блокируя её за воркером (workerID) на время leaseDuration.
	// Если свободных задач нет, возвращает (nil, nil).
	ClaimNext(ctx context.Context, workerID string, leaseDuration time.Duration) (*entity.ProcessingJob, error)

	// UpdateStatus изменяет статус задачи. Если errMsg != nil, сохраняет текст ошибки.
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.JobStatus, errMsg *string) error

	// ExtendLease продлевает аренду задачи на duration, только если она в статусе processing.
	ExtendLease(ctx context.Context, id uuid.UUID, duration time.Duration) error
}
