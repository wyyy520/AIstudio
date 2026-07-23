package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aistudio/backend/internal/agent"
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

// Chat processes an agent chat message and streams the response via SSE.
// POST /api/agent/chat
func (h *AgentHandler) Chat(c *gin.Context) {
	// Step 1: Set SSE headers (must happen before any writes to the body)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Step 2: Parse request
	var req service.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeSSEEvent(c.Writer, "error", fmt.Sprintf(`{"code":"invalid_request","message":"%s"}`, escapeJSON(err.Error())))
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		http.Error(c.Writer, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Step 3: Stream agent processing in real time
	// The callback receives events AS the agent plans and executes — no artificial chunking.
	_, err := h.svc.StreamChat(c.Request.Context(), req, func(evt agent.StreamEvent) {
		switch evt.Type {
		case "action":
			metaJSON, _ := json.Marshal(evt.Meta)
			writeSSEEvent(c.Writer, "action", fmt.Sprintf(`{"text":%q,"meta":%s}`, evt.Content, string(metaJSON)))
		case "token":
			writeSSEEvent(c.Writer, "token", fmt.Sprintf(`{"text":%q}`, evt.Content))
		case "done":
			metaJSON, _ := json.Marshal(evt.Meta)
			writeSSEEvent(c.Writer, "done", fmt.Sprintf(`{"reason":"stop","meta":%s}`, string(metaJSON)))
		case "error":
			writeSSEEvent(c.Writer, "error", fmt.Sprintf(`{"code":"agent_error","message":%q}`, evt.Content))
		}
		flusher.Flush()
	})

	if err != nil {
		writeSSEEvent(c.Writer, "error", fmt.Sprintf(`{"code":"chat_error","message":%q}`, escapeJSON(err.Error())))
		flusher.Flush()
	}
}

// escapeJSON escapes a string for safe inclusion in a JSON string value.
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

func writeSSEEvent(w http.ResponseWriter, event string, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
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

// GenerateWorkflow generates a workflow JSON directly from a natural language requirement.
// POST /api/agent/generate-workflow
// Uses LLM to produce a complete workflow definition without running the agent pipeline.
func (h *AgentHandler) GenerateWorkflow(c *gin.Context) {
	var req struct {
		Requirement string `json:"requirement" binding:"required"`
		ProjectID   string `json:"projectId"`
		UserID      string `json:"userId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	// Call the dedicated workflow generator (single LLM call, no agent pipeline)
	result, err := h.svc.GenerateWorkflowFromNL(c.Request.Context(), req.ProjectID, req.Requirement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"workflow": result,
		},
	})
}

// GenerateAndRunWorkflow generates a workflow and runs it immediately.
// POST /api/agent/generate-and-run
// Delegates to the agent's Chat flow which handles creation + execution.
func (h *AgentHandler) GenerateAndRunWorkflow(c *gin.Context) {
	var req struct {
		Requirement string `json:"requirement" binding:"required"`
		ProjectID   string `json:"projectId"`
		UserID      string `json:"userId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	chatReq := service.ChatRequest{
		Message:   req.Requirement,
		ProjectID: req.ProjectID,
		UserID:    req.UserID,
	}
	resp, err := h.svc.Chat(c.Request.Context(), chatReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"workflow_id": resp.WorkflowID,
			"task_id":     resp.TaskID,
			"summary":     resp.Summary,
			"status":      resp.Status,
			"steps":       resp.Steps,
		},
	})
}