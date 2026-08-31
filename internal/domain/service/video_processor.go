// Package service содержит интерфейсы внешних сервисов, используемых доменом.
package service

import "context"

// VideoMetadata содержит извлечённые технические метаданные видеофайла.
type VideoMetadata struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	DurationMS int64  `json:"duration_ms"`
	Codec      string `json:"codec"`
}

// VideoProcessor описывает контракт для валидации и обработки видео.
type VideoProcessor interface {
	// Validate проверяет видеофайл и возвращает его метаданные.
	Validate(ctx context.Context, filePath string) (*VideoMetadata, error)

	// ProcessToHLS конвертирует видеофайл в HLS-поток в указанной директории.
	ProcessToHLS(ctx context.Context, filePath, outputDir string) error

	// ProbeMetadata извлекает метаданные видеофайла.
	ProbeMetadata(ctx context.Context, filePath string) (*VideoMetadata, error)
}
