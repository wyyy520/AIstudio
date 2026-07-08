package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// Tool defines the interface for actions the Agent can perform.
// Each tool is a discrete capability that the Agent can invoke.
type Tool interface {
	Name() string
	Description() string
	Parameters() []ToolParameter
	Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error)
}

// ToolParameter describes a single parameter for a tool.
type ToolParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "number", "boolean", "object"
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ToolResult is the result of executing a tool.
type ToolResult struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error,omitempty"`
}

// ToolRegistry manages all registered tools.
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry creates a new tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *ToolRegistry) Register(t Tool) {
	r.tools[t.Name()] = t
}

// Get retrieves a tool by name.
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// List returns all registered tools with their descriptions.
func (r *ToolRegistry) List() []ToolInfo {
	var result []ToolInfo
	for _, t := range r.tools {
		result = append(result, ToolInfo{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Parameters(),
		})
	}
	return result
}

// ToolInfo is a lightweight description of a tool for LLM prompts.
type ToolInfo struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  []ToolParameter `json:"parameters"`
}

// ---- Tool Registry Accessors (used by Agent) ----

// CheckEnvironmentTool executes an environment check via the external checker.
type CheckEnvironmentTool struct {
	CheckFn func(ctx context.Context) (map[string]interface{}, error)
}

func (t *CheckEnvironmentTool) Name() string        { return "check_environment" }
func (t *CheckEnvironmentTool) Description() string  { return "Check the current AI development environment status including Python, CUDA, and dependencies." }
func (t *CheckEnvironmentTool) Parameters() []ToolParameter { return nil }

func (t *CheckEnvironmentTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	log.Println("[agent-tool] checking environment...")
	if t.CheckFn == nil {
		return &ToolResult{Success: false, Error: "check_environment: not configured"}, nil
	}
	data, err := t.CheckFn(ctx)
	if err != nil {
		return &ToolResult{Success: false, Error: err.Error()}, nil
	}
	return &ToolResult{Success: true, Data: data}, nil
}

// ListPluginsTool lists installed/available plugins.
type ListPluginsTool struct {
	ListFn func(ctx context.Context) ([]map[string]interface{}, error)
}

func (t *ListPluginsTool) Name() string       { return "list_plugins" }
func (t *ListPluginsTool) Description() string { return "List all available plugins and their capabilities." }
func (t *ListPluginsTool) Parameters() []ToolParameter { return nil }

func (t *ListPluginsTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	log.Println("[agent-tool] listing plugins...")
	if t.ListFn == nil {
		return &ToolResult{Success: false, Error: "list_plugins: not configured"}, nil
	}
	plugins, err := t.ListFn(ctx)
	if err != nil {
		return &ToolResult{Success: false, Error: err.Error()}, nil
	}
	return &ToolResult{Success: true, Data: map[string]interface{}{"plugins": plugins}}, nil
}

// CreateWorkflowTool generates a workflow definition.
type CreateWorkflowTool struct {
	CreateFn func(ctx context.Context, name string, workflowJSON json.RawMessage) (string, error)
}

func (t *CreateWorkflowTool) Name() string       { return "create_workflow" }
func (t *CreateWorkflowTool) Description() string { return "Create a new AI workflow from a JSON definition." }
func (t *CreateWorkflowTool) Parameters() []ToolParameter {
	return []ToolParameter{
		{Name: "name", Type: "string", Description: "Name of the workflow", Required: true},
		{Name: "workflow", Type: "object", Description: "The workflow JSON definition", Required: true},
	}
}

func (t *CreateWorkflowTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	name, _ := params["name"].(string)
	if name == "" {
		return &ToolResult{Success: false, Error: "workflow name is required"}, nil
	}

	workflowData, err := json.Marshal(params["workflow"])
	if err != nil {
		return &ToolResult{Success: false, Error: fmt.Sprintf("invalid workflow JSON: %v", err)}, nil
	}

	log.Printf("[agent-tool] creating workflow: %s", name)
	if t.CreateFn == nil {
		return &ToolResult{Success: false, Error: "create_workflow: not configured"}, nil
	}
	id, err := t.CreateFn(ctx, name, workflowData)
	if err != nil {
		return &ToolResult{Success: false, Error: err.Error()}, nil
	}

	return &ToolResult{Success: true, Data: map[string]interface{}{
		"workflow_id": id,
		"name":        name,
	}}, nil
}

// RunWorkflowTool submits a workflow for execution.
type RunWorkflowTool struct {
	RunFn func(ctx context.Context, workflowID string, params map[string]interface{}) (string, error)
}

func (t *RunWorkflowTool) Name() string        { return "run_workflow" }
func (t *RunWorkflowTool) Description() string  { return "Execute a workflow by its ID with optional parameters." }
func (t *RunWorkflowTool) Parameters() []ToolParameter {
	return []ToolParameter{
		{Name: "workflow_id", Type: "string", Description: "ID of the workflow to run", Required: true},
		{Name: "parameters", Type: "object", Description: "Optional runtime parameters", Required: false},
	}
}

func (t *RunWorkflowTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	workflowID, _ := params["workflow_id"].(string)
	if workflowID == "" {
		return &ToolResult{Success: false, Error: "workflow_id is required"}, nil
	}

	runtimeParams, _ := params["parameters"].(map[string]interface{})

	log.Printf("[agent-tool] running workflow: %s", workflowID)
	if t.RunFn == nil {
		return &ToolResult{Success: false, Error: "run_workflow: not configured"}, nil
	}
	taskID, err := t.RunFn(ctx, workflowID, runtimeParams)
	if err != nil {
		return &ToolResult{Success: false, Error: err.Error()}, nil
	}

	return &ToolResult{Success: true, Data: map[string]interface{}{
		"task_id":     taskID,
		"workflow_id": workflowID,
	}}, nil
}

// GetTaskStatusTool retrieves task status.
type GetTaskStatusTool struct {
	StatusFn func(ctx context.Context, taskID string) (map[string]interface{}, error)
}

func (t *GetTaskStatusTool) Name() string       { return "get_task_status" }
func (t *GetTaskStatusTool) Description() string { return "Check the status of a running or completed task." }
func (t *GetTaskStatusTool) Parameters() []ToolParameter {
	return []ToolParameter{
		{Name: "task_id", Type: "string", Description: "ID of the task to check", Required: true},
	}
}

func (t *GetTaskStatusTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	taskID, _ := params["task_id"].(string)
	if taskID == "" {
		return &ToolResult{Success: false, Error: "task_id is required"}, nil
	}

	log.Printf("[agent-tool] checking task status: %s", taskID)
	if t.StatusFn == nil {
		return &ToolResult{Success: false, Error: "get_task_status: not configured"}, nil
	}
	data, err := t.StatusFn(ctx, taskID)
	if err != nil {
		return &ToolResult{Success: false, Error: err.Error()}, nil
	}

	return &ToolResult{Success: true, Data: data}, nil
}

// InstallPluginTool installs a plugin by name.
type InstallPluginTool struct {
	InstallFn func(ctx context.Context, name string) error
}

func (t *InstallPluginTool) Name() string       { return "install_plugin" }
func (t *InstallPluginTool) Description() string { return "Install a plugin from the marketplace by name." }
func (t *InstallPluginTool) Parameters() []ToolParameter {
	return []ToolParameter{
		{Name: "plugin_name", Type: "string", Description: "Name of the plugin to install", Required: true},
	}
}

func (t *InstallPluginTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	name, _ := params["plugin_name"].(string)
	if name == "" {
		return &ToolResult{Success: false, Error: "plugin_name is required"}, nil
	}

	log.Printf("[agent-tool] installing plugin: %s", name)
	if t.InstallFn == nil {
		return &ToolResult{Success: false, Error: "install_plugin: not configured"}, nil
	}
	if err := t.InstallFn(ctx, name); err != nil {
		return &ToolResult{Success: false, Error: err.Error()}, nil
	}

	return &ToolResult{Success: true, Data: map[string]interface{}{
		"plugin_name": name,
		"status":      "installed",
	}}, nil
}