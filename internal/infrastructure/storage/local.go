// internal/infrastructure/storage/local.go

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
// Возвращает относительный путь к сохранённому файлу.
func (s *LocalStorage) Save(ctx context.Context, filename string, src io.Reader) (string, error) {
	_ = ctx // Явное глушение неиспользуемого контекста для линтера revive

	// Определяем тип контента и получаем восстановленный поток (с заголовком).
	mediaType, dataReader, err := entity.DetectContentType(src)
	if err != nil {
		return "", err
	}

	// Определяем расширение из исходного имени файла.
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

	// Генерируем уникальное имя.
	newName := uuid.NewString() + ext

	// Формируем относительный путь: uploads/<unique_name> и очищаем от уязвимостей.
	relPath := filepath.Clean(filepath.Join("uploads", newName))

	// Полный путь с защитой от уязвимости Path Traversal.
	fullPath := filepath.Clean(filepath.Join(s.baseDir, relPath))

	// Создаём директорию с безопасными правами 0750.
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	// Создаём файл.
	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}

	// Безопасное закрытие файла через отложенную функцию.
	defer func() { _ = f.Close() }()

	// Копируем данные из восстановленного потока.
	if _, err := io.Copy(f, dataReader); err != nil {
		return "", fmt.Errorf("copy data: %w", err)
	}

	return relPath, nil
}

// Delete удаляет файл по относительному пути. Ошибка "не найдено" игнорируется.
func (s *LocalStorage) Delete(ctx context.Context, storagePath string) error {
	_ = ctx // Явное глушение неиспользуемого контекста для линтера revive

	fullPath := filepath.Join(s.baseDir, storagePath)
	if err := os.Remove(fullPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove file: %w", err)
	}
	return nil
}
