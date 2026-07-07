package handlers

import (
	"net/http"

	"github.com/aistudio/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// PluginHandler handles plugin operations via the service layer.
type PluginHandler struct {
	svc *service.PluginService
}

// NewPluginHandler creates a new PluginHandler.
func NewPluginHandler(svc *service.PluginService) *PluginHandler {
	return &PluginHandler{svc: svc}
}

// List returns all plugins.
func (h *PluginHandler) List(c *gin.Context) {
	plugins := h.svc.List()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": plugins})
}

// Get returns a single plugin by name.
func (h *PluginHandler) Get(c *gin.Context) {
	p, err := h.svc.Get(c.Param("name"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": p})
}

// Install installs a new plugin.
func (h *PluginHandler) Install(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	if err := h.svc.Install(req.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "plugin installed"})
}

// UpdateStatus enables or disables a plugin.
func (h *PluginHandler) UpdateStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	if err := h.svc.UpdateStatus(c.Param("name"), req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "status updated"})
}

// Uninstall removes a plugin.
func (h *PluginHandler) Uninstall(c *gin.Context) {
	if err := h.svc.Uninstall(c.Param("name")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "plugin uninstalled"})
}

// Execute runs a plugin with input.
func (h *PluginHandler) Execute(c *gin.Context) {
	var req struct {
		Input map[string]interface{} `json:"input"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	result, err := h.svc.Execute(c.Request.Context(), c.Param("name"), req.Input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": result})
}