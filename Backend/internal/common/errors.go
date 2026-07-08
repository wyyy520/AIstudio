package common

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorCode represents a machine-readable error identifier.
type ErrorCode string

const (
	ErrCodeBadRequest    ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden     ErrorCode = "FORBIDDEN"
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
	ErrCodeTimeout       ErrorCode = "TIMEOUT"
	ErrCodeEnvError      ErrorCode = "ENV_ERROR"
	ErrCodePluginError   ErrorCode = "PLUGIN_ERROR"
	ErrCodeWorkflowError ErrorCode = "WORKFLOW_ERROR"
	ErrCodeTaskError     ErrorCode = "TASK_ERROR"
	ErrCodeEngineError   ErrorCode = "ENGINE_ERROR"
	ErrCodeMCPError      ErrorCode = "MCP_ERROR"
	ErrCodeAgentError    ErrorCode = "AGENT_ERROR"
	ErrCodeConfigError   ErrorCode = "CONFIG_ERROR"
)

// APIError represents a structured API error response.
type APIError struct {
	Code     ErrorCode              `json:"code"`
	Module   string                 `json:"module"`
	Message  string                 `json:"message"`
	Solution string                 `json:"solution,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Module, e.Code, e.Message)
}

// NewAPIError creates a new API error.
func NewAPIError(code ErrorCode, module, message, solution string) *APIError {
	return &APIError{
		Code:     code,
		Module:   module,
		Message:  message,
		Solution: solution,
	}
}

// WithDetails adds extra details to the error.
func (e *APIError) WithDetails(details map[string]interface{}) *APIError {
	e.Details = details
	return e
}

// ToJSON converts the error to JSON bytes.
func (e *APIError) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// RespondError sends a structured error response via Gin.
func RespondError(c *gin.Context, httpStatus int, err *APIError) {
	c.JSON(httpStatus, gin.H{
		"code":    -1,
		"message": err.Message,
		"data":    err,
	})
}

// RespondInternalError sends a 500 error with structured format.
func RespondInternalError(c *gin.Context, module string, err error) {
	apiErr := NewAPIError(ErrCodeInternal, module, err.Error(), "查看日志获取详细信息")
	RespondError(c, http.StatusInternalServerError, apiErr)
}

// ---- Common error factory functions ----

// EnvNotFound creates a "Python not found" error.
func EnvNotFound(msg string) *APIError {
	return NewAPIError(ErrCodeEnvError, "python", msg, "请安装 Python 3.9+ 或检查 python_path 配置")
}

// PluginNotFound creates a plugin not found error.
func PluginNotFound(name string) *APIError {
	return NewAPIError(ErrCodePluginError, "plugin", "插件未找到: "+name, "请在插件商店安装该插件")
}

// WorkflowInvalid creates a workflow validation error.
func WorkflowInvalid(msg string) *APIError {
	return NewAPIError(ErrCodeWorkflowError, "workflow", msg, "请检查工作流定义是否正确")
}

// TaskFailed creates a task failure error.
func TaskFailed(taskID, msg string) *APIError {
	return NewAPIError(ErrCodeTaskError, "task", "任务执行失败: "+msg, "查看任务日志了解详情").
		WithDetails(map[string]interface{}{"task_id": taskID})
}

// EngineError creates a Python engine error.
func EngineError(msg string) *APIError {
	return NewAPIError(ErrCodeEngineError, "engine", msg, "检查 Python 环境和依赖安装")
}

// MCPUnavailable creates an MCP connection error.
func MCPUnavailable(server string) *APIError {
	return NewAPIError(ErrCodeMCPError, "mcp", "MCP 服务器不可用: "+server, "检查 MCP 服务器是否已连接")
}

// ConfigMissing creates a configuration error.
func ConfigMissing(key string) *APIError {
	return NewAPIError(ErrCodeConfigError, "config", "配置缺失: "+key, "请检查配置文件")
}