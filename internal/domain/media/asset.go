package media

import "time"

type MediaType string

const (
	MediaTypeJPEG MediaType = "image/jpeg"
	MediaTypePNG  MediaType = "image/png"
	MediaTypeMP4  MediaType = "video/mp4"
)

type AssetStatus string

const (
	StatusUploaded   AssetStatus = "uploaded"
	StatusQueued     AssetStatus = "queued"
	StatusProcessing AssetStatus = "processing"
	StatusReady      AssetStatus = "ready"
	StatusFailed     AssetStatus = "failed"
	StatusDeleting   AssetStatus = "deleting"
)

type MediaAsset struct {
	ID               string
	OriginalFileName string
	MediaType        MediaType
	Status           AssetStatus
	StoragePath      string
	HlsPath          *string
	SizeBytes        int64
	ChecksumSHA256   string
	Metadata         map[string]any
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
