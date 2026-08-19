package media

import (
	"context"
	"io"
	"time"
)

type MediaRepository interface {
	Save(ctx context.Context, asset *MediaAsset) error
	GetByID(ctx context.Context, id string) (*MediaAsset, error)
	List(ctx context.Context, filter ListFilter) ([]*MediaAsset, error)
	UpdateStatus(ctx context.Context, id string, status AssetStatus) error
	Delete(ctx context.Context, id string) error
}

type ListFilter struct {
	Status *AssetStatus
	Type   *MediaType
	Limit  int
	Offset int
}

type JobQueue interface {
	Enqueue(ctx context.Context, assetID string) (*ProcessingJob, error)
	Claim(ctx context.Context, owner string, leaseDuration time.Duration) (*ProcessingJob, error)
	Heartbeat(ctx context.Context, jobID string, owner string, leaseDuration time.Duration) error
	MarkDone(ctx context.Context, jobID string, owner string) error
	MarkFailed(ctx context.Context, jobID string, owner string, errMsg string) error
}

type VideoProcessor interface {
	Probe(ctx context.Context, filePath string) (*VideoMetadata, error)
	ConvertToHLS(ctx context.Context, inputPath, outputDir string) error
}

type VideoMetadata struct {
	Duration   time.Duration
	Width      int
	Height     int
	VideoCodec string
	AudioCodec string
}

type FileStorage interface {
	Save(ctx context.Context, src io.Reader, dstPath string) error
	Delete(ctx context.Context, path string) error
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Move(ctx context.Context, src, dst string) error
}
