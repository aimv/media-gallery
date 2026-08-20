package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aimv/media-gallery/internal/domain/cms"
	"github.com/google/uuid"
)

type CMSHandler struct {
	repo cms.ContentBlockRepository
}

func NewCMSHandler(repo cms.ContentBlockRepository) *CMSHandler {
	return &CMSHandler{repo: repo}
}

// SaveBlock обрабатывает POST /api/cms/blocks.
// Принимает JSON с полями: id (опционально), page_key, block_type, data, sort_order, media_ids.
func (h *CMSHandler) SaveBlock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID        string         `json:"id"`
		PageKey   string         `json:"page_key"`
		BlockType string         `json:"block_type"`
		Data      map[string]any `json:"data"`
		SortOrder int            `json:"sort_order"`
		MediaIDs  []string       `json:"media_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	block := &cms.ContentBlock{
		ID:        req.ID,
		PageKey:   req.PageKey,
		BlockType: req.BlockType,
		Data:      req.Data,
		SortOrder: req.SortOrder,
		MediaIDs:  req.MediaIDs,
	}
	if block.ID == "" {
		block.ID = uuid.New().String()
	}

	if err := h.repo.Save(r.Context(), block); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(block)
}

// GetBlock обрабатывает GET /api/cms/blocks/{id}.
func (h *CMSHandler) GetBlock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing block id", http.StatusBadRequest)
		return
	}

	block, err := h.repo.GetByIDWithMedia(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "block not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}
