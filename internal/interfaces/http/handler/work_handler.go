package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/application/dto"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/application/service"
	httpdto "github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/interfaces/http/dto"
)

type WorkHandler struct {
	submissionService *service.SubmissionService
	maxFileSize       int64 // 50MB
}

func NewWorkHandler(ss *service.SubmissionService, maxFileSize int64) *WorkHandler {
	return &WorkHandler{
		submissionService: ss,
		maxFileSize:       maxFileSize,
	}
}

// SubmitWork godoc
// @Summary      Submit work for plagiarism check
// @Description  Upload a work file and get instant plagiarism report
// @Tags         works
// @Accept       multipart/form-data
// @Produce      json
// @Param        assignment_id formData string true "Assignment ID (UUID)"
// @Param        student_id formData string true "Student ID (UUID)"
// @Param        file formData file true "Work file (TXT or MD)"
// @Success      202 {object} httpdto.APIResponse{data=dto.SubmitWorkResponse}
// @Failure      400 {object} httpdto.APIResponse
// @Failure      413 {object} httpdto.APIResponse
// @Router       /api/v1/works [post]
func (h *WorkHandler) SubmitWork(c *gin.Context) {
	// 1. Валидируем query параметры через Gin Bind
	var req httpdto.SubmitWorkRequest
	if err := c.ShouldBind(&req); err != nil {
		resp := httpdto.NewErrorResponse("VALIDATION_ERROR", "Invalid request parameters", err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 2. Получаем FileHeader (не сам файл пока)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		resp := httpdto.NewErrorResponse("VALIDATION_ERROR", "Missing file or invalid multipart form", err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 3. Проверяем размер файла
	if fileHeader.Size > h.maxFileSize {
		resp := httpdto.NewErrorResponse(
			"FILE_TOO_LARGE",
			fmt.Sprintf("File exceeds max size of %d bytes", h.maxFileSize),
			"",
		)
		c.JSON(http.StatusRequestEntityTooLarge, resp)
		return
	}

	// 4. Проверяем MIME type
	// Примечание: Content-Type приходит от клиента и может быть подделан,
	// но для базовой валидации достаточно. В идеале использовать http.DetectContentType
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType != "text/plain" && mimeType != "text/markdown" &&
		mimeType != "application/octet-stream" { // fallback для .md файлов
		resp := httpdto.NewErrorResponse(
			"UNSUPPORTED_FORMAT",
			"Only TXT and MD files are supported",
			fmt.Sprintf("Got: %s", mimeType),
		)
		c.JSON(http.StatusUnsupportedMediaType, resp)
		return
	}

	// 5. Открываем файл для чтения (Multipart.File реализует io.Reader)
	file, err := fileHeader.Open()
	if err != nil {
		resp := httpdto.NewErrorResponse("INTERNAL_ERROR", "Failed to open uploaded file", err.Error())
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	defer file.Close()

	// 6. Вызываем сервис
	serviceReq := dto.SubmitWorkRequest{
		AssignmentID: req.AssignmentID,
		StudentID:    req.StudentID,
	}

	result, err := h.submissionService.SubmitWork(
		c.Request.Context(),
		serviceReq,
		file,                // io.Reader
		fileHeader.Filename, // string
		fileHeader.Size,     // int64
		mimeType,            // string
	)

	if err != nil {
		// Логируем
		fmt.Printf("Submission error: %v\n", err)
		resp := httpdto.NewErrorResponse("INTERNAL_ERROR", "Failed to process submission", err.Error())
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	// 7. Возвращаем успех
	resp := httpdto.NewSuccessResponse(result)
	c.JSON(http.StatusAccepted, resp)
}
