package common

import (
	"encoding/json"
	"fmt"
)

type ErrorCode string

const (
	ErrCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeInternal       ErrorCode = "INTERNAL_ERROR"
	ErrCodeTimeout        ErrorCode = "TIMEOUT"
	ErrCodeEnvError       ErrorCode = "ENV_ERROR"
	ErrCodePluginError    ErrorCode = "PLUGIN_ERROR"
	ErrCodeWorkflowError  ErrorCode = "WORKFLOW_ERROR"
	ErrCodeTaskError      ErrorCode = "TASK_ERROR"
	ErrCodeEngineError    ErrorCode = "ENGINE_ERROR"
	ErrCodeMCPError       ErrorCode = "MCP_ERROR"
	ErrCodeAgentError     ErrorCode = "AGENT_ERROR"
	ErrCodeConfigError    ErrorCode = "CONFIG_ERROR"
)

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

func NewAPIError(code ErrorCode, module, message, solution string) *APIError {
	return &APIError{
		Code:     code,
		Module:   module,
		Message:  message,
		Solution: solution,
	}
}

func (e *APIError) WithDetails(details map[string]interface{}) *APIError {
	e.Details = details
	return e
}

func (e *APIError) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

func EnvNotFound(msg string) *APIError {
	return NewAPIError(ErrCodeEnvError, "python", msg, "请安装 Python 3.9+ 或检查 python_path 配置")
}

func PluginNotFound(name string) *APIError {
	return NewAPIError(ErrCodePluginError, "plugin", "插件未找到: "+name, "请在插件商店安装该插件")
}

func WorkflowInvalid(msg string) *APIError {
	return NewAPIError(ErrCodeWorkflowError, "workflow", msg, "请检查工作流定义是否正确")
}

func TaskFailed(taskID, msg string) *APIError {
	return NewAPIError(ErrCodeTaskError, "task", "任务执行失败: "+msg, "查看任务日志了解详情").
		WithDetails(map[string]interface{}{"task_id": taskID})
}

func EngineError(msg string) *APIError {
	return NewAPIError(ErrCodeEngineError, "engine", msg, "检查 Python 环境和依赖安装")
}

func MCPUnavailable(server string) *APIError {
	return NewAPIError(ErrCodeMCPError, "mcp", "MCP 服务器不可用: "+server, "检查 MCP 服务器是否已连接")
}

func ConfigMissing(key string) *APIError {
	return NewAPIError(ErrCodeConfigError, "config", "配置缺失: "+key, "请检查配置文件")
}
