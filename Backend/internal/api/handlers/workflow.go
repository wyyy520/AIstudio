package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aistudio/backend/internal/common"
	"github.com/aistudio/backend/internal/service"
	"github.com/aistudio/backend/internal/task"
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

// List returns all workflows, optionally filtered by project ID.
func (h *WorkflowHandler) List(c *gin.Context) {
	projectID := c.Query("projectId")
	workflows, err := h.svc.List(projectID)
	if err != nil {
		common.RespondInternalError(c, "workflow", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": workflows})
}

// Get returns a single workflow by ID.
func (h *WorkflowHandler) Get(c *gin.Context) {
	wf, err := h.svc.Get(c.Param("id"))
	if err != nil {
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "workflow", "工作流未找到: "+c.Param("id"), "请检查工作流ID是否正确",
		))
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
		common.RespondError(c, http.StatusBadRequest, common.NewAPIError(
			common.ErrCodeBadRequest, "workflow", "请求参数无效: "+err.Error(), "请检查必填字段",
		))
		return
	}

	wf, err := h.svc.Create(req.ProjectID, req.Name, req.Definition)
	if err != nil {
		common.RespondInternalError(c, "workflow", err)
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
		common.RespondError(c, http.StatusBadRequest, common.NewAPIError(
			common.ErrCodeBadRequest, "workflow", "请求参数无效: "+err.Error(), "请检查请求格式",
		))
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
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "workflow", "工作流未找到: "+c.Param("id"), "请检查工作流ID是否正确",
		))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "updated", "data": wf})
}

// Delete removes a workflow.
func (h *WorkflowHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Param("id")); err != nil {
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "workflow", "工作流未找到: "+c.Param("id"), "请检查工作流ID是否正确",
		))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "deleted"})
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

	wf, err := h.svc.Get(workflowID)
	if err != nil {
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "workflow", "工作流未找到: "+workflowID, "请检查工作流ID是否正确",
		))
		return
	}

	if wf.Definition == "" {
		common.RespondError(c, http.StatusBadRequest, common.NewAPIError(
			common.ErrCodeWorkflowError, "workflow", "工作流定义为空", "请先设计工作流再运行",
		))
		return
	}

	// Parse definition to extract payload for the task
	var definition map[string]interface{}
	if err := json.Unmarshal([]byte(wf.Definition), &definition); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.NewAPIError(
			common.ErrCodeWorkflowError, "workflow", "工作流定义解析失败: "+err.Error(), "请检查工作流定义格式",
		))
		return
	}

	// Build the task payload
	payload := map[string]interface{}{
		"workflow_id":   workflowID,
		"project_id":    wf.ProjectID,
		"name":          wf.Name,
		"definition":    definition,
		"nodes":         definition["nodes"],
		"edges":         definition["edges"],
		"triggered_by":  "api",
		"triggered_at":  time.Now().Format(time.RFC3339),
	}

	// If task service is available, use the full chain
	if h.taskSvc != nil {
		taskID, err := h.taskSvc.Create(c.Request.Context(), service.CreateTaskRequest{
			ProjectID:  fmt.Sprintf("%d", wf.ProjectID),
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

		// Update workflow status to running
		h.svc.Update(workflowID, map[string]interface{}{"status": "running"})

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "workflow submitted",
			"data": gin.H{
				"task_id":     taskID,
				"workflow_id": workflowID,
				"status":      "waiting",
				"message":     "工作流已提交到任务队列，可通过 WebSocket 或 GET /api/tasks/:id/status 跟踪进度",
			},
		})
		return
	}

	// Fallback: direct execution (for environments without task manager)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Minute)
	defer cancel()

	result, err := h.svc.Run(ctx, workflowID)
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