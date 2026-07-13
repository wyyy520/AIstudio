package handlers

import (
	"net/http"

	"github.com/aistudio/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type InstallRequest struct {
	ManifestURL string `json:"manifest_url" binding:"required"`
}

type ExecuteRequest struct {
	Input  map[string]interface{} `json:"input"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// PluginHandler handles plugin operations via the service layer (V2).
// Plugins are pure declarations �?no install, uninstall, or execution.
type PluginHandler struct {
	svc *service.PluginService
}

// NewPluginHandler creates a new PluginHandler.
func NewPluginHandler(svc *service.PluginService) *PluginHandler {
	return &PluginHandler{svc: svc}
}

// List returns all registered plugins as summaries.
func (h *PluginHandler) List(c *gin.Context) {
	plugins := h.svc.List()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": plugins})
}

// Get returns a single plugin by name or ID.
func (h *PluginHandler) Get(c *gin.Context) {
	name := c.Param("name")
	p, err := h.svc.Get(name)
	if err != nil {
		p, err = h.svc.GetByID(name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": p})
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

// GetPlugins handles GET /api/plugins (returns all plugins).
func (h *PluginHandler) GetPlugins(c *gin.Context) {
	plugins := h.svc.List()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": plugins})
}

// GetPluginByID handles GET /api/plugin/:id.
func (h *PluginHandler) GetPluginByID(c *gin.Context) {
	id := c.Param("id")
	p, err := h.svc.GetByID(id)
	if err != nil {
		p, err = h.svc.Get(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": p})
}

// GetNodes returns all node types from enabled plugins.
func (h *PluginHandler) GetNodes(c *gin.Context) {
	nodes := h.svc.GetNodeTypes()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nodes})
}

// Install installs a plugin from a manifest URL.
func (h *PluginHandler) Install(c *gin.Context) {
	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	task, err := h.svc.Install(c.Request.Context(), req.ManifestURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "install started", "data": gin.H{
		"task_id": task.ID,
		"status":  task.Status,
	}})
}

// Uninstall removes a plugin by name.
func (h *PluginHandler) Uninstall(c *gin.Context) {
	name := c.Param("name")
	if err := h.svc.Uninstall(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "plugin uninstalled"})
}

// InstallStatus returns the install status for a plugin.
func (h *PluginHandler) InstallStatus(c *gin.Context) {
	name := c.Param("name")
	status := h.svc.InstallStatus(name)
	if status == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "no install task found for plugin"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": status})
}

// Execute runs a plugin with the given input.
func (h *PluginHandler) Execute(c *gin.Context) {
	name := c.Param("name")
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	result, err := h.svc.Execute(c.Request.Context(), name, req.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": result})
}
