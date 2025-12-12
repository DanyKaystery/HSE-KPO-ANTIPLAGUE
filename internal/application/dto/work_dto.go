package dto

import (
	"time"

	"github.com/google/uuid"
)

type SubmitWorkRequest struct {
	AssignmentID string `json:"assignment_id" binding:"required,uuid"`
	StudentID    string `json:"student_id" binding:"required,uuid"`
}

type SubmitWorkResponse struct {
	WorkID      uuid.UUID      `json:"work_id"`
	SubmittedAt time.Time      `json:"submitted_at"`
	Plagiarism  PlagiarismInfo `json:"plagiarism_check"`
}

type PlagiarismInfo struct {
	IsPlagiarized bool    `json:"is_plagiarized"`
	Score         float64 `json:"score"`
	Status        string  `json:"status"`
}
