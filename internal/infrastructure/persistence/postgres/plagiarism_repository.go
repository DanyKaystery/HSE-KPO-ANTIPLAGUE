package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/plagiarism"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/shared"
)

type PlagiarismRepository struct {
	db *sqlx.DB
}

func NewPlagiarismRepository(db *sqlx.DB) *PlagiarismRepository {
	return &PlagiarismRepository{db: db}
}

type reportDB struct {
	ID            uuid.UUID       `db:"id"`
	WorkID        uuid.UUID       `db:"work_id"`
	IsPlagiarized bool            `db:"is_plagiarized"`
	Score         float64         `db:"similarity_score"`
	MatchedWorkID *uuid.UUID      `db:"matched_with_work_id"`
	DetailsJSON   json.RawMessage `db:"analysis_details"`
	CreatedAt     time.Time       `db:"created_at"`
}

func (r *PlagiarismRepository) Save(ctx context.Context, report *plagiarism.Report) error {
	detailsBytes, err := json.Marshal(report.Details)
	if err != nil {
		return err
	}

	model := reportDB{
		ID:            report.ID,
		WorkID:        report.WorkID,
		IsPlagiarized: report.IsPlagiarized,
		Score:         report.Score,
		MatchedWorkID: report.MatchedWorkID,
		DetailsJSON:   detailsBytes,
		CreatedAt:     report.CreatedAt,
	}

	query := `
		INSERT INTO plagiarism_reports (
			id, work_id, is_plagiarized, similarity_score, 
			matched_with_work_id, analysis_details, created_at
		) VALUES (
			:id, :work_id, :is_plagiarized, :similarity_score, 
			:matched_with_work_id, :analysis_details, :created_at
		)
	`

	_, err = r.db.NamedExecContext(ctx, query, model)
	return err
}

func (r *PlagiarismRepository) GetByWorkID(ctx context.Context, workID uuid.UUID) (*plagiarism.Report, error) {
	var model reportDB
	query := "SELECT * FROM plagiarism_reports WHERE work_id = $1 ORDER BY created_at DESC LIMIT 1"

	err := r.db.GetContext(ctx, &model, query, workID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	var details plagiarism.AnalysisDetails
	if len(model.DetailsJSON) > 0 {
		if err := json.Unmarshal(model.DetailsJSON, &details); err != nil {
			return nil, err
		}
	}

	return &plagiarism.Report{
		ID:            model.ID,
		WorkID:        model.WorkID,
		IsPlagiarized: model.IsPlagiarized,
		Score:         model.Score,
		MatchedWorkID: model.MatchedWorkID,
		Details:       details,
		CreatedAt:     model.CreatedAt,
	}, nil
}
