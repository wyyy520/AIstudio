package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aistudio/backend/internal/agent"
)

// AgentMCPTool is an adapter that wraps an MCP tool as an agent-compatible Tool.
// This allows the Agent to call MCP tools just like any other agent tool.
type AgentMCPTool struct {
	MCPTool      MCPTool
	CallFunc     func(ctx context.Context, serverName, toolName string, input map[string]interface{}) (map[string]interface{}, error)
}

// Name returns the tool name, prefixed with "mcp_".
func (t *AgentMCPTool) Name() string {
	return "mcp_" + t.MCPTool.ServerName + "_" + t.MCPTool.Name
}

// Description returns a description of the tool.
func (t *AgentMCPTool) Description() string {
	return fmt.Sprintf("[MCP] %s: %s", t.MCPTool.ServerName, t.MCPTool.Description)
}

// Parameters returns the tool parameters derived from the input schema.
func (t *AgentMCPTool) Parameters() []agent.ToolParameter {
	var params []agent.ToolParameter

	for name, prop := range t.MCPTool.InputSchema.Properties {
		required := false
		for _, r := range t.MCPTool.InputSchema.Required {
			if r == name {
				required = true
				break
			}
		}
		params = append(params, agent.ToolParameter{
			Name:        name,
			Type:        prop.Type,
			Description: prop.Description,
			Required:    required,
		})
	}

	return params
}

// Execute calls the MCP tool through the provided call function.
func (t *AgentMCPTool) Execute(ctx context.Context, params map[string]interface{}) (*agent.ToolResult, error) {
	log.Printf("[mcp-tool] executing %s.%s", t.MCPTool.ServerName, t.MCPTool.Name)

	if t.MCPTool.IsMock {
		// Return mock response directly
		return &agent.ToolResult{
			Success: true,
			Data:    t.MCPTool.MockResponse,
		}, nil
	}

	if t.CallFunc == nil {
		return &agent.ToolResult{
			Success: false,
			Error:   "MCP tool not connected: no call function configured",
		}, nil
	}

	output, err := t.CallFunc(ctx, t.MCPTool.ServerName, t.MCPTool.Name, params)
	if err != nil {
		return &agent.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &agent.ToolResult{
		Success: true,
		Data:    output,
	}, nil
}

// MCPToolInfo provides a summary of an MCP tool for discovery.
type MCPToolInfo struct {
	Name        string     `json:"name"`
	ServerName  string     `json:"server_name"`
	Description string     `json:"description"`
	InputSchema JSONSchema `json:"input_schema"`
	IsMock      bool       `json:"is_mock"`
}

// ToMCPToolInfo converts an MCPTool to MCPToolInfo.
func (t *MCPTool) ToMCPToolInfo() MCPToolInfo {
	return MCPToolInfo{
		Name:        t.Name,
		ServerName:  t.ServerName,
		Description: t.Description,
		InputSchema: t.InputSchema,
		IsMock:      t.IsMock,
	}
}

// ---- MCPConfig Loader ----

// MCPConfigsFile is the structure of mcp.json.
type MCPConfigsFile struct {
	Servers []MCPConfig `json:"servers"`
}

// LoadMCPConfigs loads MCP server configurations from JSON data.
func LoadMCPConfigs(data []byte) ([]MCPConfig, error) {
	var configs MCPConfigsFile
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse mcp.json: %w", err)
	}
	return configs.Servers, nil
}

// SaveMCPConfigs serializes MCP server configurations to JSON.
func SaveMCPConfigs(configs []MCPConfig) ([]byte, error) {
	file := MCPConfigsFile{Servers: configs}
	return json.MarshalIndent(file, "", "  ")
}

// ---- Workflow Node Integration ----

// MCPNodeExecutor executes MCP nodes within workflows.
// It implements the workflow.ExecutableNode interface.
type MCPNodeExecutor struct {
	Client    *Client
	ServerName string
	ToolName   string
}

// NewMCPNodeExecutor creates a new MCP node executor.
func NewMCPNodeExecutor(client *Client, serverName, toolName string) *MCPNodeExecutor {
	return &MCPNodeExecutor{
		Client:     client,
		ServerName: serverName,
		ToolName:   toolName,
	}
}

// ExecuteNode runs the MCP tool as a workflow node.
func (e *MCPNodeExecutor) ExecuteNode(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	start := time.Now()
	log.Printf("[mcp-workflow-node] executing %s.%s", e.ServerName, e.ToolName)

	resp, err := e.Client.CallTool(ctx, e.ServerName, e.ToolName, input)
	if err != nil {
		return nil, fmt.Errorf("mcp node execution failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("mcp node returned error: %s", resp.Error)
	}

	duration := time.Since(start)
	log.Printf("[mcp-workflow-node] %s.%s completed in %v", e.ServerName, e.ToolName, duration)

	return resp.Output, nil
}