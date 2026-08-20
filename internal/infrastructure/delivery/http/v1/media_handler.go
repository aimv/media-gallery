package v1

import (
	"encoding/json"
	"net/http"

	"github.com/aimv/media-gallery/internal/usecase/media"
)

type MediaHandler struct {
	usecase       *media.MediaUsecase
	maxUploadSize int64
}

func NewMediaHandler(usecase *media.MediaUsecase, maxUploadSize int64) *MediaHandler {
	return &MediaHandler{
		usecase:       usecase,
		maxUploadSize: maxUploadSize,
	}
}

// Upload обрабатывает POST /api/media/upload.
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Ограничение размера тела запроса
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadSize)

	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		http.Error(w, "file too large or invalid form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	size := header.Size
	if size == 0 {
		http.Error(w, "file is empty", http.StatusBadRequest)
		return
	}
	if size > h.maxUploadSize {
		http.Error(w, "file exceeds upload limit", http.StatusRequestEntityTooLarge)
		return
	}

	input := media.UploadInput{
		File:       file,
		FileHeader: header,
		Size:       size,
	}

	asset, err := h.usecase.Upload(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]any{
		"id":     asset.ID,
		"status": asset.Status,
	})
}
