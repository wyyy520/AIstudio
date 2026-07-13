package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aistudio/backend/internal/common"
	"github.com/aistudio/backend/internal/service"
	"github.com/aistudio/backend/internal/task"
	"github.com/aistudio/backend/internal/workflow"
	"github.com/gin-gonic/gin"
)

// WorkflowHandler handles workflow CRUD operations via the service layer.
type WorkflowHandler struct {
	svc     *service.WorkflowService
	taskSvc *service.TaskService
}

// NewWorkflowHandler creates a new WorkflowHandler.
func NewWorkflowHandler(svc *service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{svc: svc}
}

// SetTaskService sets the task service for the workflow run chain.
func (h *WorkflowHandler) SetTaskService(taskSvc *service.TaskService) {
	h.taskSvc = taskSvc
}

// List returns all workflow metadata entries, optionally filtered by project ID.
func (h *WorkflowHandler) List(c *gin.Context) {
	projectID := c.Query("projectId")
	workflows, err := h.svc.List(projectID)
	if err != nil {
		common.RespondInternalError(c, "workflow", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": workflows})
}

// Get returns a single workflow's full definition.
// Reads workflow.json from the project directory (Single Source of Truth).
func (h *WorkflowHandler) Get(c *gin.Context) {
	id := c.Param("id")

	// Get metadata from DB
	meta, err := h.svc.Get(id)
	if err != nil {
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "workflow", "工作流未找到: "+id, "请检查工作流ID是否正确",
		))
		return
	}

	// Read workflow.json from disk
	wf, err := h.svc.ReadWorkflowFile(meta.Path)
	if err != nil {
		common.RespondInternalError(c, "workflow", fmt.Errorf("读取工作流文件失败: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": wf})
}

// GetWorkflowByProject reads workflow.json from the project directory.
// GET /api/v1/projects/:id/workflow
func (h *WorkflowHandler) GetWorkflowByProject(c *gin.Context) {
	projectID := c.Param("id")

	wf, err := h.svc.ReadByProject(projectID)
	if err != nil {
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "workflow", "项目工作流未找到: "+projectID, "请检查项目ID是否正确",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": wf})
}

// Create creates a new workflow metadata entry and a default workflow.json.
func (h *WorkflowHandler) Create(c *gin.Context) {
	var req struct {
		ProjectID string `json:"projectId" binding:"required"`
		Name      string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.NewAPIError(
			common.ErrCodeBadRequest, "workflow", "请求参数无效: "+err.Error(), "请检查必填字段",
		))
		return
	}

	meta, err := h.svc.Create(req.ProjectID, req.Name)
	if err != nil {
		common.RespondInternalError(c, "workflow", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "created", "data": meta})
}

// Update updates the workflow.json file on disk.
// PUT /api/v1/projects/:id/workflow
func (h *WorkflowHandler) Update(c *gin.Context) {
	projectID := c.Param("id")

	var wf workflow.Workflow
	if err := c.ShouldBindJSON(&wf); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.NewAPIError(
			common.ErrCodeBadRequest, "workflow", "请求参数无效: "+err.Error(), "请检查JSON格式",
		))
		return
	}

	if err := h.svc.WriteByProject(projectID, &wf); err != nil {
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "workflow", "项目工作流未找到: "+projectID, "请检查项目ID是否正确",
		))
		return
	}

	// Update metadata
	_ = h.svc.UpdateMetadata(projectID, wf.Name)

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "updated", "data": wf})
}

// Delete removes a workflow metadata entry and optionally the workflow.json file.
func (h *WorkflowHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "workflow", "工作流未找到: "+id, "请检查工作流ID是否正确",
		))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "deleted"})
}

// Validate validates the workflow.json for a project.
// POST /api/v1/projects/:id/workflow/validate
func (h *WorkflowHandler) Validate(c *gin.Context) {
	projectID := c.Param("id")

	wf, err := h.svc.ReadByProject(projectID)
	if err != nil {
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "workflow", "项目工作流未找到: "+projectID, "请检查项目ID是否正确",
		))
		return
	}

	result := workflow.Validate(wf)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "validated", "data": result})
}

// Run executes a workflow via the full execution chain:
//
//	Frontend Click Run
//	  ↓
//	POST /api/workflows/:id/run
//	  ↓
//	Backend creates Task via Task Manager
//	  ↓
//	Task Manager dispatches to workflow.TaskHandler
//	  ↓
//	Workflow Engine executes nodes
//	  ↓
//	Plugin Manager resolves node types
//	  ↓
//	Python Engine executes plugin code
//	  ↓
//	Returns task_id for status tracking via WebSocket
func (h *WorkflowHandler) Run(c *gin.Context) {
	workflowID := c.Param("id")

	meta, err := h.svc.Get(workflowID)
	if err != nil {
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "workflow", "工作流未找到: "+workflowID, "请检查工作流ID是否正确",
		))
		return
	}

	// Read workflow definition from file
	wf, err := h.svc.ReadWorkflowFile(meta.Path)
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, common.NewAPIError(
			common.ErrCodeWorkflowError, "workflow", "读取工作流文件失败: "+err.Error(), "请检查workflow.json是否有效",
		))
		return
	}

	// Build the task payload
	payload := map[string]interface{}{
		"workflowId":  workflowID,
		"projectId":   meta.ProjectID,
		"name":        wf.Name,
		"nodes":       wf.Nodes,
		"edges":       wf.Edges,
		"triggeredBy": "api",
		"triggeredAt": time.Now().Format(time.RFC3339),
	}

	// If task service is available, use the full chain
	if h.taskSvc != nil {
		ctx := c.Request.Context()
		taskID, err := h.taskSvc.Create(ctx, service.CreateTaskRequest{
			ProjectID:  fmt.Sprintf("%d", meta.ProjectID),
			WorkflowID: workflowID,
			Type:       "workflow",
			Name:       wf.Name,
			Handler:    "workflow",
			Priority:   int(task.PriorityNormal),
			Payload:    payload,
		})
		if err != nil {
			common.RespondError(c, http.StatusInternalServerError, common.NewAPIError(
				common.ErrCodeTaskError, "task", "创建任务失败: "+err.Error(), "查看日志了解详情",
			))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "workflow submitted",
			"data": gin.H{
				"id":         taskID,
				"workflowId": workflowID,
				"status":     "waiting",
				"message":    "工作流已提交到任务队列，可通过 WebSocket 或 GET /api/tasks/:id/status 跟踪进度",
			},
		})
		return
	}

	// Fallback: direct execution (for environments without task manager)
	result, err := h.svc.Run(workflowID)
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, common.NewAPIError(
			common.ErrCodeWorkflowError, "workflow", "工作流执行失败: "+err.Error(), "查看日志了解详情",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

// ListNodeTypes returns all registered workflow node types.
func (h *WorkflowHandler) ListNodeTypes(c *gin.Context) {
	nodeTypes := h.svc.ListNodeTypes()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nodeTypes})
}

// UpdateWorkflow is an alias for Update.
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	h.Update(c)
}

// SaveWorkflow reads the current workflow.json, merges with request, and saves.
func (h *WorkflowHandler) SaveWorkflow(c *gin.Context) {
	projectID := c.Param("id")

	var patch workflow.Workflow
	if err := c.ShouldBindJSON(&patch); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.NewAPIError(
			common.ErrCodeBadRequest, "workflow", "请求参数无效: "+err.Error(), "请检查JSON格式",
		))
		return
	}

	// Read current workflow
	wf, err := h.svc.ReadByProject(projectID)
	if err != nil {
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "workflow", "项目工作流未找到: "+projectID, "请检查项目ID是否正确",
		))
		return
	}

	// Merge nodes and edges (preserving existing if not provided)
	if patch.Nodes != nil {
		wf.Nodes = patch.Nodes
	}
	if patch.Edges != nil {
		wf.Edges = patch.Edges
	}
	if patch.Target != "" {
		wf.Target = patch.Target
	}
	if patch.Name != "" {
		wf.Name = patch.Name
	}
	if patch.Description != "" {
		wf.Description = patch.Description
	}

	if err := h.svc.WriteByProject(projectID, wf); err != nil {
		common.RespondInternalError(c, "workflow", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "saved", "data": wf})
}

// ValidateBody validates a workflow definition from the request body.
func (h *WorkflowHandler) ValidateBody(c *gin.Context) {
	var wf workflow.Workflow
	if err := c.ShouldBindJSON(&wf); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.NewAPIError(
			common.ErrCodeBadRequest, "workflow", "请求参数无效: "+err.Error(), "请检查JSON格式",
		))
		return
	}

	result := workflow.Validate(&wf)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "validated", "data": result})
}

// SerializeJSON returns the JSON representation of a workflow for a project.
func (h *WorkflowHandler) SerializeJSON(c *gin.Context) {
	projectID := c.Param("id")

	wf, err := h.svc.ReadByProject(projectID)
	if err != nil {
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "workflow", "项目工作流未找到: "+projectID, "请检查项目ID是否正确",
		))
		return
	}

	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		common.RespondInternalError(c, "workflow", err)
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}
