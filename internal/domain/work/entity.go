package work

import (
	"time"

	"github.com/google/uuid"
)

type Work struct {
	ID           uuid.UUID
	AssignmentID uuid.UUID
	StudentID    uuid.UUID
	FileID       uuid.UUID
	SubmittedAt  time.Time
}

func NewWork(assignmentID, studentID, fileID uuid.UUID) *Work {
	return &Work{
		ID:           uuid.New(),
		AssignmentID: assignmentID,
		StudentID:    studentID,
		FileID:       fileID,
		SubmittedAt:  time.Now(),
	}
}
