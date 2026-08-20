package media

import "time"

type JobStatus string

const (
	JobStatusQueued     JobStatus = "queued"
	JobStatusProcessing JobStatus = "processing"
	JobStatusDone       JobStatus = "done"
	JobStatusFailed     JobStatus = "failed"
)

type ProcessingJob struct {
	ID             string
	AssetID        string
	Status         JobStatus
	Attempt        int
	MaxAttempts    int
	LeaseOwner     string
	LeaseExpiresAt time.Time
	LastError      *string // изменено с string на *string для поддержки NULL
	CreatedAt      time.Time
	StartedAt      *time.Time
	FinishedAt     *time.Time
}
