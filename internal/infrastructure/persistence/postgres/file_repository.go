package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/file"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/shared"
)

type FileRepository struct {
	db *sqlx.DB
}

func NewFileRepository(db *sqlx.DB) *FileRepository {
	return &FileRepository{db: db}
}

type fileDB struct {
	ID           uuid.UUID `db:"id"`
	OriginalName string    `db:"filename"`
	StoragePath  string    `db:"storage_path"`
	Size         int64     `db:"file_size"`
	MimeType     string    `db:"mime_type"`
	Hash         string    `db:"content_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

func (r *FileRepository) Save(ctx context.Context, f *file.File) error {
	model := fileDB{
		ID:           f.ID,
		OriginalName: f.OriginalName,
		StoragePath:  f.StoragePath,
		Size:         f.Size,
		MimeType:     f.MimeType,
		Hash:         f.Hash,
		CreatedAt:    f.CreatedAt,
	}

	query := `
		INSERT INTO files (id, filename, storage_path, file_size, mime_type, content_hash, created_at)
		VALUES (:id, :filename, :storage_path, :file_size, :mime_type, :content_hash, :created_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	return err
}

func (r *FileRepository) GetByID(ctx context.Context, id uuid.UUID) (*file.File, error) {
	var model fileDB
	err := r.db.GetContext(ctx, &model, "SELECT id, filename, storage_path, file_size, mime_type, content_hash, created_at FROM files WHERE id = $1", id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	return &file.File{
		ID:           model.ID,
		OriginalName: model.OriginalName,
		StoragePath:  model.StoragePath,
		Size:         model.Size,
		MimeType:     model.MimeType,
		Hash:         model.Hash,
		CreatedAt:    model.CreatedAt,
	}, nil
}

func (r *FileRepository) GetByHash(ctx context.Context, hash string) (*file.File, error) {
	var model fileDB
	err := r.db.GetContext(ctx, &model, "SELECT id, filename, storage_path, file_size, mime_type, content_hash, created_at FROM files WHERE content_hash = $1 LIMIT 1", hash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	return &file.File{
		ID:           model.ID,
		OriginalName: model.OriginalName,
		StoragePath:  model.StoragePath,
		Size:         model.Size,
		MimeType:     model.MimeType,
		Hash:         model.Hash,
		CreatedAt:    model.CreatedAt,
	}, nil
}
