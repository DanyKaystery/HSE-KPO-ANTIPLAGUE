package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/dto"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/docs"
)

const (
	StorageServiceURL  = "http://storage:9091"
	AnalysisServiceURL = "http://analysis:9092"
)

type AnalysisResponse struct {
	Score               float64 `json:"score"`
	SimilarityScore     float64 `json:"similarity_score"`
	IsPlagiarized       bool    `json:"is_plagiarized"`
	IsPlagiarizedCamel  bool    `json:"isPlagiarized"`
	IsPlagiarizedPascal bool    `json:"IsPlagiarized"`
}

// @title HSE KPO Antiplague Gateway API
// @version 1.0
// @description API Gateway for the Distributed Plagiarism Detection System
// @host localhost:9090
// @BasePath /
func main() {
	r := gin.Default()

	// Настройка CORS (опционально, но полезно для фронтенда)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		api.POST("/works", submitWorkHandler)
		api.GET("/works/:work_id/wordcloud", getWordCloudHandler)
	}

	log.Println("🚀 Gateway Service running on :9090")
	if err := r.Run(":9090"); err != nil {
		log.Fatal(err)
	}
}

// SubmitWork godoc
// @Summary Submit work for plagiarism check
// @Description Upload a work file, save it to storage, and trigger analysis
// @Tags works
// @Accept multipart/form-data
// @Produce json
// @Param assignment_id formData string true "Assignment ID"
// @Param student_id formData string true "Student ID"
// @Param file formData file true "Work file"
// @Success 202 {object} dto.SubmitWorkResponse "Успешная проверка"
// @Failure 400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure 503 {object} dto.ErrorResponse "Сервис недоступен"
// @Router /api/v1/works [post]
func submitWorkHandler(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Error: "File is required"})
		return
	}

	assignmentID := c.PostForm("assignment_id")
	studentID := c.PostForm("student_id")
	if assignmentID == "" || studentID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Error: "assignment_id and student_id are required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Error: "Failed to open file"})
		return
	}
	defer file.Close()

	storageResp, err := uploadToStorage(file, fileHeader.Filename, assignmentID, studentID)
	if err != nil {
		log.Printf("Storage upload failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{Success: false, Error: "Storage service unavailable"})
		return
	}

	analysisResp, err := sendToAnalysis(storageResp["file_id"].(string), assignmentID, storageResp["file_name"].(string))

	response := dto.SubmitWorkResponse{
		Success: true,
		Data: dto.WorkResponseData{
			WorkID:      storageResp["file_id"].(string),
			SubmittedAt: time.Now(),
			PlagiarismCheck: dto.PlagiarismInfo{
				Status: "pending",
				Score:  0,
			},
		},
		Timestamp: time.Now(),
	}

	if err != nil {
		log.Printf("Analysis failed: %v", err)
	} else {
		finalScore := analysisResp.Score
		if analysisResp.SimilarityScore > finalScore {
			finalScore = analysisResp.SimilarityScore
		}

		isPlag := analysisResp.IsPlagiarized || analysisResp.IsPlagiarizedCamel || analysisResp.IsPlagiarizedPascal

		response.Data.PlagiarismCheck = dto.PlagiarismInfo{
			Status:        "checked",
			Score:         finalScore,
			IsPlagiarized: isPlag,
		}
	}

	c.JSON(http.StatusAccepted, response)
}

// GenerateWordCloud godoc
// @Summary Generate Word Cloud
// @Description Get a URL to the word cloud image generated from the work text
// @Tags works
// @Produce json
// @Param work_id path string true "Work ID"
// @Success 200 {object} dto.WordCloudResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/works/{work_id}/wordcloud [get]
func getWordCloudHandler(c *gin.Context) {
	workID := c.Param("work_id")

	url := fmt.Sprintf("%s/internal/analyze/%s/wordcloud", AnalysisServiceURL, workID)
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{Success: false, Error: "Analysis service unavailable"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{Success: false, Error: "Failed to generate word cloud"})
		return
	}

	body, _ := io.ReadAll(resp.Body)
	c.Data(http.StatusOK, "application/json", body)
}

func uploadToStorage(file multipart.File, filename, assignmentID, studentID string) (map[string]interface{}, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", filename)
	io.Copy(part, file)
	writer.WriteField("assignment_id", assignmentID)
	writer.WriteField("student_id", studentID)
	writer.Close()

	resp, err := http.Post(StorageServiceURL+"/internal/files", writer.FormDataContentType(), body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("storage returned %d", resp.StatusCode)
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	return res, nil
}

func sendToAnalysis(fileID, assignmentID, filename string) (*AnalysisResponse, error) {
	reqBody := map[string]string{
		"work_id":       fileID,
		"file_id":       fileID,
		"assignment_id": assignmentID,
		"filename":      filename,
	}
	jsonValue, _ := json.Marshal(reqBody)

	resp, err := http.Post(AnalysisServiceURL+"/internal/analyze", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("analysis returned %d", resp.StatusCode)
	}

	var res AnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}
