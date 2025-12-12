package file

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, file *File) error
	GetByID(ctx context.Context, id uuid.UUID) (*File, error)
	GetByHash(ctx context.Context, hash string) (*File, error) // Для проверки дубликатов
}

type Storage interface {
	Upload(ctx context.Context, fileID uuid.UUID, content io.Reader) (string, error)
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
}

type TextExtractor interface {
	ExtractText(content io.Reader, mimeType string) (string, error)
}
