package media

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockMediaRepository struct {
	saveCalled bool
	saveErr    error
	savedAsset *entity.MediaAsset
}

func (m *mockMediaRepository) Save(_ context.Context, asset *entity.MediaAsset) error {
	m.saveCalled = true
	m.savedAsset = asset
	return m.saveErr
}

func (m *mockMediaRepository) FindByID(_ context.Context, _ uuid.UUID) (*entity.MediaAsset, error) {
	return nil, nil
}

func (m *mockMediaRepository) UpdateStatus(_ context.Context, _ uuid.UUID, _ entity.MediaStatus) error {
	return nil
}

func (m *mockMediaRepository) UpdateMetadata(_ context.Context, _ uuid.UUID, _ int, _ int, _ int64, _ string) error {
	return nil
}

type mockFileStorage struct {
	saveCalled bool
	savePath   string
	saveErr    error
}

func (m *mockFileStorage) Save(_ context.Context, _ string, _ io.Reader) (string, error) {
	m.saveCalled = true
	return m.savePath, m.saveErr
}

func (m *mockFileStorage) Delete(_ context.Context, _ string) error {
	return nil
}

// --- Tests ---

func TestUploadUseCase_Execute(t *testing.T) {
	tests := []struct {
		name              string
		contentType       string
		size              int64
		storagePath       string
		storageErr        error
		repoErr           error
		wantErr           bool
		wantRepoSaved     bool
		wantStorageCalled bool
	}{
		{
			name:              "success",
			contentType:       "image/png",
			size:              1024,
			storagePath:       "uploads/test.png",
			storageErr:        nil,
			repoErr:           nil,
			wantErr:           false,
			wantRepoSaved:     true,
			wantStorageCalled: true,
		},
		{
			name:              "storage error",
			contentType:       "image/png",
			size:              1024,
			storagePath:       "",
			storageErr:        errors.New("disk full"),
			repoErr:           nil,
			wantErr:           true,
			wantRepoSaved:     false,
			wantStorageCalled: true,
		},
		{
			name:              "invalid content type",
			contentType:       "text/plain",
			size:              1024,
			storagePath:       "",
			storageErr:        nil,
			repoErr:           nil,
			wantErr:           true,
			wantRepoSaved:     false,
			wantStorageCalled: false,
		},
		{
			name:              "repository error",
			contentType:       "image/jpeg",
			size:              1024,
			storagePath:       "uploads/test.jpg",
			storageErr:        nil,
			repoErr:           errors.New("db error"),
			wantErr:           true,
			wantRepoSaved:     true,
			wantStorageCalled: true,
		},
		{
			name:              "file size limit exceeded",
			contentType:       "video/mp4",
			size:              entity.MaxUploadSizeBytes + 1,
			storagePath:       "",
			storageErr:        nil,
			repoErr:           nil,
			wantErr:           true,
			wantRepoSaved:     false,
			wantStorageCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMediaRepository{saveErr: tt.repoErr}
			storage := &mockFileStorage{savePath: tt.storagePath, saveErr: tt.storageErr}

			uc := NewUploadUseCase(repo, storage)
			asset, err := uc.Execute(context.Background(), "test.bin", tt.contentType, tt.size, bytes.NewReader([]byte("data")))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if asset == nil {
					t.Fatal("expected asset, got nil")
				}
				if asset.MediaType != entity.MediaType(tt.contentType) {
					t.Errorf("media type = %q, want %q", asset.MediaType, tt.contentType)
				}
				if asset.StoragePath != tt.storagePath {
					t.Errorf("storage path = %q, want %q", asset.StoragePath, tt.storagePath)
				}
				if asset.ID == uuid.Nil {
					t.Error("expected non-zero asset ID")
				}
			}

			if repo.saveCalled != tt.wantRepoSaved {
				t.Errorf("repo.saveCalled = %v, want %v", repo.saveCalled, tt.wantRepoSaved)
			}

			if storage.saveCalled != tt.wantStorageCalled {
				t.Errorf("storage.saveCalled = %v, want %v", storage.saveCalled, tt.wantStorageCalled)
			}
		})
	}
}
