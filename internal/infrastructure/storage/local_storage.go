package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) *LocalStorage {
	return &LocalStorage{root: root}
}

func (s *LocalStorage) Save(ctx context.Context, src io.Reader, dstPath string) error {
	fullPath := filepath.Join(s.root, dstPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, src); err != nil {
		return fmt.Errorf("copy data: %w", err)
	}
	return nil
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.root, path)
	if err := os.RemoveAll(fullPath); err != nil {
		return fmt.Errorf("delete path: %w", err)
	}
	return nil
}

func (s *LocalStorage) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.root, path)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return file, nil
}

func (s *LocalStorage) Move(ctx context.Context, src, dst string) error {
	srcPath := filepath.Join(s.root, src)
	dstPath := filepath.Join(s.root, dst)
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("create dst dir: %w", err)
	}
	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("move: %w", err)
	}
	return nil
}
