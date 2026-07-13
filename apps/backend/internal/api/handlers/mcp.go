package handlers

import (
	"net/http"

	"github.com/aistudio/backend/internal/mcp"
	"github.com/aistudio/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// MCPHandler handles MCP-related API requests.
type MCPHandler struct {
	svc *service.MCPService
}

// NewMCPHandler creates a new MCPHandler.
func NewMCPHandler(svc *service.MCPService) *MCPHandler {
	return &MCPHandler{svc: svc}
}

// ListTools returns all registered MCP tools.
// GET /api/mcp/tools
func (h *MCPHandler) ListTools(c *gin.Context) {
	serverName := c.Query("server")
	var tools []mcp.MCPTool
	if serverName != "" {
		tools = h.svc.ListToolsByServer(serverName)
	} else {
		tools = h.svc.ListTools()
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": tools})
}

// ListServers returns all registered MCP servers.
// GET /api/mcp/servers
func (h *MCPHandler) ListServers(c *gin.Context) {
	servers := h.svc.ListServers()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": servers})
}

// GetStatus returns the MCP system status.
// GET /api/mcp/status
func (h *MCPHandler) GetStatus(c *gin.Context) {
	status := h.svc.GetStatus()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": status})
}

// ConnectRequest is the request body for connecting to an MCP server.
type ConnectRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Transport   string `json:"transport"`
	Endpoint    string `json:"endpoint"`
	Enabled     bool   `json:"enabled"`
	TimeoutMs   int    `json:"timeout_ms"`
}

// Connect connects to an MCP server.
// POST /api/mcp/connect
func (h *MCPHandler) Connect(c *gin.Context) {
	var req ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	config := mcp.MCPConfig{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Transport:   mcp.TransportType(req.Transport),
		Endpoint:    req.Endpoint,
		Enabled:     req.Enabled,
		TimeoutMs:   req.TimeoutMs,
	}
	if config.Transport == "" {
		config.Transport = mcp.TransportMock
	}
	if config.Type == "" {
		config.Type = "other"
	}

	if err := h.svc.Connect(c.Request.Context(), config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	// Save config after successful connection
	h.svc.SaveConfig()

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "connected", "data": config})
}

// Disconnect disconnects from an MCP server.
// POST /api/mcp/disconnect
func (h *MCPHandler) Disconnect(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	if err := h.svc.Disconnect(req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	h.svc.SaveConfig()

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "disconnected"})
}

// CallRequest is the request body for calling an MCP tool.
type CallRequest struct {
	ServerName string                 `json:"server_name" binding:"required"`
	ToolName   string                 `json:"tool_name" binding:"required"`
	Input      map[string]interface{} `json:"input"`
	TimeoutMs  int                    `json:"timeout_ms"`
}

// Call calls an MCP tool.
// POST /api/mcp/call
func (h *MCPHandler) Call(c *gin.Context) {
	var req CallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	if req.Input == nil {
		req.Input = make(map[string]interface{})
	}

	resp, err := h.svc.Call(c.Request.Context(), mcp.CallRequest{
		ServerName: req.ServerName,
		ToolName:   req.ToolName,
		Input:      req.Input,
		TimeoutMs:  req.TimeoutMs,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusOK, gin.H{"code": -1, "message": resp.Error, "data": resp})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}

// AddServer adds a server configuration without connecting.
// POST /api/mcp/servers
func (h *MCPHandler) AddServer(c *gin.Context) {
	var req ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	config := mcp.MCPConfig{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Transport:   mcp.TransportType(req.Transport),
		Endpoint:    req.Endpoint,
		Enabled:     req.Enabled,
		TimeoutMs:   req.TimeoutMs,
	}
	if config.Transport == "" {
		config.Transport = mcp.TransportMock
	}
	if config.Type == "" {
		config.Type = "other"
	}

	if err := h.svc.AddServer(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	h.svc.SaveConfig()

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "server added", "data": config})
}

// RemoveServer removes a server configuration.
// DELETE /api/mcp/servers/:name
func (h *MCPHandler) RemoveServer(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "server name is required"})
		return
	}

	if err := h.svc.RemoveServer(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	h.svc.SaveConfig()

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "server removed"})
}

// ExportConfig returns the current MCP configuration as JSON.
// GET /api/mcp/config
func (h *MCPHandler) ExportConfig(c *gin.Context) {
	data, err := h.svc.ExportConfigJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}