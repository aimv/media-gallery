package video

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFFmpegProcessor_ProcessToHLS_Cancel(t *testing.T) {
	// Создаём временный скрипт-заглушку, который блокируется, не порождая детей.
	script := filepath.Join(t.TempDir(), "fake_ffmpeg.sh")
	// Используем exec, чтобы процесс sleep заменил оболочку.
	content := "#!/bin/sh\nexec sleep 10\n"
	// #nosec G306 -- временный исполняемый скрипт в тестовой директории
	if err := os.WriteFile(script, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}

	processor := NewFFmpegProcessor(2)
	processor.ffmpegPath = script

	// Создаём временную директорию заранее, чтобы не вызывать t.TempDir() в горутине.
	outputDir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- processor.ProcessToHLS(ctx, "fake_input.mp4", outputDir)
	}()

	// Даём время на запуск процесса, затем отменяем контекст.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected error due to context cancellation, got nil")
		}
		// При отмене контекста exec.CommandContext убивает процесс,
		// поэтому конкретная ошибка может быть "signal: killed".
		// Достаточно убедиться, что ошибка возникла.
	case <-time.After(5 * time.Second):
		t.Fatal("ProcessToHLS did not return after cancellation")
	}
}
