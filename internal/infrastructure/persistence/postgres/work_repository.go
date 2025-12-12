package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/shared"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/work"
)

type WorkRepository struct {
	db *sqlx.DB
}

func NewWorkRepository(db *sqlx.DB) *WorkRepository {
	return &WorkRepository{db: db}
}

type workDB struct {
	ID           uuid.UUID `db:"id"`
	AssignmentID uuid.UUID `db:"assignment_id"`
	StudentID    uuid.UUID `db:"student_id"`
	FileID       uuid.UUID `db:"file_id"`
	SubmittedAt  time.Time `db:"submitted_at"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (r *WorkRepository) toDBModel(w *work.Work) workDB {
	return workDB{
		ID:           w.ID,
		AssignmentID: w.AssignmentID,
		StudentID:    w.StudentID,
		FileID:       w.FileID,
		SubmittedAt:  w.SubmittedAt,
	}
}

func (r *WorkRepository) toDomainEntity(w workDB) *work.Work {
	return &work.Work{
		ID:           w.ID,
		AssignmentID: w.AssignmentID,
		StudentID:    w.StudentID,
		FileID:       w.FileID,
		SubmittedAt:  w.SubmittedAt,
	}
}

func (r *WorkRepository) Save(ctx context.Context, w *work.Work) error {
	model := r.toDBModel(w)

	query := `
		INSERT INTO works (id, assignment_id, student_id, file_id, submitted_at)
		VALUES (:id, :assignment_id, :student_id, :file_id, :submitted_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to save work: %w", err)
	}
	return nil
}

func (r *WorkRepository) GetByID(ctx context.Context, id uuid.UUID) (*work.Work, error) {
	var model workDB
	err := r.db.GetContext(ctx, &model, "SELECT * FROM works WHERE id = $1", id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get work: %w", err)
	}

	return r.toDomainEntity(model), nil
}

func (r *WorkRepository) FindByAssignmentID(ctx context.Context, assignmentID uuid.UUID) ([]*work.Work, error) {
	var models []workDB
	err := r.db.SelectContext(ctx, &models, "SELECT * FROM works WHERE assignment_id = $1", assignmentID)
	if err != nil {
		return nil, err
	}

	result := make([]*work.Work, len(models))
	for i, m := range models {
		result[i] = r.toDomainEntity(m)
	}
	return result, nil
}

func (r *WorkRepository) Exists(ctx context.Context, studentID, assignmentID uuid.UUID) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM works WHERE student_id = $1 AND assignment_id = $2)"
	err := r.db.QueryRowContext(ctx, query, studentID, assignmentID).Scan(&exists)
	return exists, err
}
