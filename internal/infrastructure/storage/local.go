// Package storage предоставляет реализацию файлового хранилища на локальном диске.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aimv/media-gallery/internal/domain/entity"
	"github.com/google/uuid"
)

// LocalStorage реализует repository.FileStorage для локальной файловой системы.
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage создаёт хранилище, сохраняющее файлы в baseDir.
func NewLocalStorage(baseDir string) *LocalStorage {
	return &LocalStorage{baseDir: baseDir}
}

// Save сохраняет поток данных в файл с уникальным именем внутри поддиректории uploads.
func (s *LocalStorage) Save(ctx context.Context, filename string, src io.Reader) (string, error) {
	_ = ctx

	mediaType, dataReader, err := entity.DetectContentType(src)
	if err != nil {
		return "", err
	}

	ext := filepath.Ext(filename)
	if ext == "" {
		switch mediaType {
		case entity.MediaTypeJPEG:
			ext = ".jpg"
		case entity.MediaTypePNG:
			ext = ".png"
		case entity.MediaTypeMP4:
			ext = ".mp4"
		}
	}

	newName := uuid.NewString() + ext
	relPath := filepath.Clean(filepath.Join("uploads", newName))
	fullPath := filepath.Clean(filepath.Join(s.baseDir, relPath))

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, dataReader); err != nil {
		return "", fmt.Errorf("copy data: %w", err)
	}

	return relPath, nil
}

// Delete удаляет файл по относительному пути. Ошибка "не найдено" игнорируется.
func (s *LocalStorage) Delete(ctx context.Context, storagePath string) error {
	_ = ctx

	fullPath := filepath.Join(s.baseDir, storagePath)
	if err := os.Remove(fullPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove file: %w", err)
	}
	return nil
}

// MoveDir атомарно перемещает директорию srcDir в destDir.
// Пути трактуются как есть (физические пути файловой системы).
// Если destDir существует, он будет удалён перед перемещением.
func (s *LocalStorage) MoveDir(ctx context.Context, srcDir, destDir string) error {
	_ = ctx

	// Гарантируем существование родительской папки назначения.
	if err := os.MkdirAll(filepath.Dir(destDir), 0o750); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	// Удаляем существующую целевую папку (например, от предыдущей неудачной попытки).
	if err := os.RemoveAll(destDir); err != nil {
		return fmt.Errorf("remove existing dest dir: %w", err)
	}

	// Атомарное перемещение в пределах одной файловой системы.
	if err := os.Rename(srcDir, destDir); err != nil {
		return fmt.Errorf("rename dir: %w", err)
	}

	return nil
}
