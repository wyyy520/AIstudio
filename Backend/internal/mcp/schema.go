package mcp

import "time"

// ConnectionStatus represents the current connection status.
type ConnectionStatus string

const (
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusConnected    ConnectionStatus = "connected"
	StatusError        ConnectionStatus = "error"
)

// TransportType indicates how the MCP server is connected.
type TransportType string

const (
	TransportTCP     TransportType = "tcp"
	TransportUDS     TransportType = "uds"
	TransportHTTP    TransportType = "http"
	TransportProcess TransportType = "process" // local subprocess
	TransportMock    TransportType = "mock"
)

// MCPConfig is the configuration for an MCP server.
// This is typically loaded from mcp.json.
type MCPConfig struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`   // "simulation", "data_source", "calculation", "other"
	Description  string            `json:"description"`
	Transport    TransportType      `json:"transport"`
	Endpoint     string            `json:"endpoint"` // "localhost:8080", "/var/run/mcp.sock", "./sumo"
	APIKey       string            `json:"api_key,omitempty"`
	TimeoutMs    int               `json:"timeout_ms,omitempty"`
	Enabled      bool              `json:"enabled"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// MCPTool represents a tool available from an MCP server.
type MCPTool struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	InputSchema  JSONSchema             `json:"input_schema"`
	OutputSchema JSONSchema             `json:"output_schema"`
	ServerName   string                 `json:"server_name"`
	IsMock       bool                   `json:"is_mock"`
	MockResponse map[string]interface{} `json:"mock_response,omitempty"`
}

// JSONSchema defines the JSON schema for a tool's input/output.
type JSONSchema struct {
	Type        string                 `json:"type"` // "object", "array", "string", "number", "boolean"
	Properties  map[string]JSONSchema `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Items       *JSONSchema            `json:"items,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// MCPConnection represents an active connection to an MCP server.
type MCPConnection struct {
	Name         string           `json:"name"`
	Config       MCPConfig        `json:"config"`
	Status       ConnectionStatus `json:"status"`
	LastConnected *time.Time      `json:"last_connected,omitempty"`
	LastError    string           `json:"last_error,omitempty"`
	Tools        []MCPTool        `json:"tools"`
	ConnectedAt  time.Time        `json:"connected_at"`
}

// CallRequest is a request to invoke an MCP tool.
type CallRequest struct {
	ServerName string                 `json:"server_name"`
	ToolName   string                 `json:"tool_name"`
	Input      map[string]interface{} `json:"input"`
	TimeoutMs  int                    `json:"timeout_ms,omitempty"`
}

// CallResponse is the response from an MCP tool invocation.
type CallResponse struct {
	Success    bool                   `json:"success"`
	Output     map[string]interface{} `json:"output"`
	Error      string                 `json:"error,omitempty"`
	DurationMs int64                  `json:"duration_ms"`
}

// MCPEvent represents an event from the MCP server.
type MCPEvent struct {
	ServerName string                 `json:"server_name"`
	EventType  string                 `json:"event_type"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
}

// MCPNodeParameters for workflow node configuration.
type MCPNodeParameters struct {
	ServerName string `json:"server_name"`
	ToolName   string `json:"tool_name"`
}

// MockMCPConfig creates a mock configuration for development.
func MockMCPConfig(name, toolType string) MCPConfig {
	return MCPConfig{
		Name:        name,
		Type:        toolType,
		Description: "Mock " + name + " for development",
		Transport:   TransportMock,
		Endpoint:    "mock://" + name,
		Enabled:     true,
	}
}