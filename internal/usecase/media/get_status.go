package media

import (
	"context"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/aimv/media-gallery/internal/domain/repository"
	"github.com/google/uuid"
)

// GetStatusUseCase реализует сценарий получения текущего состояния медиафайла.
type GetStatusUseCase struct {
	mediaRepo repository.MediaRepository
}

// NewGetStatusUseCase создаёт use-case с указанной зависимостью.
func NewGetStatusUseCase(mediaRepo repository.MediaRepository) *GetStatusUseCase {
	return &GetStatusUseCase{mediaRepo: mediaRepo}
}

// Execute возвращает медиафайл по идентификатору.
// Если ассет не найден, репозиторий возвращает apperror.ErrNotFound.
func (u *GetStatusUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.MediaAsset, error) {
	return u.mediaRepo.FindByID(ctx, id)
}
