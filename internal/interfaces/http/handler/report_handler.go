package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/application/service"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/shared"
	httpdto "github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/interfaces/http/dto"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(rs *service.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: rs,
	}
}

// GetReport godoc
// @Summary      Get plagiarism report for a work
// @Description  Retrieve the plagiarism check report for a specific work
// @Tags         reports
// @Produce      json
// @Param        work_id path string true "Work ID (UUID)"
// @Success      200 {object} httpdto.APIResponse
// @Failure      404 {object} httpdto.APIResponse
// @Router       /api/v1/works/{work_id}/reports [get]
func (h *ReportHandler) GetReport(c *gin.Context) {
	workIDStr := c.Param("work_id")

	workID, err := uuid.Parse(workIDStr)
	if err != nil {
		resp := httpdto.NewErrorResponse("VALIDATION_ERROR", "Invalid work_id format", "")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	report, err := h.reportService.GetReportByWorkID(c.Request.Context(), workID)
	if err != nil {
		if err == shared.ErrNotFound {
			resp := httpdto.NewErrorResponse("NOT_FOUND", "Report not found", "")
			c.JSON(http.StatusNotFound, resp)
			return
		}
		resp := httpdto.NewErrorResponse("INTERNAL_ERROR", "Failed to get report", err.Error())
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := httpdto.NewSuccessResponse(report)
	c.JSON(http.StatusOK, resp)
}

// GetAssignmentReports godoc
// @Summary      Get all plagiarism reports for an assignment
// @Description  Retrieve plagiarism reports for all works in an assignment
// @Tags         reports
// @Produce      json
// @Param        assignment_id query string true "Assignment ID (UUID)"
// @Success      200 {object} httpdto.APIResponse
// @Failure      400 {object} httpdto.APIResponse
// @Router       /api/v1/reports [get]
func (h *ReportHandler) GetAssignmentReports(c *gin.Context) {
	assignmentIDStr := c.Query("assignment_id")
	if assignmentIDStr == "" {
		resp := httpdto.NewErrorResponse("VALIDATION_ERROR", "Missing assignment_id query parameter", "")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		resp := httpdto.NewErrorResponse("VALIDATION_ERROR", "Invalid assignment_id format", "")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	reports, err := h.reportService.GetReportsByAssignmentID(c.Request.Context(), assignmentID)
	if err != nil {
		resp := httpdto.NewErrorResponse("INTERNAL_ERROR", "Failed to get reports", err.Error())
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := httpdto.NewSuccessResponse(reports)
	c.JSON(http.StatusOK, resp)
}
