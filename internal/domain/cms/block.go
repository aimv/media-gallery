package cms

import "time"

type ContentBlock struct {
	ID        string         `json:"id"`
	PageKey   string         `json:"page_key"`
	BlockType string         `json:"block_type"`
	Data      map[string]any `json:"data"`
	SortOrder int            `json:"sort_order"`
	MediaIDs  []string       `json:"media_ids,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type BlockWithMedia struct {
	ContentBlock
	Media []MediaInfo `json:"media,omitempty"`
}

type MediaInfo struct {
	ID               string `json:"id"`
	OriginalFileName string `json:"original_file_name"`
	MediaType        string `json:"media_type"`
	Status           string `json:"status"`
}
