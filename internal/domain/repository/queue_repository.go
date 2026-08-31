package repository

import (
	"context"
	"time"

	"github.com/aimv/media-gallery/internal/domain/entity"
)

// QueueRepository описывает контракт для доступа к очереди задач обработки.
type QueueRepository interface {
	// ClaimNext атомарно выбирает следующую доступную задачу,
	// блокируя её за воркером (workerID) на время leaseDuration.
	// Если свободных задач нет, возвращает (nil, nil).
	ClaimNext(ctx context.Context, workerID string, leaseDuration time.Duration) (*entity.ProcessingJob, error)
}
