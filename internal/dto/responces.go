package dto

import "time"

type SubmitWorkResponse struct {
	Success   bool             `json:"success" example:"true"`
	Data      WorkResponseData `json:"data"`
	Timestamp time.Time        `json:"timestamp" example:"2025-12-12T19:00:00Z"`
}

type WorkResponseData struct {
	WorkID          string         `json:"work_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SubmittedAt     time.Time      `json:"submitted_at"`
	PlagiarismCheck PlagiarismInfo `json:"plagiarism_check"`
}

type PlagiarismInfo struct {
	IsPlagiarized bool    `json:"is_plagiarized" example:"true"`
	Score         float64 `json:"score" example:"0.95"`
	Status        string  `json:"status" example:"checked"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"Invalid input"`
}

type WordCloudResponse struct {
	WorkID       string `json:"work_id" example:"550e..."`
	WordCloudURL string `json:"word_cloud_url" example:"https://quickchart.io/..."`
}
