package entity

import "testing"

func TestMediaAsset_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name    string
		current MediaStatus
		next    MediaStatus
		want    bool
	}{
		// Валидные переходы
		{"uploaded -> queued", StatusUploaded, StatusQueued, true},
		{"queued -> processing", StatusQueued, StatusProcessing, true},
		{"processing -> ready", StatusProcessing, StatusReady, true},
		{"processing -> failed", StatusProcessing, StatusFailed, true},
		{"failed -> queued (retry)", StatusFailed, StatusQueued, true},
		{"ready -> deleting", StatusReady, StatusDeleting, true},
		{"failed -> deleting", StatusFailed, StatusDeleting, true},

		// Невалидные переходы
		{"uploaded -> processing", StatusUploaded, StatusProcessing, false},
		{"uploaded -> ready", StatusUploaded, StatusReady, false},
		{"queued -> ready", StatusQueued, StatusReady, false},
		{"queued -> failed", StatusQueued, StatusFailed, false},
		{"processing -> queued", StatusProcessing, StatusQueued, false},
		{"processing -> deleting", StatusProcessing, StatusDeleting, false},
		{"ready -> queued", StatusReady, StatusQueued, false},
		{"ready -> failed", StatusReady, StatusFailed, false},
		{"failed -> ready", StatusFailed, StatusReady, false},
		{"deleting -> uploaded", StatusDeleting, StatusUploaded, false},
		{"deleting -> queued", StatusDeleting, StatusQueued, false},
		{"deleting -> processing", StatusDeleting, StatusProcessing, false},
		{"deleting -> ready", StatusDeleting, StatusReady, false},
		{"deleting -> failed", StatusDeleting, StatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := &MediaAsset{Status: tt.current}
			if got := asset.CanTransitionTo(tt.next); got != tt.want {
				t.Errorf("CanTransitionTo(%q) = %v, want %v", tt.next, got, tt.want)
			}
		})
	}
}
