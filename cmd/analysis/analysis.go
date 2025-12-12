package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/plagiarism"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/work"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/infrastructure/persistence/postgres"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/infrastructure/text"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/pkg/config"
)

var storageServiceURL string

func main() {
	_ = godotenv.Load(".env.local")
	cfg := config.LoadConfig()

	storageServiceURL = "http://localhost:9091"
	if cfg.Env == "docker" {
		storageServiceURL = "http://storage:9091"
	}

	dbCfg := postgres.Config{
		Host: cfg.DBHost, Port: cfg.DBPort, User: cfg.DBUser,
		Password: cfg.DBPassword, DBName: cfg.DBName, SSLMode: cfg.DBSSLMode,
	}
	db, err := postgres.NewConnection(dbCfg)
	if err != nil {
		log.Fatalf("Analysis Service: DB Connection failed: %v", err)
	}

	workRepo := postgres.NewWorkRepository(db)
	plagRepo := postgres.NewPlagiarismRepository(db)

	detector := plagiarism.NewShingleDetector()
	extractor := text.NewSimpleExtractor()

	r := gin.Default()

	r.POST("/internal/analyze", func(c *gin.Context) {
		analyzeHandler(c, db, workRepo, plagRepo, detector, extractor)
	})

	r.GET("/internal/analyze/:work_id/wordcloud", func(c *gin.Context) {
		wordCloudHandler(c, workRepo, extractor)
	})

	port := ":9092"
	log.Printf("🧠 Analysis Service running on %s", port)
	r.Run(port)
}

type AnalyzeRequest struct {
	WorkID       uuid.UUID `json:"work_id"`
	FileID       uuid.UUID `json:"file_id"`
	AssignmentID uuid.UUID `json:"assignment_id"`
	Filename     string    `json:"filename"`
}

func analyzeHandler(
	c *gin.Context,
	db *sqlx.DB,
	workRepo work.Repository,
	plagRepo plagiarism.Repository,
	detector *plagiarism.ShingleDetector,
	extractor *text.SimpleExtractor,
) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	content, err := downloadFileFromStorage(req.FileID)
	if err != nil {
		log.Printf("Failed to download file: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to download file from storage"})
		return
	}

	mimeType := "text/plain"
	currentText, _ := extractor.ExtractText(bytes.NewReader(content), mimeType)

	log.Printf("DEBUG: Extracted text length: %d", len(currentText))

	otherWorks, err := workRepo.FindByAssignmentID(c.Request.Context(), req.AssignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch other works"})
		return
	}

	maxScore := 0.0
	var matchID *uuid.UUID

	log.Printf("DEBUG: Found %d other works for assignment", len(otherWorks))

	for _, w := range otherWorks {
		if w.ID == req.WorkID {
			continue
		}

		otherContent, err := downloadFileFromStorage(w.FileID)
		if err != nil {
			continue
		}

		otherText, _ := extractor.ExtractText(bytes.NewReader(otherContent), mimeType)

		score, _ := detector.Compare(currentText, otherText)
		if score > maxScore {
			maxScore = score
			id := w.ID
			matchID = &id
		}

		log.Printf("DEBUG: Final Result - Score: %f, IsPlagiarized: %v", maxScore, maxScore > 0.85)
	}

	report := plagiarism.NewReport(req.WorkID, maxScore, 0.85)
	if matchID != nil {
		report.SetMatch(*matchID, plagiarism.AnalysisDetails{AlgorithmUsed: "shingle"})
	}
	plagRepo.Save(c.Request.Context(), report)

	c.JSON(http.StatusOK, report)
}

func wordCloudHandler(c *gin.Context, workRepo work.Repository, extractor *text.SimpleExtractor) {
	workIDStr := c.Param("work_id")
	workID, err := uuid.Parse(workIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
		return
	}

	workEntity, err := workRepo.GetByID(c.Request.Context(), workID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Work not found"})
		return
	}

	content, err := downloadFileFromStorage(workEntity.FileID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to get file content"})
		return
	}

	textStr, err := extractor.ExtractText(bytes.NewReader(content), "text/plain")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract text"})
		return
	}

	if len(textStr) > 1000 {
		textStr = textStr[:1000]
	}

	quickChartURL := fmt.Sprintf("https://quickchart.io/wordcloud?text=%s&format=png&width=800&height=600", url.QueryEscape(textStr))

	c.JSON(http.StatusOK, gin.H{
		"work_id":        workID,
		"word_cloud_url": quickChartURL,
	})
}

func downloadFileFromStorage(fileID uuid.UUID) ([]byte, error) {
	url := fmt.Sprintf("%s/internal/files/%s/content", storageServiceURL, fileID.String())

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("storage returned %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
