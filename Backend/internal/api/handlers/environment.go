package handlers

import (
	"net/http"

	"github.com/aistudio/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type EnvironmentHandler struct {
	svc *service.EnvironmentService
}

func NewEnvironmentHandler(svc *service.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{svc: svc}
}

func (h *EnvironmentHandler) GetStatus(c *gin.Context) {
	status := h.svc.GetStatus()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    status,
	})
}
