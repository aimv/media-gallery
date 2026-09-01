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
	"github.com/aimv/media-gallery/internal/pkg/apperror"
	"github.com/aimv/media-gallery/internal/usecase/media"
	"github.com/google/uuid"
)

// --- Mock-реализации для тестов ---

type mockMediaRepository struct {
	asset   *entity.MediaAsset
	findErr error
}

func (m *mockMediaRepository) Save(_ context.Context, _ *entity.MediaAsset) error {
	return nil
}

func (m *mockMediaRepository) FindByID(_ context.Context, _ uuid.UUID) (*entity.MediaAsset, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.asset, nil
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

// --- Tests ---

func TestMediaHandler_Upload_Success(t *testing.T) {
	repo := &mockMediaRepository{}
	storage := &mockFileStorage{}
	uploadUC := media.NewUploadUseCase(repo, storage)
	getStatusUC := media.NewGetStatusUseCase(repo)

	handler := NewMediaHandler(uploadUC, getStatusUC)

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

	if _, ok := resp["id"]; !ok {
		t.Error("missing 'id' in response")
	}
	if _, ok := resp["status"]; !ok {
		t.Error("missing 'status' in response")
	}
}

func TestMediaHandler_GetStatus_Success(t *testing.T) {
	assetID := uuid.New()
	asset := &entity.MediaAsset{
		ID:               assetID,
		OriginalFilename: "video.mp4",
		MediaType:        entity.MediaTypeMP4,
		Status:           entity.StatusReady,
		SizeBytes:        1024,
		Width:            1920,
		Height:           1080,
		DurationMS:       10500,
		Codec:            "h264",
	}

	repo := &mockMediaRepository{asset: asset}
	uploadUC := media.NewUploadUseCase(repo, &mockFileStorage{})
	getStatusUC := media.NewGetStatusUseCase(repo)

	handler := NewMediaHandler(uploadUC, getStatusUC)

	req := httptest.NewRequest(http.MethodGet, "/api/media/"+assetID.String(), nil)
	req.SetPathValue("id", assetID.String())

	rec := httptest.NewRecorder()
	handler.GetStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp struct {
		ID               string `json:"id"`
		Status           string `json:"status"`
		OriginalFilename string `json:"original_filename"`
		SizeBytes        int64  `json:"size_bytes"`
		HlsPlaylistURL   string `json:"hls_playlist_url"`
		Metadata         struct {
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			DurationMS int64  `json:"duration_ms"`
			Codec      string `json:"codec"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if resp.ID != assetID.String() {
		t.Errorf("id = %q, want %q", resp.ID, assetID.String())
	}
	if resp.Status != string(entity.StatusReady) {
		t.Errorf("status = %q, want %q", resp.Status, entity.StatusReady)
	}
	if resp.OriginalFilename != "video.mp4" {
		t.Errorf("original_filename = %q, want %q", resp.OriginalFilename, "video.mp4")
	}
	if resp.SizeBytes != 1024 {
		t.Errorf("size_bytes = %d, want 1024", resp.SizeBytes)
	}

	wantURL := "/storage/hls/" + assetID.String() + "/master.m3u8"
	if resp.HlsPlaylistURL != wantURL {
		t.Errorf("hls_playlist_url = %q, want %q", resp.HlsPlaylistURL, wantURL)
	}

	if resp.Metadata.Width != 1920 {
		t.Errorf("metadata.width = %d, want 1920", resp.Metadata.Width)
	}
	if resp.Metadata.Height != 1080 {
		t.Errorf("metadata.height = %d, want 1080", resp.Metadata.Height)
	}
	if resp.Metadata.DurationMS != 10500 {
		t.Errorf("metadata.duration_ms = %d, want 10500", resp.Metadata.DurationMS)
	}
	if resp.Metadata.Codec != "h264" {
		t.Errorf("metadata.codec = %q, want h264", resp.Metadata.Codec)
	}
}

func TestMediaHandler_GetStatus_NotFound(t *testing.T) {
	repo := &mockMediaRepository{findErr: apperror.ErrNotFound}
	uploadUC := media.NewUploadUseCase(repo, &mockFileStorage{})
	getStatusUC := media.NewGetStatusUseCase(repo)

	handler := NewMediaHandler(uploadUC, getStatusUC)

	assetID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/media/"+assetID.String(), nil)
	req.SetPathValue("id", assetID.String())

	rec := httptest.NewRecorder()
	handler.GetStatus(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'error' object in response, got: %v", resp)
	}
	if errObj["code"] != apperror.ErrNotFound.Code {
		t.Errorf("error.code = %v, want %q", errObj["code"], apperror.ErrNotFound.Code)
	}
}
