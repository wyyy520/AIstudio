package handlers

import (
	"net/http"

	"github.com/aistudio/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// AgentHandler handles AI agent chat and task delegation.
type AgentHandler struct {
	svc *service.AgentService
}

// NewAgentHandler creates a new AgentHandler.
func NewAgentHandler(svc *service.AgentService) *AgentHandler {
	return &AgentHandler{svc: svc}
}

// Chat processes an agent chat message and returns a response.
// POST /api/agent/chat
func (h *AgentHandler) Chat(c *gin.Context) {
	var req service.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	resp, err := h.svc.Chat(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}

// PlanOnly analyzes a message and returns a plan without executing.
// POST /api/agent/plan
func (h *AgentHandler) PlanOnly(c *gin.Context) {
	var req service.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	plan, err := h.svc.PlanOnly(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": plan})
}