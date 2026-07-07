package handlers

import (
	"net/http"

	"github.com/aistudio/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// WorkflowHandler handles workflow CRUD operations via the service layer.
type WorkflowHandler struct {
	svc *service.WorkflowService
}

// NewWorkflowHandler creates a new WorkflowHandler.
func NewWorkflowHandler(svc *service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{svc: svc}
}

// List returns all workflows, optionally filtered by project ID.
func (h *WorkflowHandler) List(c *gin.Context) {
	projectID := c.Query("projectId")
	workflows, err := h.svc.List(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": workflows})
}

// Get returns a single workflow by ID.
func (h *WorkflowHandler) Get(c *gin.Context) {
	wf, err := h.svc.Get(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": wf})
}

// Create creates a new workflow.
func (h *WorkflowHandler) Create(c *gin.Context) {
	var req struct {
		ProjectID  uint   `json:"projectId" binding:"required"`
		Name       string `json:"name" binding:"required"`
		Definition string `json:"definition"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	wf, err := h.svc.Create(req.ProjectID, req.Name, req.Definition)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "created", "data": wf})
}

// Update updates an existing workflow.
func (h *WorkflowHandler) Update(c *gin.Context) {
	var req struct {
		Name       string `json:"name"`
		Definition string `json:"definition"`
		Status     string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Definition != "" {
		updates["definition"] = req.Definition
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	wf, err := h.svc.Update(c.Param("id"), updates)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "updated", "data": wf})
}

// Delete removes a workflow.
func (h *WorkflowHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "deleted"})
}

// Run executes a workflow by its ID.
// This is the preserved "workflow run API".
func (h *WorkflowHandler) Run(c *gin.Context) {
	result, err := h.svc.Run(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": result})
}

// ListNodeTypes returns all registered workflow node types.
// This is the preserved "node list API".
func (h *WorkflowHandler) ListNodeTypes(c *gin.Context) {
	nodeTypes := h.svc.ListNodeTypes()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nodeTypes})
}