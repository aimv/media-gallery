package media

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/aimv/media-gallery/internal/domain/service"
	"github.com/google/uuid"
)

// --- Mocks ---

type pvMockMediaRepository struct {
	asset              *entity.MediaAsset
	findErr            error
	statusUpdates      []entity.MediaStatus
	updateStatusErr    error
	updateMetadataCall bool
	updateMetadataErr  error
}

func (m *pvMockMediaRepository) Save(_ context.Context, _ *entity.MediaAsset) error { return nil }

func (m *pvMockMediaRepository) FindByID(_ context.Context, _ uuid.UUID) (*entity.MediaAsset, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.asset, nil
}

func (m *pvMockMediaRepository) UpdateStatus(_ context.Context, _ uuid.UUID, status entity.MediaStatus) error {
	m.statusUpdates = append(m.statusUpdates, status)
	return m.updateStatusErr
}

func (m *pvMockMediaRepository) UpdateMetadata(_ context.Context, _ uuid.UUID, _ int, _ int, _ int64, _ string) error {
	m.updateMetadataCall = true
	return m.updateMetadataErr
}

type pvMockQueueRepository struct {
	jobStatusUpdates []entity.JobStatus
	errMsgUpdates    []*string
	updateStatusErr  error
	extendLeaseCalls int
	extendLeaseErr   error
}

func (m *pvMockQueueRepository) ClaimNext(_ context.Context, _ string, _ time.Duration) (*entity.ProcessingJob, error) {
	return nil, nil
}

func (m *pvMockQueueRepository) UpdateStatus(_ context.Context, _ uuid.UUID, status entity.JobStatus, errMsg *string) error {
	m.jobStatusUpdates = append(m.jobStatusUpdates, status)
	if errMsg != nil {
		v := *errMsg
		m.errMsgUpdates = append(m.errMsgUpdates, &v)
	} else {
		m.errMsgUpdates = append(m.errMsgUpdates, nil)
	}
	return m.updateStatusErr
}

func (m *pvMockQueueRepository) ExtendLease(_ context.Context, _ uuid.UUID, _ time.Duration) error {
	m.extendLeaseCalls++
	return m.extendLeaseErr
}

type pvMockVideoProcessor struct {
	processErr   error
	processSleep time.Duration
	probeErr     error
	metadata     *service.VideoMetadata
}

func (m *pvMockVideoProcessor) Validate(_ context.Context, _ string) (*service.VideoMetadata, error) {
	return m.metadata, nil
}

func (m *pvMockVideoProcessor) ProcessToHLS(_ context.Context, _, _ string) error {
	if m.processSleep > 0 {
		time.Sleep(m.processSleep)
	}
	return m.processErr
}

func (m *pvMockVideoProcessor) ProbeMetadata(_ context.Context, _ string) (*service.VideoMetadata, error) {
	if m.probeErr != nil {
		return nil, m.probeErr
	}
	if m.metadata == nil {
		return &service.VideoMetadata{Width: 1920, Height: 1080, DurationMS: 10000, Codec: "h264"}, nil
	}
	return m.metadata, nil
}

// pvMockFileStorage — минимальная реализация repository.FileStorage для тестов.
type pvMockFileStorage struct {
	moveDirCalls int
	moveDirErr   error
}

func (m *pvMockFileStorage) Save(_ context.Context, _ string, _ io.Reader) (string, error) {
	return "", nil
}

func (m *pvMockFileStorage) Delete(_ context.Context, _ string) error { return nil }

func (m *pvMockFileStorage) MoveDir(_ context.Context, _, _ string) error {
	m.moveDirCalls++
	return m.moveDirErr
}

// --- Helpers ---

func equalMediaStatuses(got, want []entity.MediaStatus) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func equalJobStatuses(got, want []entity.JobStatus) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// --- Tests ---

func TestProcessVideoUseCase_Execute_Success(t *testing.T) {
	assetID := uuid.New()
	jobID := uuid.New()
	asset := &entity.MediaAsset{
		ID:          assetID,
		StoragePath: "uploads/test.mp4",
	}
	job := &entity.ProcessingJob{
		ID:      jobID,
		AssetID: assetID,
	}

	mediaRepo := &pvMockMediaRepository{asset: asset}
	queueRepo := &pvMockQueueRepository{}
	processor := &pvMockVideoProcessor{
		processSleep: 80 * time.Millisecond,
		metadata: &service.VideoMetadata{
			Width:      1920,
			Height:     1080,
			DurationMS: 10000,
			Codec:      "h264",
		},
	}
	storage := &pvMockFileStorage{}

	uc := NewProcessVideoUseCase(mediaRepo, queueRepo, processor, storage, t.TempDir())
	uc.heartbeatInterval = 20 * time.Millisecond

	if err := uc.Execute(context.Background(), job); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	wantAssetStatuses := []entity.MediaStatus{entity.StatusProcessing, entity.StatusReady}
	if !equalMediaStatuses(mediaRepo.statusUpdates, wantAssetStatuses) {
		t.Errorf("asset statuses = %v, want %v", mediaRepo.statusUpdates, wantAssetStatuses)
	}

	wantJobStatuses := []entity.JobStatus{entity.JobStatusProcessing, entity.JobStatusSuccess}
	if !equalJobStatuses(queueRepo.jobStatusUpdates, wantJobStatuses) {
		t.Errorf("job statuses = %v, want %v", queueRepo.jobStatusUpdates, wantJobStatuses)
	}

	if queueRepo.extendLeaseCalls == 0 {
		t.Error("ExtendLease was not called")
	}

	if !mediaRepo.updateMetadataCall {
		t.Error("UpdateMetadata was not called")
	}

	if storage.moveDirCalls == 0 {
		t.Error("MoveDir was not called")
	}
}

func TestProcessVideoUseCase_Execute_ProcessError(t *testing.T) {
	assetID := uuid.New()
	jobID := uuid.New()
	asset := &entity.MediaAsset{
		ID:          assetID,
		StoragePath: "uploads/test.mp4",
	}
	job := &entity.ProcessingJob{
		ID:      jobID,
		AssetID: assetID,
	}

	mediaRepo := &pvMockMediaRepository{asset: asset}
	queueRepo := &pvMockQueueRepository{}
	processor := &pvMockVideoProcessor{
		processErr: errors.New("ffmpeg crashed"),
	}
	storage := &pvMockFileStorage{}

	uc := NewProcessVideoUseCase(mediaRepo, queueRepo, processor, storage, t.TempDir())
	uc.heartbeatInterval = 20 * time.Millisecond

	err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}

	wantAssetStatuses := []entity.MediaStatus{entity.StatusProcessing, entity.StatusFailed}
	if !equalMediaStatuses(mediaRepo.statusUpdates, wantAssetStatuses) {
		t.Errorf("asset statuses = %v, want %v", mediaRepo.statusUpdates, wantAssetStatuses)
	}

	wantJobStatuses := []entity.JobStatus{entity.JobStatusProcessing, entity.JobStatusFailed}
	if !equalJobStatuses(queueRepo.jobStatusUpdates, wantJobStatuses) {
		t.Errorf("job statuses = %v, want %v", queueRepo.jobStatusUpdates, wantJobStatuses)
	}

	if len(queueRepo.errMsgUpdates) == 0 {
		t.Fatal("expected errMsg to be recorded for failed job")
	}
	lastErrMsg := queueRepo.errMsgUpdates[len(queueRepo.errMsgUpdates)-1]
	if lastErrMsg == nil {
		t.Fatal("expected non-nil error message on failed job")
	}
	wantErrMsg := "process to hls: ffmpeg crashed"
	if *lastErrMsg != wantErrMsg {
		t.Errorf("errMsg = %q, want %q", *lastErrMsg, wantErrMsg)
	}

	if storage.moveDirCalls != 0 {
		t.Error("MoveDir should not have been called on process error")
	}
}

func TestProcessVideoUseCase_Execute_ProbeError(t *testing.T) {
	assetID := uuid.New()
	jobID := uuid.New()
	asset := &entity.MediaAsset{
		ID:          assetID,
		StoragePath: "uploads/test.mp4",
	}
	job := &entity.ProcessingJob{
		ID:      jobID,
		AssetID: assetID,
	}

	mediaRepo := &pvMockMediaRepository{asset: asset}
	queueRepo := &pvMockQueueRepository{}
	processor := &pvMockVideoProcessor{
		probeErr: errors.New("invalid stream"),
	}
	storage := &pvMockFileStorage{}

	uc := NewProcessVideoUseCase(mediaRepo, queueRepo, processor, storage, t.TempDir())
	uc.heartbeatInterval = 20 * time.Millisecond

	err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}

	wantAssetStatuses := []entity.MediaStatus{entity.StatusProcessing, entity.StatusFailed}
	if !equalMediaStatuses(mediaRepo.statusUpdates, wantAssetStatuses) {
		t.Errorf("asset statuses = %v, want %v", mediaRepo.statusUpdates, wantAssetStatuses)
	}

	wantJobStatuses := []entity.JobStatus{entity.JobStatusProcessing, entity.JobStatusFailed}
	if !equalJobStatuses(queueRepo.jobStatusUpdates, wantJobStatuses) {
		t.Errorf("job statuses = %v, want %v", queueRepo.jobStatusUpdates, wantJobStatuses)
	}

	if mediaRepo.updateMetadataCall {
		t.Error("UpdateMetadata should not have been called on probe error")
	}

	if storage.moveDirCalls != 0 {
		t.Error("MoveDir should not have been called on probe error")
	}
}
