// internal/infrastructure/delivery/http/handlers/media.go

// Package handlers содержит HTTP-обработчики REST API.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/aimv/media-gallery/internal/pkg/apperror"
	"github.com/aimv/media-gallery/internal/usecase/media"
	"github.com/google/uuid"
)

// MediaHandler обрабатывает запросы к медиафайлам.
type MediaHandler struct {
	uploadUC    *media.UploadUseCase
	getStatusUC *media.GetStatusUseCase
}

// NewMediaHandler создаёт обработчик с указанными use-case'ами.
func NewMediaHandler(uploadUC *media.UploadUseCase, getStatusUC *media.GetStatusUseCase) *MediaHandler {
	return &MediaHandler{
		uploadUC:    uploadUC,
		getStatusUC: getStatusUC,
	}
}

// Upload обрабатывает загрузку нового медиафайла.
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, entity.MaxUploadSizeBytes)

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, apperror.NewAppError("invalid_input", "missing or invalid file field", http.StatusBadRequest))
		return
	}
	defer func() { _ = file.Close() }()

	contentType := header.Header.Get("Content-Type")
	size := header.Size

	asset, err := h.uploadUC.Execute(r.Context(), header.Filename, contentType, size, file)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(asset)
}

// mediaStatusResponse — DTO для ответа GET /api/media/:id.
type mediaStatusResponse struct {
	ID               string        `json:"id"`
	Status           string        `json:"status"`
	OriginalFilename string        `json:"original_filename"`
	SizeBytes        int64         `json:"size_bytes"`
	HlsPlaylistURL   string        `json:"hls_playlist_url"`
	Metadata         mediaMetadata `json:"metadata"`
}

// mediaMetadata — вложенные технические метаданные в ответе.
type mediaMetadata struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	DurationMS int64  `json:"duration_ms"`
	Codec      string `json:"codec"`
}

// GetStatus обрабатывает запрос текущего состояния медиафайла по ID.
func (h *MediaHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, apperror.NewAppError(
			apperror.ErrInvalidInput.Code,
			"invalid media id",
			apperror.ErrInvalidInput.HTTPStatus,
		))
		return
	}

	asset, err := h.getStatusUC.Execute(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	// Ссылка на HLS появляется только для полностью готовых ассетов.
	hlsURL := ""
	if asset.Status == entity.StatusReady {
		hlsURL = "/storage/hls/" + asset.ID.String() + "/master.m3u8"
	}

	resp := mediaStatusResponse{
		ID:               asset.ID.String(),
		Status:           string(asset.Status),
		OriginalFilename: asset.OriginalFilename,
		SizeBytes:        asset.SizeBytes,
		HlsPlaylistURL:   hlsURL,
		Metadata: mediaMetadata{
			Width:      asset.Width,
			Height:     asset.Height,
			DurationMS: asset.DurationMS,
			Codec:      asset.Codec,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// writeError отправляет JSON-ответ с ошибкой, мапя ошибки приложения на HTTP-статусы.
func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		w.WriteHeader(appErr.HTTPStatus)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
		})
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
}
