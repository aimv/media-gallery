package video

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aimv/media-gallery/internal/pkg/apperror"
)

// writeMockScript создаёт временный исполняемый shell-скрипт с заданным содержимым.
func writeMockScript(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "mock.sh")
	// #nosec G306 -- временный исполняемый скрипт в тестовой директории
	if err := os.WriteFile(script, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	// #nosec G302 -- скрипту необходим бит выполнения для запуска в тесте
	if err := os.Chmod(script, 0o700); err != nil {
		t.Fatal(err)
	}
	return script
}

func TestFFmpegProcessor_ProcessToHLS_Cancel(t *testing.T) {
	// Создаём временный скрипт-заглушку, который блокируется, не порождая детей.
	script := writeMockScript(t, "#!/bin/sh\nexec sleep 10\n")

	processor := NewFFmpegProcessor(2)
	processor.ffmpegPath = script

	outputDir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- processor.ProcessToHLS(ctx, "fake_input.mp4", outputDir)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected error due to context cancellation, got nil")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ProcessToHLS did not return after cancellation")
	}
}

func TestFFmpegProcessor_ProbeMetadata_Success(t *testing.T) {
	jsonOutput := `{"streams": [{"width": 1920, "height": 1080, "duration": "10.500000", "codec_name": "h264"}]}`
	script := writeMockScript(t, "#!/bin/sh\ncat <<'EOF'\n"+jsonOutput+"\nEOF\n")

	processor := NewFFmpegProcessor(2)
	processor.ffprobePath = script

	meta, err := processor.ProbeMetadata(context.Background(), "fake_input.mp4")
	if err != nil {
		t.Fatalf("ProbeMetadata() unexpected error: %v", err)
	}
	if meta == nil {
		t.Fatal("expected metadata, got nil")
	}
	if meta.Width != 1920 {
		t.Errorf("Width = %d, want 1920", meta.Width)
	}
	if meta.Height != 1080 {
		t.Errorf("Height = %d, want 1080", meta.Height)
	}
	if meta.DurationMS != 10500 {
		t.Errorf("DurationMS = %d, want 10500", meta.DurationMS)
	}
	if meta.Codec != "h264" {
		t.Errorf("Codec = %q, want h264", meta.Codec)
	}
}

func TestFFmpegProcessor_Validate_InvalidData(t *testing.T) {
	tests := []struct {
		name       string
		jsonOutput string
	}{
		{
			name:       "empty streams array",
			jsonOutput: `{"streams": []}`,
		},
		{
			name:       "malformed json",
			jsonOutput: `{"streams": [{"width": 1920,`,
		},
		{
			name:       "incomplete stream fields",
			jsonOutput: `{"streams": [{"width": 0, "height": 0, "codec_name": ""}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := writeMockScript(t, "#!/bin/sh\ncat <<'EOF'\n"+tt.jsonOutput+"\nEOF\n")

			processor := NewFFmpegProcessor(2)
			processor.ffprobePath = script

			meta, err := processor.Validate(context.Background(), "fake_input.mp4")
			if err == nil {
				t.Fatal("Validate() expected error, got nil")
			}
			if meta != nil {
				t.Errorf("Validate() expected nil metadata, got %+v", meta)
			}

			var appErr *apperror.AppError
			if !errors.As(err, &appErr) {
				t.Fatalf("expected *apperror.AppError, got %T: %v", err, err)
			}
			if appErr.Code != apperror.ErrInvalidInput.Code {
				t.Errorf("error code = %q, want %q", appErr.Code, apperror.ErrInvalidInput.Code)
			}
		})
	}
}
