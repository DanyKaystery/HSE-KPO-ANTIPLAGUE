package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type LocalFileStorage struct {
	BaseDir string
}

func NewLocalFileStorage(baseDir string) *LocalFileStorage {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create storage dir: %v", err))
	}
	return &LocalFileStorage{BaseDir: baseDir}
}

func (s *LocalFileStorage) Upload(ctx context.Context, fileID uuid.UUID, content io.Reader) (string, error) {
	filename := fileID.String()
	fullPath := filepath.Join(s.BaseDir, filename)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, content); err != nil {
		return "", fmt.Errorf("failed to save content: %w", err)
	}

	return filename, nil
}

func (s *LocalFileStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.BaseDir, path)
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found on disk")
		}
		return nil, err
	}
	return file, nil
}

func (s *LocalFileStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.BaseDir, path)
	return os.Remove(fullPath)
}
