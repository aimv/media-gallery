package entity

import (
	"time"

	"github.com/google/uuid"
)

// MediaAsset представляет медиафайл (изображение или видео) и его метаданные.
type MediaAsset struct {
	ID               uuid.UUID      `json:"id"`                        // уникальный идентификатор
	OriginalFilename string         `json:"original_filename"`         // исходное имя файла
	MediaType        MediaType      `json:"media_type"`                // тип медиафайла
	Status           MediaStatus    `json:"status"`                    // текущий статус обработки
	StoragePath      string         `json:"storage_path"`              // относительный путь к оригиналу
	HlsPath          string         `json:"hls_path,omitempty"`        // относительный путь к HLS-плейлисту (пусто, если не готово)
	SizeBytes        int64          `json:"size_bytes"`                // размер файла в байтах
	ChecksumSHA256   string         `json:"checksum_sha256,omitempty"` // SHA-256 чексумма файла
	Width            int            `json:"width,omitempty"`           // ширина (для изображений и видео)
	Height           int            `json:"height,omitempty"`          // высота (для изображений и видео)
	DurationMS       int64          `json:"duration_ms,omitempty"`     // длительность в миллисекундах (для видео)
	Codec            string         `json:"codec,omitempty"`           // основной кодек (для видео)
	Metadata         map[string]any `json:"metadata"`                  // дополнительные метаданные (JSONB)
	CreatedAt        time.Time      `json:"created_at"`                // дата создания
	UpdatedAt        time.Time      `json:"updated_at"`                // дата последнего обновления
	DeletedAt        *time.Time     `json:"deleted_at,omitempty"`      // дата мягкого удаления (nil, если активен)
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
