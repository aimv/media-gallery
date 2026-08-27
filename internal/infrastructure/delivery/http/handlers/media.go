// Package handlers содержит HTTP-обработчики REST API.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aimv/media-gallery/internal/pkg/apperror"
	"github.com/aimv/media-gallery/internal/usecase/media"
)

const maxUploadSize = 500 << 20 // 500 МБ

// MediaHandler обрабатывает запросы к медиафайлам.
type MediaHandler struct {
	uploadUC *media.UploadUseCase
}

// NewMediaHandler создаёт обработчик с указанным use-case.
func NewMediaHandler(uploadUC *media.UploadUseCase) *MediaHandler {
	return &MediaHandler{uploadUC: uploadUC}
}

// Upload обрабатывает загрузку нового медиафайла.
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

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

	// Неизвестная ошибка — внутренняя.
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
}
