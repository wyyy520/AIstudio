package mcp

import (
	"context"
	"fmt"

	"github.com/aistudio/backend/internal/workflow"
)

// RuntimeAdapter adapts the MCP Manager to the workflow.MCPRuntime interface.
// This allows the workflow engine to call MCP tools transparently.
type RuntimeAdapter struct {
	manager *Manager
}

// NewRuntimeAdapter creates a new RuntimeAdapter.
func NewRuntimeAdapter(manager *Manager) *RuntimeAdapter {
	return &RuntimeAdapter{manager: manager}
}

// Connect establishes a connection to an MCP server.
func (a *RuntimeAdapter) Connect(ctx context.Context, server string) error {
	config, ok := a.manager.GetServerConfig(server)
	if !ok {
		return fmt.Errorf("server not found: %s", server)
	}
	return a.manager.Connect(ctx, config)
}

// Disconnect removes a connection to an MCP server.
func (a *RuntimeAdapter) Disconnect(ctx context.Context, server string) error {
	return a.manager.Disconnect(server)
}

// ListTools returns all tools available on a server.
func (a *RuntimeAdapter) ListTools(ctx context.Context, server string) ([]workflow.MCPTool, error) {
	tools := a.manager.ListToolsByServer(server)

	result := make([]workflow.MCPTool, 0, len(tools))
	for _, t := range tools {
		result = append(result, workflow.MCPTool{
			Name:        t.Name,
			Description: t.Description,
			Parameters: map[string]interface{}{
				"type":       t.InputSchema.Type,
				"properties": t.InputSchema.Properties,
				"required":   t.InputSchema.Required,
			},
			ReturnType: t.OutputSchema.Type,
		})
	}

	return result, nil
}

// CallTool invokes a tool on an MCP server.
func (a *RuntimeAdapter) CallTool(ctx context.Context, server string, toolName string, params map[string]interface{}) (*workflow.MCPResult, error) {
	resp, err := a.manager.CallTool(ctx, server, toolName, params)
	if err != nil {
		return &workflow.MCPResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &workflow.MCPResult{
		Success: resp.Success,
		Output:  resp.Output,
		Error:   resp.Error,
	}, nil
}

// ExecuteWorkflow sends a workflow to an MCP server for execution.
func (a *RuntimeAdapter) ExecuteWorkflow(ctx context.Context, server string, workflowJSON []byte) (*workflow.MCPResult, error) {
	// For first version, this delegates to a generic workflow execution tool
	resp, err := a.manager.CallTool(ctx, server, "execute_workflow", map[string]interface{}{
		"workflow": string(workflowJSON),
	})
	if err != nil {
		return &workflow.MCPResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &workflow.MCPResult{
		Success: resp.Success,
		Output:  resp.Output,
		Error:   resp.Error,
	}, nil
}

// Ensure RuntimeAdapter implements workflow.MCPRuntime
var _ workflow.MCPRuntime = (*RuntimeAdapter)(nil)