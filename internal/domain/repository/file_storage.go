package repository

import (
	"context"
	"io"
)

// FileStorage описывает контракт для работы с бинарными файлами.
type FileStorage interface {
	// Save сохраняет поток данных из src в файл с заданным именем filename
	// и возвращает относительный путь к сохранённому файлу.
	Save(ctx context.Context, filename string, src io.Reader) (string, error)

	// Delete удаляет файл по указанному относительному пути storagePath.
	Delete(ctx context.Context, storagePath string) error

	// MoveDir атомарно перемещает директорию srcDir в destDir.
	// Если destDir уже существует, он будет удалён перед перемещением.
	// Гарантирует, что клиенты увидят либо старое состояние destDir,
	// либо полностью готовое новое содержимое, но не частично скопированные файлы.
	MoveDir(ctx context.Context, srcDir, destDir string) error
}
