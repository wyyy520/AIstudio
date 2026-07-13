// Package diagnostic provides AI-powered error analysis for the platform.
//
// Diagnostic is responsible for:
//   1. Analyzing execution logs to detect errors and root causes
//   2. Translating technical error messages to human-readable form
//   3. Mapping errors to specific workflow nodes
//   4. Generating fix suggestions
//
// Diagnostic does NOT modify logs — that's LogCenter's job.
// Diagnostic does NOT modify workflows — that's the user's job.
package diagnostic

import (
	"context"
	"fmt"
	"time"

	"github.com/aistudio/backend/internal/logcenter"
	"github.com/aistudio/backend/internal/workflow"
	"gorm.io/gorm"
)

// ============================================================================
// Diagnostic Interface
// ============================================================================

// Diagnostic provides AI-powered error analysis.
type Diagnostic interface {
	// Analyze analyzes a log entry and returns diagnostic results.
	Analyze(ctx context.Context, entry *logcenter.Entry, wf *workflow.Workflow) (*Result, error)

	// AnalyzeTask analyzes all logs for a task.
	AnalyzeTask(ctx context.Context, taskID string, entries []logcenter.Entry) (*TaskDiagnostic, error)

	// Translate translates a technical error message to human-readable form.
	Translate(ctx context.Context, message string, lang string) (string, error)

	// SuggestFix generates fix suggestions for a diagnostic result.
	SuggestFix(ctx context.Context, result *Result) ([]*FixSuggestion, error)

	// GetHistory returns the diagnostic history for a task.
	GetHistory(taskID string) ([]*Result, error)
}

// ============================================================================
// Types
// ============================================================================

// Result contains the diagnostic analysis of a single error.
type Result struct {
	ID          string           `json:"id"`
	TaskID      string           `json:"taskId,omitempty"`
	Timestamp   time.Time        `json:"timestamp"`
	Severity    Severity         `json:"severity"`
	Summary     string           `json:"summary"`
	Detail      string           `json:"detail"`
	WorkflowID  string           `json:"workflowId,omitempty"`
	NodeID      string           `json:"nodeId,omitempty"`
	NodeName    string           `json:"nodeName,omitempty"`
	RootCause   string           `json:"rootCause"`
	Suggestions []*FixSuggestion `json:"suggestions,omitempty"`
	RawLog      string           `json:"rawLog"`
	AnalyzedAt  time.Time        `json:"analyzedAt"`
}

// TaskDiagnostic contains the full diagnostic analysis of a task execution.
type TaskDiagnostic struct {
	TaskID      string    `json:"taskId"`
	WorkflowID  string    `json:"workflowId,omitempty"`
	OverallStatus string  `json:"overallStatus"` // success, warning, error
	Results     []*Result `json:"results,omitempty"`
	Summary     string    `json:"summary"`      // Overall summary
	Duration    string    `json:"duration,omitempty"`
	AnalyzedAt  time.Time `json:"analyzedAt"`
}

// Severity represents the severity of a diagnostic result.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// FixSuggestion represents a suggested fix for an error.
type FixSuggestion struct {
	ID          string    `json:"id"`
	Type        FixType   `json:"type"`     // parameter, dependency, environment, workflow, code
	Title       string    `json:"title"`     // Short title
	Description string    `json:"description"` // Detailed description
	AutoFix     bool      `json:"autoFix"`   // Can be applied automatically
	Confidence  float64   `json:"confidence"` // 0.0 to 1.0
	Action      *FixAction `json:"action,omitempty"`
}

// FixType identifies the type of fix.
type FixType string

const (
	FixTypeParameter   FixType = "parameter"
	FixTypeDependency  FixType = "dependency"
	FixTypeEnvironment FixType = "environment"
	FixTypeWorkflow    FixType = "workflow"
	FixTypeCode        FixType = "code"
)

// FixAction represents an actionable fix.
type FixAction struct {
	Type    string      `json:"type"`    // update_config, install_package, modify_workflow, run_command
	Target  string      `json:"target"`  // What to modify
	Value   interface{} `json:"value"`   // New value
	Command string      `json:"command,omitempty"` // Shell command to run
}

// ============================================================================
// Engine Implementation
// ============================================================================

// DiagnosticRecord is the database model for persisted diagnostic results.
type DiagnosticRecord struct {
	ID         string    `gorm:"primaryKey"`
	TaskID     string    `gorm:"index"`
	WorkflowID string    `gorm:"index"`
	Severity   string
	Summary    string
	Detail     string
	NodeID     string
	RootCause  string
	CreatedAt  time.Time
}

// Engine is the default Diagnostic implementation.
// It provides rule-based analysis with optional LLM-powered enhancement.
type Engine struct {
	rules     []Rule
	llmClient LLMClient
	db        *gorm.DB
}

// Rule defines a diagnostic rule for pattern matching.
type Rule struct {
	ID          string
	Pattern     string
	Severity    Severity
	Summary     string
	RootCause   string
	Suggestions []*FixSuggestion
}

// LLMClient is an optional interface for LLM-powered analysis.
type LLMClient interface {
	AnalyzeLog(ctx context.Context, logContent string, workflowContext string) (*Result, error)
	Translate(ctx context.Context, message string, targetLang string) (string, error)
}

// NewEngine creates a new Diagnostic Engine.
func NewEngine(db *gorm.DB) *Engine {
	if db != nil {
		db.AutoMigrate(&DiagnosticRecord{})
	}
	return &Engine{
		rules: defaultRules(),
		db:    db,
	}
}

// WithLLM sets the LLM client for enhanced analysis.
func (e *Engine) WithLLM(client LLMClient) *Engine {
	e.llmClient = client
	return e
}

// Analyze analyzes a log entry and returns diagnostic results.
func (e *Engine) Analyze(ctx context.Context, entry *logcenter.Entry, wf *workflow.Workflow) (*Result, error) {
	// 1. Try LLM-based analysis first
	if e.llmClient != nil {
		result, err := e.llmClient.AnalyzeLog(ctx, entry.Raw, getWorkflowContext(wf))
		if err == nil && result != nil {
			result.AnalyzedAt = time.Now()
			return result, nil
		}
	}

	// 2. Fall back to rule-based analysis
	result := e.analyzeWithRules(entry, wf)
	result.AnalyzedAt = time.Now()
	return result, nil
}

// AnalyzeTask analyzes all logs for a task.
func (e *Engine) AnalyzeTask(ctx context.Context, taskID string, entries []logcenter.Entry) (*TaskDiagnostic, error) {
	results := make([]*Result, 0, len(entries))
	hasError := false

	for _, entry := range entries {
		if entry.Level == logcenter.LevelError || entry.Level == logcenter.LevelFatal {
			result, err := e.Analyze(ctx, &entry, nil)
			if err == nil {
				result.TaskID = taskID
				results = append(results, result)
				hasError = true
			}
		}
	}

	status := "success"
	if hasError {
		status = "error"
	}

	return &TaskDiagnostic{
		TaskID:        taskID,
		OverallStatus: status,
		Results:       results,
		Summary:       formatSummary(results),
		AnalyzedAt:    time.Now(),
	}, nil
}

// Translate translates a technical error message to human-readable form.
func (e *Engine) Translate(ctx context.Context, message string, lang string) (string, error) {
	// Try LLM translation first
	if e.llmClient != nil {
		translated, err := e.llmClient.Translate(ctx, message, lang)
		if err == nil {
			return translated, nil
		}
	}

	// Fall back to rule-based translation
	return e.translateWithRules(message, lang), nil
}

// SuggestFix generates fix suggestions for a diagnostic result.
func (e *Engine) SuggestFix(ctx context.Context, result *Result) ([]*FixSuggestion, error) {
	return result.Suggestions, nil
}

// SaveResult persists a diagnostic result to the database.
func (e *Engine) SaveResult(ctx context.Context, result *Result) error {
	if e.db == nil {
		return nil
	}
	record := DiagnosticRecord{
		ID:         result.ID,
		TaskID:     result.TaskID,
		WorkflowID: result.WorkflowID,
		Severity:   string(result.Severity),
		Summary:    result.Summary,
		Detail:     result.Detail,
		NodeID:     result.NodeID,
		RootCause:  result.RootCause,
		CreatedAt:  result.Timestamp,
	}
	return e.db.Create(&record).Error
}

// GetHistory returns the diagnostic history for a task.
func (e *Engine) GetHistory(taskID string) ([]*Result, error) {
	if e.db == nil {
		return nil, nil
	}
	var records []DiagnosticRecord
	if err := e.db.Where("task_id = ?", taskID).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	var results []*Result
	for _, r := range records {
		results = append(results, &Result{
			ID:         r.ID,
			TaskID:     r.TaskID,
			Timestamp:  r.CreatedAt,
			Severity:   Severity(r.Severity),
			Summary:    r.Summary,
			Detail:     r.Detail,
			NodeID:     r.NodeID,
			RootCause:  r.RootCause,
		})
	}
	return results, nil
}

// ============================================================================
// Rule-Based Analysis
// ============================================================================

func (e *Engine) analyzeWithRules(entry *logcenter.Entry, wf *workflow.Workflow) *Result {
	msg := entry.Message
	raw := entry.Raw

	for _, rule := range e.rules {
		if containsSubstring(msg, rule.Pattern) || containsSubstring(raw, rule.Pattern) {
			// Map to workflow node if possible
			nodeID, nodeName := e.mapToNode(msg, wf)

			return &Result{
				ID:          generateID(),
				Timestamp:   entry.Timestamp,
				Severity:    rule.Severity,
				Summary:     rule.Summary,
				Detail:      msg,
				WorkflowID:  entry.WorkflowID,
				NodeID:      nodeID,
				NodeName:    nodeName,
				RootCause:   rule.RootCause,
				Suggestions: rule.Suggestions,
				RawLog:      raw,
			}
		}
	}

	// Default: unknown error
	return &Result{
		ID:        generateID(),
		Timestamp: entry.Timestamp,
		Severity:  SeverityError,
		Summary:   "Unknown error occurred",
		Detail:    msg,
		RawLog:    raw,
		RootCause: "Could not automatically determine the root cause",
		Suggestions: []*FixSuggestion{
			{
				ID:          "check-logs",
				Type:        FixTypeCode,
				Title:       "Check full logs",
				Description: "Review the complete log output for more details",
				Confidence:  0.5,
			},
		},
	}
}

func (e *Engine) mapToNode(message string, wf *workflow.Workflow) (string, string) {
	if wf == nil {
		return "", ""
	}

	for _, node := range wf.Nodes {
		if containsSubstring(message, node.Name) || containsSubstring(message, node.ID) {
			return node.ID, node.Name
		}
		// Check config values
		for _, val := range node.Config {
			if str, ok := val.(string); ok {
				if containsSubstring(message, str) {
					return node.ID, node.Name
				}
			}
		}
	}
	return "", ""
}

func (e *Engine) translateWithRules(message string, lang string) string {
	// Common error translations
	translations := map[string]map[string]string{
		"Chinese": {
			"CUDA out of memory": "CUDA 显存不足，请减小 batch_size 或使用更小的模型",
			"No module named":    "缺少 Python 包，请使用 pip install 安装",
			"FileNotFoundError":  "文件未找到，请检查路径是否正确",
			"Connection refused": "连接被拒绝，请检查服务是否正在运行",
			"Permission denied":  "权限不足，请检查文件权限",
			"Timeout":            "操作超时，请检查网络连接或增加超时时间",
			"SyntaxError":        "Python 语法错误，请检查代码",
			"KeyError":           "字典键不存在，请检查键名",
			"IndexError":         "列表索引越界，请检查索引值",
			"ValueError":         "数值错误，请检查参数值",
			"TypeError":          "类型错误，请检查参数类型",
			"AttributeError":     "对象没有该属性，请检查对象类型",
			"ImportError":        "导入错误，请检查包是否正确安装",
			"RuntimeError":       "运行时错误，请检查日志获取详细信息",
			"ZeroDivisionError":  "除零错误，请检查除数是否为零",
		},
	}

	if lang == "Chinese" || lang == "zh" || lang == "zh-CN" {
		for pattern, translation := range translations["Chinese"] {
			if containsSubstring(message, pattern) {
				return translation
			}
		}
	}

	return message
}

// ============================================================================
// Default Rules
// ============================================================================

func defaultRules() []Rule {
	return []Rule{
		{
			ID:        "cuda-oom",
			Pattern:   "CUDA out of memory",
			Severity:  SeverityError,
			Summary:   "GPU 显存不足（CUDA Out of Memory）",
			RootCause: "训练批量大小（batch_size）过大或模型过大，导致 GPU 显存耗尽",
			Suggestions: []*FixSuggestion{
				{
					ID:          "reduce-batch-size",
					Type:        FixTypeParameter,
					Title:       "Reduce batch size",
					Description: "Reduce batch_size to 8, 4, or 2 to fit in GPU memory",
					AutoFix:     true,
					Confidence:  0.9,
					Action: &FixAction{
						Type:   "update_config",
						Target: "batch_size",
						Value:  "8",
					},
				},
				{
					ID:          "use-smaller-model",
					Type:        FixTypeParameter,
					Title:       "Use smaller model",
					Description: "Switch to a smaller model variant (e.g., yolov8n instead of yolov8l)",
					AutoFix:     true,
					Confidence:  0.8,
					Action: &FixAction{
						Type:   "update_config",
						Target: "model",
						Value:  "yolov8n.pt",
					},
				},
				{
					ID:          "enable-gradient-checkpoint",
					Type:        FixTypeParameter,
					Title:       "Enable gradient checkpointing",
					Description: "Enable gradient checkpointing to reduce memory usage",
					AutoFix:     false,
					Confidence:  0.7,
				},
			},
		},
		{
			ID:        "module-not-found",
			Pattern:   "No module named",
			Severity:  SeverityError,
			Summary:   "缺少 Python 依赖包",
			RootCause: "所需的 Python 包未安装",
			Suggestions: []*FixSuggestion{
				{
					ID:          "install-package",
					Type:        FixTypeDependency,
					Title:       "Install missing package",
					Description: "Install the missing Python package using pip",
					AutoFix:     true,
					Confidence:  0.95,
					Action: &FixAction{
						Type:    "install_package",
						Target:  "pip",
						Command: "pip install <package_name>",
					},
				},
			},
		},
		{
			ID:        "file-not-found",
			Pattern:   "FileNotFoundError",
			Severity:  SeverityError,
			Summary:   "文件未找到",
			RootCause: "指定的文件路径不存在或无法访问",
			Suggestions: []*FixSuggestion{
				{
					ID:          "check-path",
					Type:        FixTypeParameter,
					Title:       "Check file path",
					Description: "Verify the file path exists and is accessible",
					AutoFix:     false,
					Confidence:  0.8,
				},
			},
		},
		{
			ID:        "connection-refused",
			Pattern:   "Connection refused",
			Severity:  SeverityError,
			Summary:   "连接被拒绝",
			RootCause: "目标服务未运行或端口未被监听",
			Suggestions: []*FixSuggestion{
				{
					ID:          "start-service",
					Type:        FixTypeEnvironment,
					Title:       "Start the service",
					Description: "Ensure the required service is running",
					AutoFix:     false,
					Confidence:  0.7,
				},
			},
		},
		{
			ID:        "timeout",
			Pattern:   "Timeout",
			Severity:  SeverityWarning,
			Summary:   "操作超时",
			RootCause: "操作超过预设时间限制",
			Suggestions: []*FixSuggestion{
				{
					ID:          "increase-timeout",
					Type:        FixTypeParameter,
					Title:       "Increase timeout",
					Description: "Increase the timeout value for the operation",
					AutoFix:     true,
					Confidence:  0.8,
					Action: &FixAction{
						Type:   "update_config",
						Target: "timeout",
						Value:  "3600",
					},
				},
			},
		},
		{
			ID:        "disk-full",
			Pattern:   "No space left on device",
			Severity:  SeverityCritical,
			Summary:   "磁盘空间不足",
			RootCause: "磁盘空间已满，无法写入数据",
			Suggestions: []*FixSuggestion{
				{
					ID:          "clean-disk",
					Type:        FixTypeEnvironment,
					Title:       "Free disk space",
					Description: "Clean up temporary files and old models to free disk space",
					AutoFix:     false,
					Confidence:  0.9,
				},
			},
		},
		{
			ID:        "permission-denied",
			Pattern:   "Permission denied",
			Severity:  SeverityError,
			Summary:   "权限不足",
			RootCause: "当前用户没有足够的权限访问该资源",
			Suggestions: []*FixSuggestion{
				{
					ID:          "check-permissions",
					Type:        FixTypeEnvironment,
					Title:       "Check file permissions",
					Description: "Ensure the file has the correct permissions",
					AutoFix:     false,
					Confidence:  0.8,
				},
			},
		},
	}
}

// ============================================================================
// Helpers
// ============================================================================

func getWorkflowContext(wf *workflow.Workflow) string {
	if wf == nil {
		return ""
	}
	ctx := fmt.Sprintf("Workflow: %s (%s)\nTarget: %s\nNodes: %d\n",
		wf.Name, wf.ID, wf.Target, len(wf.Nodes))
	for _, node := range wf.Nodes {
		ctx += fmt.Sprintf("  - %s (%s): %s\n", node.Name, node.Type, node.Description)
	}
	return ctx
}

func formatSummary(results []*Result) string {
	if len(results) == 0 {
		return "No errors detected"
	}

	errorCount := 0
	warningCount := 0
	for _, r := range results {
		if r.Severity == SeverityError || r.Severity == SeverityCritical {
			errorCount++
		} else if r.Severity == SeverityWarning {
			warningCount++
		}
	}

	return fmt.Sprintf("Found %d errors and %d warnings", errorCount, warningCount)
}

func generateID() string {
	return fmt.Sprintf("diag-%d", time.Now().UnixNano())
}

func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}