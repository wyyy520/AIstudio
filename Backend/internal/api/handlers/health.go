package handlers

import (
	"net/http"

	"github.com/aistudio/backend/internal/database"
	"github.com/gin-gonic/gin"
)

// HealthHandler provides health check endpoints.
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check returns the overall service health.
func (h *HealthHandler) Check(c *gin.Context) {
	dbReady := database.IsReady()

	status := "ok"
	if !dbReady {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "AIStudio Backend is running",
		"data": gin.H{
			"status":   status,
			"database": dbReady,
		},
	})
}