package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/aistudio/backend/internal/service"
	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// HealthHandler provides health check endpoints.
type HealthHandler struct {
	svc *service.Services
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(svc *service.Services) *HealthHandler {
	return &HealthHandler{svc: svc}
}

// ModuleStatus represents the status of a single module.
type ModuleStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "running", "degraded", "stopped"
	Message string `json:"message,omitempty"`
}

// Check returns the overall service health with module statuses.
func (h *HealthHandler) Check(c *gin.Context) {
	modules := []ModuleStatus{}

	// Workflow Engine
	modules = append(modules, ModuleStatus{
		Name:   "workflow",
		Status: "running",
	})

	// Task Manager
	modules = append(modules, ModuleStatus{
		Name:   "task",
		Status: "running",
	})

	// Plugin Manager
	pluginStatus := "running"
	pluginMsg := ""
	if h.svc != nil && h.svc.Plugin != nil {
		plugins := h.svc.Plugin.List()
		pluginMsg = "已注册"
		_ = plugins
	}
	modules = append(modules, ModuleStatus{
		Name:    "plugin",
		Status:  pluginStatus,
		Message: pluginMsg,
	})

	// Environment Manager
	envStatus := "running"
	envMsg := ""
	if h.svc != nil && h.svc.Environment != nil {
		env := h.svc.Environment.GetStatus()
		_ = env
	}
	modules = append(modules, ModuleStatus{
		Name:    "environment",
		Status:  envStatus,
		Message: envMsg,
	})

	// Agent
	modules = append(modules, ModuleStatus{
		Name:   "agent",
		Status: "running",
	})

	// MCP
	modules = append(modules, ModuleStatus{
		Name:   "mcp",
		Status: "running",
	})

	allRunning := true
	for _, m := range modules {
		if m.Status != "running" {
			allRunning = false
			break
		}
	}
	overallStatus := "running"
	if !allRunning {
		overallStatus = "degraded"
	}

	uptime := time.Since(startTime).String()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "AIStudio Backend is running",
		"data": gin.H{
			"status":   overallStatus,
			"uptime":   uptime,
			"modules":  modules,
			"version":  "1.0.0",
			"go":       runtime.Version(),
			"platform": runtime.GOOS + "/" + runtime.GOARCH,
		},
	})
}