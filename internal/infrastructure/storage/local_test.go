package storage

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aimv/media-gallery/internal/pkg/apperror"
)

func TestLocalStorage_Save_Success(t *testing.T) {
	baseDir := t.TempDir()
	store := NewLocalStorage(baseDir)

	// Минимальный валидный PNG (сигнатура + мусор).
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00}

	ctx := context.Background()
	path, err := store.Save(ctx, "test.png", bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	if !strings.HasSuffix(path, ".png") {
		t.Errorf("expected path to end with .png, got %q", path)
	}

	fullPath := filepath.Join(baseDir, path)
	if _, err := os.Stat(fullPath); err != nil {
		t.Errorf("saved file not found at %s: %v", fullPath, err)
	}
}

func TestLocalStorage_Save_InvalidType(t *testing.T) {
	baseDir := t.TempDir()
	store := NewLocalStorage(baseDir)

	// Текстовые данные — недопустимый тип.
	textData := []byte("hello world, this is not an image")

	ctx := context.Background()
	_, err := store.Save(ctx, "test.txt", bytes.NewReader(textData))
	if err == nil {
		t.Fatal("Save() expected error for invalid content type, got nil")
	}

	var appErr *apperror.AppError
	ok := errors.As(err, &appErr)
	if !ok {
		t.Fatalf("expected *apperror.AppError, got %T", err)
	}
	if appErr.Code != apperror.ErrInvalidInput.Code {
		t.Errorf("expected code %q, got %q", apperror.ErrInvalidInput.Code, appErr.Code)
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	baseDir := t.TempDir()
	store := NewLocalStorage(baseDir)

	// Создаём временный файл вручную с безопасными правами 0750.
	relPath := filepath.Clean(filepath.Join("uploads", "dummy.txt"))
	fullPath := filepath.Clean(filepath.Join(baseDir, relPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
		t.Fatal(err)
	}
	// Безопасные права на запись файла 0600 (только для владельца)
	if err := os.WriteFile(fullPath, []byte("temp"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := store.Delete(ctx, relPath); err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}

	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Errorf("file should be removed, but still exists")
	}

	// Повторное удаление не должно возвращать ошибку.
	if err := store.Delete(ctx, relPath); err != nil {
		t.Errorf("Delete() on non-existent file should not error, got: %v", err)
	}
}
