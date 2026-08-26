// Package repository содержит интерфейсы (порты) для доступа к данным.
package repository

import (
	"context"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/google/uuid"
)

// MediaRepository описывает контракт для хранения и извлечения метаданных медиафайлов.
type MediaRepository interface {
	// Save сохраняет новый медиафайл.
	Save(ctx context.Context, asset *entity.MediaAsset) error

	// FindByID возвращает медиафайл по идентификатору.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.MediaAsset, error)

	// UpdateStatus атомарно изменяет статус медиафайла.
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.MediaStatus) error

	// UpdateMetadata обновляет метаданные медиафайла (ширина, высота, длительность, кодек).
	UpdateMetadata(ctx context.Context, id uuid.UUID, width, height int, duration int64, codec string) error
}
