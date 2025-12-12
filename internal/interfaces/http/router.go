package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/application/service"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/interfaces/http/handler"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/interfaces/http/middleware"
)

func SetupRoutes(
	engine *gin.Engine,
	db *sqlx.DB,
	submissionSvc *service.SubmissionService,
	reportSvc *service.ReportService,
	maxFileSize int64,
) {
	engine.Use(middleware.Logger())
	engine.Use(middleware.CORS())

	healthHandler := handler.NewHealthHandler(db)
	engine.GET("/health", healthHandler.Health)

	v1 := engine.Group("/api/v1")
	{
		workHandler := handler.NewWorkHandler(submissionSvc, maxFileSize)
		v1.POST("/works", workHandler.SubmitWork)
		reportHandler := handler.NewReportHandler(reportSvc)
		v1.GET("/works/:work_id/reports", reportHandler.GetReport)
		v1.GET("/reports", reportHandler.GetAssignmentReports)
	}

}
