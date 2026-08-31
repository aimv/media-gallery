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

// JobStatus описывает состояние задачи обработки.
type JobStatus string

// Статусы задач обработки медиафайлов.
const (
	JobStatusQueued     JobStatus = "queued"
	JobStatusProcessing JobStatus = "processing"
	JobStatusSuccess    JobStatus = "success"
	JobStatusFailed     JobStatus = "failed"
)

// IsValid проверяет, что статус задачи входит в допустимые.
func (s JobStatus) IsValid() bool {
	switch s {
	case JobStatusQueued, JobStatusProcessing, JobStatusSuccess, JobStatusFailed:
		return true
	default:
		return false
	}
}
