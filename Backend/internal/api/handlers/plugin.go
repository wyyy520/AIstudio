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

// Get returns a single plugin by name or ID.
func (h *PluginHandler) Get(c *gin.Context) {
	name := c.Param("name")
	p, err := h.svc.Get(name)
	if err != nil {
		// Try by ID
		p, err = h.svc.GetByID(name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": p})
}

// Install installs a new plugin.
func (h *PluginHandler) Install(c *gin.Context) {
	var req struct {
		ManifestPath string `json:"manifest_path"`
		Name         string `json:"name"`
		URL          string `json:"url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	var result interface{}
	var err error

	if req.ManifestPath != "" {
		result, err = h.svc.Install(req.ManifestPath)
	} else if req.Name != "" {
		url := req.URL
		if url == "" {
			url = "https://plugins.aistudio.dev/" + req.Name
		}
		result, err = h.svc.InstallFromURL(req.Name, url)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "manifest_path or name is required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error(), "data": result})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "plugin installed", "data": result})
}

// Remove removes a plugin.
func (h *PluginHandler) Remove(c *gin.Context) {
	name := c.Param("name")
	if err := h.svc.Remove(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "plugin removed"})
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

// GetPlugins handles GET /api/plugins (returns all plugins).
func (h *PluginHandler) GetPlugins(c *gin.Context) {
	plugins := h.svc.List()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": plugins})
}

// GetPluginByID handles GET /api/plugin/:id (returns a single plugin by ID).
func (h *PluginHandler) GetPluginByID(c *gin.Context) {
	id := c.Param("id")
	p, err := h.svc.GetByID(id)
	if err != nil {
		// Try by name
		p, err = h.svc.Get(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": p})
}

// InstallPlugin handles POST /api/plugin/install.
func (h *PluginHandler) InstallPlugin(c *gin.Context) {
	var req struct {
		ManifestPath string `json:"manifest_path"`
		Name         string `json:"name" binding:"required"`
		URL          string `json:"url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	url := req.URL
	if url == "" {
		url = "https://plugins.aistudio.dev/" + req.Name
	}

	result, err := h.svc.InstallFromURL(req.Name, url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error(), "data": result})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "plugin installed", "data": result})
}

// RemovePlugin handles POST /api/plugin/remove.
func (h *PluginHandler) RemovePlugin(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	if err := h.svc.Remove(req.Name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "plugin removed"})
}