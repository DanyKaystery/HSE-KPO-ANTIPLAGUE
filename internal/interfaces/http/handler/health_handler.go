package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	httpdto "github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/interfaces/http/dto"
)

type HealthHandler struct {
	db *sqlx.DB
}

func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health godoc
// @Summary      System health check
// @Description  Check if the system and its dependencies are healthy
// @Tags         system
// @Produce      json
// @Success      200 {object} httpdto.APIResponse{data=httpdto.HealthCheckResponse}
// @Router       /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	dbStatus := "ok"
	if err := h.db.Ping(); err != nil {
		dbStatus = "error"
	}

	health := httpdto.HealthCheckResponse{
		Status:   "ok",
		Database: dbStatus,
		Storage:  "ok",
	}

	resp := httpdto.NewSuccessResponse(health)
	c.JSON(http.StatusOK, resp)
}
