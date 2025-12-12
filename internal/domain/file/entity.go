package file

import (
	"github.com/google/uuid"
	"time"
)

type File struct {
	ID           uuid.UUID
	OriginalName string
	StoragePath  string
	Size         int64
	MimeType     string
	Hash         string
	CreatedAt    time.Time
}

func NewFile(name, storagePath, mimeType, hash string, size int64) *File {
	return &File{
		ID:           uuid.New(),
		OriginalName: name,
		StoragePath:  storagePath,
		Size:         size,
		MimeType:     mimeType,
		Hash:         hash,
		CreatedAt:    time.Now(),
	}
}
