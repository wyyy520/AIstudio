package handlers

import (
	"fmt"
	"net/http"
	"strings"

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
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	var req service.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeSSEEvent(c.Writer, "error", fmt.Sprintf(`{"code":"invalid_request","message":"%s"}`, err.Error()))
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		http.Error(c.Writer, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	resp, err := h.svc.Chat(c.Request.Context(), req)
	if err != nil {
		writeSSEEvent(c.Writer, "error", fmt.Sprintf(`{"code":"chat_error","message":"%s"}`, err.Error()))
		flusher.Flush()
		return
	}

	writeSSEEvent(c.Writer, "action", `{"type":"planning","description":"Planning the response"}`)
	flusher.Flush()

	summary := resp.Summary
	if summary != "" {
		chunkSize := 5
		for i := 0; i < len(summary); i += chunkSize {
			end := i + chunkSize
			if end > len(summary) {
				end = len(summary)
			}
			chunk := summary[i:end]
			chunk = strings.ReplaceAll(chunk, "\\", "\\\\")
			chunk = strings.ReplaceAll(chunk, "\"", "\\\"")
			chunk = strings.ReplaceAll(chunk, "\n", "\\n")
			writeSSEEvent(c.Writer, "token", fmt.Sprintf(`{"text":"%s"}`, chunk))
			flusher.Flush()
		}
	}

	writeSSEEvent(c.Writer, "done", fmt.Sprintf(`{"reason":"stop","workflowId":"%s"}`, resp.WorkflowID))
	flusher.Flush()
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

// GenerateWorkflow generates a workflow from a natural language requirement.
// POST /api/agent/generate-workflow
// Delegates to the agent's Chat flow; the agent creates the workflow via its tools.
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
			"summary":     resp.Summary,
			"status":      resp.Status,
			"steps":       resp.Steps,
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