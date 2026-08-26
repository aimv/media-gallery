package postgres

import (
	"testing"

	"github.com/aimv/media-gallery/internal/domain/repository"
)

func TestNewMediaRepository(t *testing.T) {
	repo := NewMediaRepository(nil)
	if repo == nil {
		t.Fatal("NewMediaRepository() returned nil")
	}
}

func TestMediaRepositoryImplementsInterface(t *testing.T) {
	_ = t
	var _ repository.MediaRepository = (*MediaRepository)(nil)
}
