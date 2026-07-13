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

// GetStatus returns the current environment status (quick, cached).
func (h *EnvironmentHandler) GetStatus(c *gin.Context) {
	status := h.svc.GetStatus()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    status,
	})
}

// Check runs a full environment check and returns detailed results.
func (h *EnvironmentHandler) Check(c *gin.Context) {
	result := h.svc.Check()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

// GetRepairPlan returns a repair plan without executing it.
func (h *EnvironmentHandler) GetRepairPlan(c *gin.Context) {
	plan := h.svc.GetRepairPlan()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    plan,
	})
}

// Repair executes the repair plan to fix environment issues.
func (h *EnvironmentHandler) Repair(c *gin.Context) {
	result := h.svc.Repair()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

// InstallDependency installs a single dependency.
type installRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *EnvironmentHandler) InstallDependency(c *gin.Context) {
	var req installRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Missing required field: name",
		})
		return
	}

	if err := h.svc.InstallDependency(req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"name":    req.Name,
			"status":  "installed",
		},
	})
}

// GetLogs returns all environment operation logs.
func (h *EnvironmentHandler) GetLogs(c *gin.Context) {
	logs := h.svc.GetLogs()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    logs,
	})
}

// ClearLogs clears the environment operation logs.
func (h *EnvironmentHandler) ClearLogs(c *gin.Context) {
	h.svc.ClearLogs()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}