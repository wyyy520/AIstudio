package workflow

import (
	"context"
)

type TaskInfo struct {
	TaskID     string                 `json:"task_id"`
	Status     string                 `json:"status"`
	WorkflowID string                 `json:"workflow_id"`
	Progress   float64                `json:"progress"`
	Error      string                 `json:"error,omitempty"`
	NodeOutputs map[string]NodeResult `json:"node_outputs,omitempty"`
}

type TaskManager interface {
	SubmitTask(ctx context.Context, wf *Workflow) (string, error)
	GetTask(ctx context.Context, taskID string) (*TaskInfo, error)
	UpdateTask(ctx context.Context, taskID string, info TaskInfo) error
	ListTasks(ctx context.Context, workflowID string, status string) ([]TaskInfo, error)
	CancelTask(ctx context.Context, taskID string) error
}

type PluginInfo struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	ConfigSchema map[string]interface{} `json:"config_schema"`
	NodeTypes   []NodeDefinition       `json:"node_types"`
	Enabled     bool                   `json:"enabled"`
}

type PluginManager interface {
	GetPlugin(ctx context.Context, name string) (*PluginInfo, error)
	ListPlugins(ctx context.Context, pluginType string) ([]PluginInfo, error)
	RegisterPlugin(ctx context.Context, info PluginInfo) error
	UnregisterPlugin(ctx context.Context, name string) error
	EnablePlugin(ctx context.Context, name string) error
	DisablePlugin(ctx context.Context, name string) error
}

type PythonExecutionResult struct {
	Outputs map[string]interface{} `json:"outputs"`
	Error   string                 `json:"error,omitempty"`
	Logs    []LogEntry             `json:"logs,omitempty"`
}

type PythonEngine interface {
	ExecuteScript(ctx context.Context, script string, inputs map[string]interface{}) (*PythonExecutionResult, error)
	ExecuteFile(ctx context.Context, filePath string, inputs map[string]interface{}) (*PythonExecutionResult, error)
	InstallPackage(ctx context.Context, packageName string) error
	CreateEnvironment(ctx context.Context, name string) error
}

type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	ReturnType  string                 `json:"return_type"`
}

type MCPResult struct {
	Success bool                   `json:"success"`
	Output  interface{}            `json:"output,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Logs    []LogEntry             `json:"logs,omitempty"`
}

type MCPRuntime interface {
	Connect(ctx context.Context, server string) error
	Disconnect(ctx context.Context, server string) error
	ListTools(ctx context.Context, server string) ([]MCPTool, error)
	CallTool(ctx context.Context, server string, toolName string, params map[string]interface{}) (*MCPResult, error)
	ExecuteWorkflow(ctx context.Context, server string, workflowJSON []byte) (*MCPResult, error)
}

type EngineOptions struct {
	TaskManager   TaskManager
	PluginManager PluginManager
	PythonEngine  PythonEngine
	MCPRuntime    MCPRuntime
}
