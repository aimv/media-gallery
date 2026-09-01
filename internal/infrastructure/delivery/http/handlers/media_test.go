package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/aimv/media-gallery/internal/usecase/media"
	"github.com/google/uuid"
)

// --- Mock-реализации для тестов ---

type mockMediaRepository struct{}

func (m *mockMediaRepository) Save(_ context.Context, _ *entity.MediaAsset) error {
	return nil
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

type mockFileStorage struct{}

func (m *mockFileStorage) Save(_ context.Context, _ string, _ io.Reader) (string, error) {
	return "uploads/test.png", nil
}
func (m *mockFileStorage) Delete(_ context.Context, _ string) error {
	return nil
}
func (m *mockFileStorage) MoveDir(_ context.Context, _, _ string) error {
	return nil
}

func TestMediaHandler_Upload_Success(t *testing.T) {
	// Подготавливаем use-case с моками.
	repo := &mockMediaRepository{}
	storage := &mockFileStorage{}
	uploadUC := media.NewUploadUseCase(repo, storage)

	handler := NewMediaHandler(uploadUC)

	// Формируем multipart-запрос.
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Создаём часть формы с явным указанием заголовков.
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="test.png"`)
	header.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}

	// Записываем сигнатуру PNG.
	if _, err := part.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}); err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	handler.Upload(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	// Проверяем наличие полей ассета.
	if _, ok := resp["id"]; !ok {
		t.Error("missing 'id' in response")
	}
	if _, ok := resp["status"]; !ok {
		t.Error("missing 'status' in response")
	}
}
