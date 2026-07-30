package handler

import (
	"go-gin-layered-architecture/internal/helper"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// GetHealth godoc
//
// @Summary Health Check
// @Description Check API health status
// @Tags Health
// @Produce json
// @Success 200 {object} helper.MessageResponse "OK"
// @Router /api/health [get]
func (h *HealthHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, helper.MessageResponse{
		Message: "OK",
	})
}
