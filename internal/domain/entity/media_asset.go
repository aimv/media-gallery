package entity

import (
	"time"

	"github.com/google/uuid"
)

// MediaAsset представляет медиафайл (изображение или видео) и его метаданные.
type MediaAsset struct {
	ID               uuid.UUID      // уникальный идентификатор
	OriginalFilename string         // исходное имя файла
	MediaType        MediaType      // тип медиафайла
	Status           MediaStatus    // текущий статус обработки
	StoragePath      string         // относительный путь к оригиналу
	HlsPath          string         // относительный путь к HLS-плейлисту (пусто, если не готово)
	SizeBytes        int64          // размер файла в байтах
	ChecksumSHA256   string         // SHA-256 чексумма файла
	Width            int            // ширина (для изображений и видео)
	Height           int            // высота (для изображений и видео)
	DurationMS       int64          // длительность в миллисекундах (для видео)
	Codec            string         // основной кодек (для видео)
	Metadata         map[string]any // дополнительные метаданные (JSONB)
	CreatedAt        time.Time      // дата создания
	UpdatedAt        time.Time      // дата последнего обновления
	DeletedAt        *time.Time     // дата мягкого удаления (nil, если активен)
}

// CanTransitionTo проверяет, допустим ли переход из текущего статуса в указанный.
func (a *MediaAsset) CanTransitionTo(next MediaStatus) bool {
	switch a.Status {
	case StatusUploaded:
		return next == StatusQueued
	case StatusQueued:
		return next == StatusProcessing
	case StatusProcessing:
		return next == StatusReady || next == StatusFailed
	case StatusReady:
		return next == StatusDeleting
	case StatusFailed:
		return next == StatusQueued || next == StatusDeleting
	case StatusDeleting:
		return false
	default:
		return false
	}
}
