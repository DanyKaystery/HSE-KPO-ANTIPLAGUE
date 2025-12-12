package plagiarism

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	ID            uuid.UUID
	WorkID        uuid.UUID
	IsPlagiarized bool
	Score         float64

	MatchedWorkID *uuid.UUID

	Details AnalysisDetails

	CreatedAt time.Time
}

type AnalysisDetails struct {
	AlgorithmUsed string `json:"algorithm"`
	MatchedTokens int    `json:"matched_tokens"`
	TotalTokens   int    `json:"total_tokens"`
}

func NewReport(workID uuid.UUID, score float64, threshold float64) *Report {
	return &Report{
		ID:            uuid.New(),
		WorkID:        workID,
		Score:         score,
		IsPlagiarized: score > threshold,
		CreatedAt:     time.Now(),
	}
}

func (r *Report) SetMatch(matchedWorkID uuid.UUID, details AnalysisDetails) {
	r.MatchedWorkID = &matchedWorkID
	r.Details = details
}
