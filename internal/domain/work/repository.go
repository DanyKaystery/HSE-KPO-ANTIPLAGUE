package work

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, work *Work) error
	GetByID(ctx context.Context, id uuid.UUID) (*Work, error)
	FindByAssignmentID(ctx context.Context, assignmentID uuid.UUID) ([]*Work, error)
	Exists(ctx context.Context, studentID, assignmentID uuid.UUID) (bool, error)
}
