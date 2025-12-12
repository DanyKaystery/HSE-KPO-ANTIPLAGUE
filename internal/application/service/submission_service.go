package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/google/uuid"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/application/dto"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/file"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/plagiarism"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/work"
)

type SubmissionService struct {
	workRepo      work.Repository
	fileRepo      file.Repository
	plagRepo      plagiarism.Repository
	fileStorage   file.Storage
	textExtractor file.TextExtractor
	detector      plagiarism.Detector

	threshold float64
}

func NewSubmissionService(
	wr work.Repository,
	fr file.Repository,
	pr plagiarism.Repository,
	fs file.Storage,
	te file.TextExtractor,
	det plagiarism.Detector,
) *SubmissionService {
	return &SubmissionService{
		workRepo:      wr,
		fileRepo:      fr,
		plagRepo:      pr,
		fileStorage:   fs,
		textExtractor: te,
		detector:      det,
		threshold:     0.85,
	}
}

func (s *SubmissionService) SubmitWork(
	ctx context.Context,
	req dto.SubmitWorkRequest,
	fileContent io.Reader,
	fileName string,
	fileSize int64,
	mimeType string,
) (*dto.SubmitWorkResponse, error) {

	assignmentID := uuid.MustParse(req.AssignmentID)
	studentID := uuid.MustParse(req.StudentID)

	contentBytes, err := io.ReadAll(fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	hash := sha256.Sum256(contentBytes)
	hashString := hex.EncodeToString(hash[:])

	fileID := uuid.New()
	storagePath, err := s.fileStorage.Upload(ctx, fileID, bytes.NewReader(contentBytes))
	if err != nil {
		return nil, fmt.Errorf("storage upload failed: %w", err)
	}

	fileEntity := file.NewFile(fileName, storagePath, mimeType, hashString, fileSize)
	fileEntity.ID = fileID

	if err := s.fileRepo.Save(ctx, fileEntity); err != nil {
		if delErr := s.fileStorage.Delete(ctx, storagePath); delErr != nil {
			fmt.Printf("Failed to cleanup storage after DB error: %v\n", delErr)
		}
		return nil, fmt.Errorf("file metadata save failed: %w", err)
	}

	workEntity := work.NewWork(assignmentID, studentID, fileID)
	if err := s.workRepo.Save(ctx, workEntity); err != nil {
		return nil, fmt.Errorf("work save failed: %w", err)
	}

	currentText, err := s.textExtractor.ExtractText(bytes.NewReader(contentBytes), mimeType)
	if err != nil {
		fmt.Printf("Error extracting text: %v\n", err)
		currentText = ""
	}

	otherWorks, err := s.workRepo.FindByAssignmentID(ctx, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch previous works: %w", err)
	}

	maxScore := 0.0
	var matchID *uuid.UUID

	for _, w := range otherWorks {
		if w.ID == workEntity.ID {
			continue
		}

		otherFile, err := s.fileRepo.GetByID(ctx, w.FileID)
		if err != nil {
			continue
		}

		rc, err := s.fileStorage.Download(ctx, otherFile.StoragePath)
		if err != nil {
			continue
		}

		otherText, err := s.textExtractor.ExtractText(rc, otherFile.MimeType)
		_ = rc.Close()
		if err != nil {
			continue
		}

		score, err := s.detector.Compare(currentText, otherText)
		if err != nil {
			continue
		}

		if score > maxScore {
			maxScore = score
			id := w.ID
			matchID = &id
		}
	}

	report := plagiarism.NewReport(workEntity.ID, maxScore, s.threshold)
	if matchID != nil {
		report.SetMatch(*matchID, plagiarism.AnalysisDetails{
			AlgorithmUsed: "shingle",
			TotalTokens:   len(currentText),
		})
	}

	if err := s.plagRepo.Save(ctx, report); err != nil {
		return nil, fmt.Errorf("report save failed: %w", err)
	}

	return &dto.SubmitWorkResponse{
		WorkID:      workEntity.ID,
		SubmittedAt: workEntity.SubmittedAt,
		Plagiarism: dto.PlagiarismInfo{
			IsPlagiarized: report.IsPlagiarized,
			Score:         report.Score,
			Status:        "checked",
		},
	}, nil
}
