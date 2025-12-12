package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/plagiarism"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/work"
)

type ReportService struct {
	plagRepo plagiarism.Repository
	workRepo work.Repository
}

func NewReportService(pr plagiarism.Repository, wr work.Repository) *ReportService {
	return &ReportService{
		plagRepo: pr,
		workRepo: wr,
	}
}

type ReportResponse struct {
	WorkID          uuid.UUID                  `json:"work_id"`
	IsPlagiarized   bool                       `json:"is_plagiarized"`
	SimilarityScore float64                    `json:"similarity_score"`
	MatchedWorkID   *uuid.UUID                 `json:"matched_work_id,omitempty"`
	CreatedAt       string                     `json:"created_at"`
	Details         plagiarism.AnalysisDetails `json:"details"`
}

func (s *ReportService) GetReportByWorkID(ctx context.Context, workID uuid.UUID) (*ReportResponse, error) {
	report, err := s.plagRepo.GetByWorkID(ctx, workID)
	if err != nil {
		return nil, err
	}

	return &ReportResponse{
		WorkID:          report.WorkID,
		IsPlagiarized:   report.IsPlagiarized,
		SimilarityScore: report.Score,
		MatchedWorkID:   report.MatchedWorkID,
		CreatedAt:       report.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Details:         report.Details,
	}, nil
}

func (s *ReportService) GetReportsByAssignmentID(ctx context.Context, assignmentID uuid.UUID) ([]ReportResponse, error) {
	works, err := s.workRepo.FindByAssignmentID(ctx, assignmentID)
	if err != nil {
		return nil, err
	}

	var reports []ReportResponse
	for _, w := range works {
		report, err := s.plagRepo.GetByWorkID(ctx, w.ID)
		if err != nil {
			// Если отчета нет, пропускаем или логируем
			continue
		}

		reports = append(reports, ReportResponse{
			WorkID:          report.WorkID,
			IsPlagiarized:   report.IsPlagiarized,
			SimilarityScore: report.Score,
			MatchedWorkID:   report.MatchedWorkID,
			CreatedAt:       report.CreatedAt.Format("2006-01-02T15:04:05Z"),
			Details:         report.Details,
		})
	}

	return reports, nil
}
