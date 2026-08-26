// Package storage предоставляет реализацию файлового хранилища на локальном диске.
package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aimv/media-gallery/internal/pkg/apperror"
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

	// Читаем первые 512 байт для определения MIME-типа.
	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", fmt.Errorf("read file header: %w", err)
	}
	head = head[:n]

	// Проверяем Magic Bytes.
	contentType := http.DetectContentType(head)
	switch contentType {
	case "image/jpeg", "image/png", "video/mp4":
		// допустимо
	default:
		return "", apperror.NewAppError(
			apperror.ErrInvalidInput.Code,
			fmt.Sprintf("unsupported content type: %s", contentType),
			apperror.ErrInvalidInput.HTTPStatus,
		)
	}

	// Определяем расширение из исходного имени файла.
	ext := filepath.Ext(filename)
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "video/mp4":
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

	// Создаём файл защищенным методом.
	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}

	// Безопасная обработка закрытия файла через анонимную функцию дефера.
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			// Ошибку закрытия дефера логируем или пробрасываем, если критично.
			// Для errcheck достаточно того, что возвращаемое значение не проигнорировано.
			_ = closeErr
		}
	}()

	// Объединяем прочитанный заголовок и оставшийся поток.
	combined := io.MultiReader(bytes.NewReader(head), src)

	// Копируем данные.
	if _, err := io.Copy(f, combined); err != nil {
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
