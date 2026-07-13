package handlers

import (
	"fmt"
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
	if h.svc != nil && h.svc.PluginManager != nil {
		plugins := h.svc.PluginManager.ListPlugins()
		pluginMsg = fmt.Sprintf("已注册 %d 个插件", len(plugins))
	}
	modules = append(modules, ModuleStatus{
		Name:    "plugin",
		Status:  pluginStatus,
		Message: pluginMsg,
	})

	// Environment Manager - integrate with environment status
	envStatus := "running"
	envMsg := ""
	if h.svc != nil && h.svc.EnvIntegration != nil {
		envResult := h.svc.EnvIntegration.GetManager().GetStatus()
		if envResult != nil {
			if m, ok := envResult.(map[string]interface{}); ok {
				if overall, ok := m["overall"].(string); ok {
					switch overall {
					case "ready":
						envStatus = "running"
						envMsg = "环境就绪"
					case "degraded":
						envStatus = "degraded"
						envMsg = "环境部分就绪"
					default:
						envStatus = "running"
						envMsg = "环境状态: " + overall
					}
				}
				if issues, ok := m["issues"].([]interface{}); ok && len(issues) > 0 {
					envMsg = fmt.Sprintf("环境存在 %d 个问题", len(issues))
				}
			}
		}
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
	moduleCount := len(modules)
	degradedCount := 0
	for _, m := range modules {
		if m.Status != "running" {
			allRunning = false
			degradedCount++
		}
	}
	overallStatus := "running"
	overallMsg := "All modules operational"
	if !allRunning {
		overallStatus = "degraded"
		overallMsg = fmt.Sprintf("%d/%d modules are degraded", degradedCount, moduleCount)
	}

	uptime := time.Since(startTime).String()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": overallMsg,
		"data": gin.H{
			"status":   overallStatus,
			"uptime":   uptime,
			"modules":  modules,
			"version":  "0.1.0",
			"go":       runtime.Version(),
			"platform": runtime.GOOS + "/" + runtime.GOARCH,
		},
	})
}