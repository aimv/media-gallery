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

func (m *mockMediaRepository) Save(ctx context.Context, asset *entity.MediaAsset) error {
	_ = ctx
	m.saveCalled = true
	m.savedAsset = asset
	return m.saveErr
}

func (m *mockMediaRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.MediaAsset, error) {
	_ = ctx
	_ = id
	return nil, nil
}

func (m *mockMediaRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.MediaStatus) error {
	_ = ctx
	_ = id
	_ = status
	return nil
}

func (m *mockMediaRepository) UpdateMetadata(ctx context.Context, id uuid.UUID, width, height int, duration int64, codec string) error {
	_ = ctx
	_ = id
	_ = width
	_ = height
	_ = duration
	_ = codec
	return nil
}

type mockFileStorage struct {
	saveCalled bool
	savePath   string
	saveErr    error
}

func (m *mockFileStorage) Save(ctx context.Context, filename string, src io.Reader) (string, error) {
	_ = ctx
	_ = filename
	_ = src
	m.saveCalled = true
	return m.savePath, m.saveErr
}

func (m *mockFileStorage) Delete(ctx context.Context, storagePath string) error {
	_ = ctx
	_ = storagePath
	return nil
}

// --- Tests ---

func TestUploadUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		contentType   string
		storagePath   string
		storageErr    error
		repoErr       error
		wantErr       bool
		wantRepoSaved bool
	}{
		{
			name:          "success",
			contentType:   "image/png",
			storagePath:   "uploads/test.png",
			storageErr:    nil,
			repoErr:       nil,
			wantErr:       false,
			wantRepoSaved: true,
		},
		{
			name:          "storage error",
			contentType:   "image/png",
			storagePath:   "",
			storageErr:    errors.New("disk full"),
			repoErr:       nil,
			wantErr:       true,
			wantRepoSaved: false,
		},
		{
			name:          "invalid content type",
			contentType:   "text/plain",
			storagePath:   "",
			storageErr:    nil,
			repoErr:       nil,
			wantErr:       true,
			wantRepoSaved: false,
		},
		{
			name:          "repository error",
			contentType:   "image/jpeg",
			storagePath:   "uploads/test.jpg",
			storageErr:    nil,
			repoErr:       errors.New("db error"),
			wantErr:       true,
			wantRepoSaved: true, // repo.Save всё равно вызывается, хоть и возвращает ошибку
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMediaRepository{saveErr: tt.repoErr}
			storage := &mockFileStorage{savePath: tt.storagePath, saveErr: tt.storageErr}

			uc := NewUploadUseCase(repo, storage)
			asset, err := uc.Execute(context.Background(), "test.bin", tt.contentType, 1024, bytes.NewReader([]byte("data")))

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

			// Если репозиторий не вызывался, но тип контента был валидным,
			// значит сбой произошел именно в хранилище, и оно должно было быть вызвано.
			if !tt.wantRepoSaved && tt.contentType != "text/plain" && !storage.saveCalled {
				t.Error("storage.Save should have been called")
			}
		})
	}
}
