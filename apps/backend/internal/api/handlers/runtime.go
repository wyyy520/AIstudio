package handlers

import (
	"net/http"
	"time"

	"github.com/aistudio/backend/internal/common"
	"github.com/aistudio/backend/internal/compiler"
	"github.com/aistudio/backend/internal/runtime"
	"github.com/aistudio/backend/internal/service"
	"github.com/aistudio/packages/workflow"
	"github.com/gin-gonic/gin"
)

// RuntimeHandler handles the full Compile→Generate→Execute→Status pipeline.
type RuntimeHandler struct {
	svc  *service.RuntimeService
	comp compiler.Compiler
}

// NewRuntimeHandler creates a new RuntimeHandler.
func NewRuntimeHandler(svc *service.RuntimeService, comp compiler.Compiler) *RuntimeHandler {
	return &RuntimeHandler{svc: svc, comp: comp}
}

// CompileRequest is the request body for compilation.
type CompileRequest struct {
	Target      string `json:"target"`
	ProjectName string `json:"projectName"`
}

// Compile compiles a project's workflow into runnable project files.
// POST /api/projects/:id/compile
func (h *RuntimeHandler) Compile(c *gin.Context) {
	projectID := c.Param("id")

	var req CompileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = CompileRequest{}
	}

	target := workflow.Target(req.Target)
	if target == "" {
		target = workflow.TargetPython
	}

	ctx := c.Request.Context()

	// 1. Generate project files
	result, err := h.svc.Compile(ctx, projectID, target, req.ProjectName)
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, common.NewAPIError(
			common.ErrCodeInternal, "compile", "编译失败: "+err.Error(), "查看日志了解详情",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "compilation completed",
		"data": gin.H{
			"target":      result.Target,
			"projectRoot": result.ProjectRoot,
			"entryPoints": result.EntryPoints,
			"files":       result.Files,
			"duration":    result.Duration.String(),
			"generatorId": result.GeneratorID,
		},
	})
}

// ============================================================================
// Execute Pipeline
// ============================================================================

// RunRequest is the request body for execution.
type RunRequest struct {
	Target  string `json:"target"`
	Timeout int    `json:"timeout"`
}

// Run compiles and executes a project's workflow.
// POST /api/projects/:id/run
func (h *RuntimeHandler) Run(c *gin.Context) {
	projectID := c.Param("id")

	var req RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = RunRequest{}
	}

	target := workflow.Target(req.Target)
	if target == "" {
		target = workflow.TargetPython
	}

	ctx := c.Request.Context()

	// 1. Compile (generate project files)
	result, err := h.svc.Compile(ctx, projectID, target, "")
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, common.NewAPIError(
			common.ErrCodeInternal, "run", "编译失败: "+err.Error(), "查看日志了解详情",
		))
		return
	}

	// 2. Execute the generated project
	timeout := time.Duration(req.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 300 * time.Second
	}

	entryPoint := ""
	if len(result.EntryPoints) > 0 {
		entryPoint = result.EntryPoints[0]
	}

	runResult := h.svc.ExecuteCommand(ctx, runtime.RunConfig{
		ProjectDir: result.ProjectRoot,
		EntryPoint: entryPoint,
		Timeout:    timeout,
	})

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "execution completed",
		"data": gin.H{
			"runId":       runResult.RunID,
			"status":      runResult.Status,
			"exitCode":    runResult.ExitCode,
			"stdout":      runResult.Stdout,
			"stderr":      runResult.Stderr,
			"duration":    runResult.Duration.String(),
			"projectRoot": result.ProjectRoot,
			"entryPoints": result.EntryPoints,
			"startedAt":   runResult.StartedAt,
			"completedAt": runResult.CompletedAt,
		},
	})
}

// Stop stops a running execution.
// POST /api/projects/:id/stop
func (h *RuntimeHandler) Stop(c *gin.Context) {
	runID := c.Param("runId")
	if runID == "" {
		var req struct {
			RunID string `json:"runId"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.RunID == "" {
			common.RespondError(c, http.StatusBadRequest, common.NewAPIError(
				common.ErrCodeBadRequest, "runtime", "缺少 runId 参数", "请提供 runId",
			))
			return
		}
		runID = req.RunID
	}

	if err := h.svc.Stop(c.Request.Context(), runID); err != nil {
		common.RespondError(c, http.StatusInternalServerError, common.NewAPIError(
			common.ErrCodeInternal, "runtime", "停止失败: "+err.Error(), "",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "stopped"})
}

// Status returns the runtime status for a given run.
// GET /api/runtime/status/:runId
func (h *RuntimeHandler) Status(c *gin.Context) {
	runID := c.Param("runId")

	status, err := h.svc.Status(c.Request.Context(), runID)
	if err != nil {
		common.RespondError(c, http.StatusNotFound, common.NewAPIError(
			common.ErrCodeNotFound, "runtime", "运行记录未找到: "+runID, "",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": status})
}

// ListRunning returns all currently running executions.
// GET /api/runtime/list
func (h *RuntimeHandler) ListRunning(c *gin.Context) {
	running := h.svc.ListRunning()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": running})
}

// Detect checks the runtime environment.
// POST /api/runtime/detect
func (h *RuntimeHandler) Detect(c *gin.Context) {
	var req struct {
		Python   string   `json:"python"`
		Packages []string `json:"packages"`
		GPU      bool     `json:"gpu"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Python = ">=3.9"
	}

	r := &runtime.Requirement{
		Name:     "default",
		Version:  "1.0",
		Python:   req.Python,
		Packages: req.Packages,
		GPU:      req.GPU,
	}

	report, err := h.svc.Detect(c.Request.Context(), r)
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, common.NewAPIError(
			common.ErrCodeInternal, "runtime", "环境检测失败: "+err.Error(), "",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "detection completed", "data": report})
}
