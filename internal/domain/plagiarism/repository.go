package plagiarism

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, report *Report) error
	GetByWorkID(ctx context.Context, workID uuid.UUID) (*Report, error)
}

type Detector interface {
	Compare(text1, text2 string) (float64, error)
}
