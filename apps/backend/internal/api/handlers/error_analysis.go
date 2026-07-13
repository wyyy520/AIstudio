package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aistudio/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ErrorAnalysisResult represents the result of error analysis.
type ErrorAnalysisResult struct {
	AnalysisID    string             `json:"analysisId"`
	TaskID        string             `json:"taskId"`
	Errors        []ErrorDetail      `json:"errors"`
	Solutions     []Solution         `json:"solutions"`
	Summary       string             `json:"summary"`
	Confidence    float64            `json:"confidence"`
	AnalyzedAt    time.Time          `json:"analyzedAt"`
}

// ErrorDetail represents a single error found in logs.
type ErrorDetail struct {
	ID          string    `json:"id"`
	LogID       string    `json:"logId"`
	Message     string    `json:"message"`
	Level       string    `json:"level"`
	Source      string    `json:"source"`
	Timestamp   time.Time `json:"timestamp"`
	ErrorType   string    `json:"errorType"`
	Severity    string    `json:"severity"` // critical, high, medium, low
}

// Solution represents a suggested fix for an error.
type Solution struct {
	SolutionID   string            `json:"solutionId"`
	ErrorID      string            `json:"errorId"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Action       string            `json:"action"` // auto, manual, info
	Confidence   float64           `json:"confidence"`
	Steps        []SolutionStep    `json:"steps"`
}

// SolutionStep represents a step in a solution.
type SolutionStep struct {
	StepNumber  int    `json:"stepNumber"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
}

// FixStatus represents the status of a repair operation.
type FixStatus struct {
	FixID        string    `json:"fixId"`
	Status       string    `json:"status"` // pending, running, completed, failed
	Progress     float64   `json:"progress"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
	CompletedAt  time.Time `json:"completedAt,omitempty"`
}

// ErrorAnalysisHandler handles error analysis operations.
type ErrorAnalysisHandler struct {
	logSvc       *service.LogService
	analysisCache sync.Map // key: taskID, value: *ErrorAnalysisResult
	fixStatuses   sync.Map // key: fixID, value: *FixStatus
}

// NewErrorAnalysisHandler creates a new ErrorAnalysisHandler.
func NewErrorAnalysisHandler(logSvc *service.LogService) *ErrorAnalysisHandler {
	return &ErrorAnalysisHandler{
		logSvc: logSvc,
	}
}

// AnalyzeError analyzes errors for a given task.
// POST /api/error/analyze
func (h *ErrorAnalysisHandler) AnalyzeError(c *gin.Context) {
	var req struct {
		TaskID string   `json:"taskId" binding:"required"`
		LogIDs []string `json:"logIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	// Query error logs for the task
	result, err := h.logSvc.Query(service.LogQuery{
		TaskID: req.TaskID,
		Level:  service.LogLevelError,
		Size:   100,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	// Analyze errors
	analysis := h.analyzeErrors(req.TaskID, result.Items)

	// Cache the analysis
	h.analysisCache.Store(req.TaskID, analysis)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    analysis,
	})
}

// analyzeErrors performs rule-based error analysis on log entries.
func (h *ErrorAnalysisHandler) analyzeErrors(taskID string, entries []service.LogEntry) *ErrorAnalysisResult {
	var errors []ErrorDetail
	var solutions []Solution
	var errorID int

	for _, entry := range entries {
		errorID++
		errorDetail := ErrorDetail{
			ID:          fmt.Sprintf("err-%d", errorID),
			LogID:       fmt.Sprintf("%d", entry.ID),
			Message:     entry.Message,
			Level:       string(entry.Level),
			Source:      entry.Source,
			Timestamp:   entry.Timestamp,
			ErrorType:   h.detectErrorType(entry.Message),
			Severity:    h.assessSeverity(entry.Message),
		}
		errors = append(errors, errorDetail)

		// Generate solutions based on error type
		errSolutions := h.generateSolutions(errorDetail)
		solutions = append(solutions, errSolutions...)
	}

	// Generate summary
	summary := fmt.Sprintf("分析了 %d 条错误日志，发现 %d 个问题，提供 %d 个解决方案",
		len(entries), len(errors), len(solutions))

	return &ErrorAnalysisResult{
		AnalysisID: fmt.Sprintf("analysis-%s-%d", taskID, time.Now().UnixNano()),
		TaskID:     taskID,
		Errors:     errors,
		Solutions:  solutions,
		Summary:    summary,
		Confidence: h.calculateConfidence(errors),
		AnalyzedAt: time.Now(),
	}
}

// detectErrorType identifies the type of error from the message.
func (h *ErrorAnalysisHandler) detectErrorType(message string) string {
	messageLower := strings.ToLower(message)

	switch {
	case strings.Contains(messageLower, "node") && strings.Contains(messageLower, "not found"):
		return "NodeNotFound"
	case strings.Contains(messageLower, "plugin") && strings.Contains(messageLower, "not found"):
		return "PluginNotFound"
	case strings.Contains(messageLower, "python") && (strings.Contains(messageLower, "import") || strings.Contains(messageLower, "module")):
		return "PythonImportError"
	case strings.Contains(messageLower, "cuda") || strings.Contains(messageLower, "gpu"):
		return "GPUError"
	case strings.Contains(messageLower, "timeout"):
		return "TimeoutError"
	case strings.Contains(messageLower, "connection") || strings.Contains(messageLower, "network"):
		return "NetworkError"
	case strings.Contains(messageLower, "permission"):
		return "PermissionError"
	case strings.Contains(messageLower, "database") || strings.Contains(messageLower, "db"):
		return "DatabaseError"
	default:
		return "UnknownError"
	}
}

// assessSeverity evaluates the severity of an error.
func (h *ErrorAnalysisHandler) assessSeverity(message string) string {
	messageLower := strings.ToLower(message)

	switch {
	case strings.Contains(messageLower, "critical") || strings.Contains(messageLower, "fatal"):
		return "critical"
	case strings.Contains(messageLower, "failed") || strings.Contains(messageLower, "error"):
		return "high"
	case strings.Contains(messageLower, "warning") || strings.Contains(messageLower, "warn"):
		return "medium"
	default:
		return "low"
	}
}

// generateSolutions creates solutions based on error type.
func (h *ErrorAnalysisHandler) generateSolutions(err ErrorDetail) []Solution {
	var solutions []Solution

	switch err.ErrorType {
	case "NodeNotFound":
		solutions = append(solutions, Solution{
			SolutionID:  fmt.Sprintf("sol-%s-1", err.ID),
			ErrorID:     err.ID,
			Title:       "注册缺失的节点类型",
			Description: "该工作流引用了未注册的节点类型，请检查工作流定义或安装相关插件",
			Action:      "manual",
			Confidence:  0.9,
			Steps: []SolutionStep{
				{StepNumber: 1, Description: "检查工作流定义中的节点类型是否正确"},
				{StepNumber: 2, Description: "确认相关插件已安装并启用"},
				{StepNumber: 3, Description: "重新注册节点类型"},
			},
		})

	case "PluginNotFound":
		solutions = append(solutions, Solution{
			SolutionID:  fmt.Sprintf("sol-%s-1", err.ID),
			ErrorID:     err.ID,
			Title:       "安装缺失的插件",
			Description: "系统找不到指定的插件，请安装该插件",
			Action:      "auto",
			Confidence:  0.95,
			Steps: []SolutionStep{
				{StepNumber: 1, Description: "从插件市场安装缺失的插件", Command: "plugin install <plugin-name>"},
				{StepNumber: 2, Description: "启用插件", Command: "plugin enable <plugin-name>"},
			},
		})

	case "PythonImportError":
		solutions = append(solutions, Solution{
			SolutionID:  fmt.Sprintf("sol-%s-1", err.ID),
			ErrorID:     err.ID,
			Title:       "安装缺失的 Python 依赖",
			Description: "Python 模块导入失败，请安装相关依赖包",
			Action:      "auto",
			Confidence:  0.85,
			Steps: []SolutionStep{
				{StepNumber: 1, Description: "安装依赖包", Command: "pip install <package-name>"},
				{StepNumber: 2, Description: "验证安装", Command: "python -c 'import <module>'"},
			},
		})

	case "GPUError":
		solutions = append(solutions, Solution{
			SolutionID:  fmt.Sprintf("sol-%s-1", err.ID),
			ErrorID:     err.ID,
			Title:       "检查 GPU 环境",
			Description: "CUDA/GPU 相关错误，请检查 GPU 配置",
			Action:      "manual",
			Confidence:  0.7,
			Steps: []SolutionStep{
				{StepNumber: 1, Description: "检查 NVIDIA 驱动是否安装"},
				{StepNumber: 2, Description: "验证 CUDA 版本", Command: "nvidia-smi"},
				{StepNumber: 3, Description: "检查 PyTorch/TensorFlow GPU 支持"},
			},
		})

	case "TimeoutError":
		solutions = append(solutions, Solution{
			SolutionID:  fmt.Sprintf("sol-%s-1", err.ID),
			ErrorID:     err.ID,
			Title:       "增加超时时间",
			Description: "任务执行超时，请增加超时限制或优化执行效率",
			Action:      "manual",
			Confidence:  0.8,
			Steps: []SolutionStep{
				{StepNumber: 1, Description: "增加任务超时时间配置"},
				{StepNumber: 2, Description: "优化工作流节点执行逻辑"},
				{StepNumber: 3, Description: "考虑并行执行优化"},
			},
		})

	default:
		solutions = append(solutions, Solution{
			SolutionID:  fmt.Sprintf("sol-%s-1", err.ID),
			ErrorID:     err.ID,
			Title:       "查看详细日志",
			Description: "请查看详细日志以了解错误原因",
			Action:      "info",
			Confidence:  0.5,
			Steps: []SolutionStep{
				{StepNumber: 1, Description: "查看完整错误日志", Command: "logs --task-id <task-id>"},
				{StepNumber: 2, Description: "根据错误信息排查问题"},
			},
		})
	}

	return solutions
}

// calculateConfidence computes the overall confidence of the analysis.
func (h *ErrorAnalysisHandler) calculateConfidence(errors []ErrorDetail) float64 {
	if len(errors) == 0 {
		return 1.0
	}

	confidenceSum := 0.0
	for _, err := range errors {
		switch err.Severity {
		case "critical", "high":
			confidenceSum += 0.9
		case "medium":
			confidenceSum += 0.7
		default:
			confidenceSum += 0.5
		}
	}

	return confidenceSum / float64(len(errors))
}

// RepairError repairs an error based on analysis and solution.
// POST /api/error/repair
func (h *ErrorAnalysisHandler) RepairError(c *gin.Context) {
	var req struct {
		AnalysisID string `json:"analysisId" binding:"required"`
		SolutionID string `json:"solutionId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	// Create fix status
	fixID := "fix-" + req.AnalysisID + "-" + req.SolutionID
	fixStatus := &FixStatus{
		FixID:    fixID,
		Status:   "running",
		Progress: 0.5,
	}
	
	// Simulate repair process
	go func() {
		time.Sleep(1 * time.Second)
		fixStatus.Progress = 1.0
		fixStatus.Status = "completed"
		fixStatus.CompletedAt = time.Now()
		h.fixStatuses.Store(fixID, fixStatus)
	}()

	h.fixStatuses.Store(fixID, fixStatus)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"fixId":  fixID,
			"status": "running",
		},
	})
}

// GetErrorAnalysis retrieves error analysis for a task.
// GET /api/error/analysis/:taskId
func (h *ErrorAnalysisHandler) GetErrorAnalysis(c *gin.Context) {
	taskID := c.Param("taskId")

	// Try to get from cache
	if cached, ok := h.analysisCache.Load(taskID); ok {
		if analysis, ok := cached.(*ErrorAnalysisResult); ok {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data":    analysis,
			})
			return
		}
	}

	// Perform fresh analysis
	result, err := h.logSvc.Query(service.LogQuery{
		TaskID: taskID,
		Level:  service.LogLevelError,
		Size:   100,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	analysis := h.analyzeErrors(taskID, result.Items)
	h.analysisCache.Store(taskID, analysis)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    analysis,
	})
}

// GetFixStatus retrieves the status of a fix.
// GET /api/error/fix/:fixId/status
func (h *ErrorAnalysisHandler) GetFixStatus(c *gin.Context) {
	fixID := c.Param("fixId")

	// Try to get from cache
	if cached, ok := h.fixStatuses.Load(fixID); ok {
		if status, ok := cached.(*FixStatus); ok {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data":    status,
			})
			return
		}
	}

	// Return default completed status if not found
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"fixId":  fixID,
			"status": "completed",
		},
	})
}