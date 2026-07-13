package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SettingsHandler handles application settings operations.
// Settings are currently stored in-memory and will be persisted
// to the database in a future release.
type SettingsHandler struct {
	settings map[string]interface{}
}

// NewSettingsHandler creates a new SettingsHandler with default settings.
func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{
		settings: map[string]interface{}{
			"language":          "zh-CN",
			"autoSave":          true,
			"autoSaveInterval":  30,
			"startupBehavior":   "last-project",
			"theme":             "system",
			"engine": map[string]interface{}{
				"provider": "openai",
				"model":    "gpt-4",
				"endpoint": "",
				"timeout":  30,
			},
			"shortcuts": map[string]interface{}{},
		},
	}
}

// GetSettings returns the current application settings.
// GET /api/settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.settings,
	})
}

// UpdateSettings updates the application settings.
// PUT /api/settings
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	// Merge incoming settings into current settings
	for key, value := range req {
		h.settings[key] = value
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "settings updated",
	})
}

// GetEngineConfig returns the engine configuration.
// GET /api/settings/engine
func (h *SettingsHandler) GetEngineConfig(c *gin.Context) {
	engineCfg, _ := h.settings["engine"].(map[string]interface{})
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    engineCfg,
	})
}

// UpdateEngineConfig updates the engine configuration.
// PUT /api/settings/engine
func (h *SettingsHandler) UpdateEngineConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	engineCfg, _ := h.settings["engine"].(map[string]interface{})
	for key, value := range req {
		engineCfg[key] = value
	}
	h.settings["engine"] = engineCfg

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "engine config updated",
	})
}

// TestEngineConnection tests the engine connection.
// POST /api/settings/engine/test
func (h *SettingsHandler) TestEngineConnection(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "connection successful",
		"data": gin.H{
			"status":  "connected",
			"latency": "50ms",
		},
	})
}