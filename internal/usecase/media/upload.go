// Package media содержит use-case'ы для работы с медиафайлами.
package media

import (
	"context"
	"io"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/aimv/media-gallery/internal/domain/repository"
	"github.com/aimv/media-gallery/internal/pkg/apperror"
	"github.com/google/uuid"
)

// UploadUseCase реализует сценарий загрузки и сохранения медиафайла.
type UploadUseCase struct {
	repo    repository.MediaRepository
	storage repository.FileStorage
}

// NewUploadUseCase создаёт use-case с указанными зависимостями.
func NewUploadUseCase(repo repository.MediaRepository, storage repository.FileStorage) *UploadUseCase {
	return &UploadUseCase{repo: repo, storage: storage}
}

// Execute выполняет потоковую загрузку файла, валидацию типа,
// сохранение на диск и запись метаданных в БД.
func (u *UploadUseCase) Execute(ctx context.Context, filename string, contentType string, size int64, src io.Reader) (*entity.MediaAsset, error) {
	mediaType := entity.MediaType(contentType)
	if !mediaType.IsValid() {
		return nil, apperror.ErrInvalidInput
	}

	asset := &entity.MediaAsset{
		ID:               uuid.New(),
		OriginalFilename: filename,
		MediaType:        mediaType,
		Status:           entity.StatusUploaded,
		SizeBytes:        size,
		Metadata:         make(map[string]any),
	}

	storagePath, err := u.storage.Save(ctx, filename, src)
	if err != nil {
		return nil, err
	}
	asset.StoragePath = storagePath

	if err := u.repo.Save(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}
