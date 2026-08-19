package media

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aimv/media-gallery/internal/domain/media"
	"github.com/google/uuid"
)

type MediaUsecase struct {
	storage media.FileStorage
	repo    media.MediaRepository
}

func NewMediaUsecase(storage media.FileStorage, repo media.MediaRepository) *MediaUsecase {
	return &MediaUsecase{
		storage: storage,
		repo:    repo,
	}
}

type UploadInput struct {
	File        multipart.File
	FileHeader  *multipart.FileHeader
	ContentType string // можно получить из FileHeader.Header.Get("Content-Type")
	Size        int64
}

func (u *MediaUsecase) Upload(ctx context.Context, input UploadInput) (*media.MediaAsset, error) {
	// Валидация типа по расширению и MIME
	ext := strings.ToLower(filepath.Ext(input.FileHeader.Filename))
	contentType := input.FileHeader.Header.Get("Content-Type")

	var mediaType media.MediaType
	switch {
	case ext == ".jpg" || ext == ".jpeg" || contentType == "image/jpeg":
		mediaType = media.MediaTypeJPEG
	case ext == ".png" || contentType == "image/png":
		mediaType = media.MediaTypePNG
	case ext == ".mp4" || contentType == "video/mp4":
		mediaType = media.MediaTypeMP4
	default:
		return nil, fmt.Errorf("unsupported media type: %s", contentType)
	}

	// Дополнительная проверка magic bytes (для простоты можно пропустить, но оставим)
	// В реальном проекте использовать http.DetectContentType

	// Генерация ID и путей
	assetID := newUUID()
	originalExt := filepath.Ext(input.FileHeader.Filename)
	storagePath := filepath.Join("originals", assetID, "file"+originalExt)

	// Вычисление SHA256 и сохранение файла
	hasher := sha256.New()
	tee := io.TeeReader(input.File, hasher)

	if err := u.storage.Save(ctx, tee, storagePath); err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))

	// Статус для изображений сразу ready, для видео – queued (job будет создан в repo.Save)
	status := media.StatusReady
	if mediaType == media.MediaTypeMP4 {
		status = media.StatusQueued
	}

	now := time.Now().UTC()
	asset := &media.MediaAsset{
		ID:               assetID,
		OriginalFileName: input.FileHeader.Filename,
		MediaType:        mediaType,
		Status:           status,
		StoragePath:      storagePath,
		SizeBytes:        input.Size,
		ChecksumSHA256:   checksum,
		Metadata:         map[string]any{},
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := u.repo.Save(ctx, asset); err != nil {
		// Компенсация: удалить сохранённый файл
		_ = u.storage.Delete(ctx, storagePath)
		return nil, fmt.Errorf("save asset to db: %w", err)
	}

	return asset, nil
}

func newUUID() string {
	// Генерация UUID v4 (используем github.com/google/uuid)
	return uuid.New().String()
}
