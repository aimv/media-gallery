// Package entity содержит доменные сущности и перечисления.
package entity

// MediaType определяет допустимый тип медиафайла.
type MediaType string

// Допустимые типы медиафайлов, поддерживаемые системой.
const (
	MediaTypeJPEG MediaType = "image/jpeg"
	MediaTypePNG  MediaType = "image/png"
	MediaTypeMP4  MediaType = "video/mp4"
)

// IsValid проверяет, что тип медиафайла входит в список допустимых.
func (t MediaType) IsValid() bool {
	switch t {
	case MediaTypeJPEG, MediaTypePNG, MediaTypeMP4:
		return true
	default:
		return false
	}
}

// MediaStatus описывает жизненный цикл обработки медиафайла.
type MediaStatus string

// Статусы жизненного цикла обработки медиафайлов.
const (
	StatusUploaded   MediaStatus = "uploaded"
	StatusQueued     MediaStatus = "queued"
	StatusProcessing MediaStatus = "processing"
	StatusReady      MediaStatus = "ready"
	StatusFailed     MediaStatus = "failed"
	StatusDeleting   MediaStatus = "deleting"
)

// IsValid проверяет, что статус входит в список допустимых.
func (s MediaStatus) IsValid() bool {
	switch s {
	case StatusUploaded, StatusQueued, StatusProcessing, StatusReady, StatusFailed, StatusDeleting:
		return true
	default:
		return false
	}
}
