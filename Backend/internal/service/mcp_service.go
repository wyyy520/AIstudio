package service

import (
	"context"
	"fmt"
	"log"

	appcfg "github.com/aistudio/backend/internal/config"
	"github.com/aistudio/backend/internal/mcp"
)

// MCPService provides the business logic for MCP operations.
// It exposes APIs for the HTTP handlers and integrates with the agent system.
type MCPService struct {
	manager *mcp.Manager
	config  appcfg.MCPConfig
}

// NewMCPService creates a new MCP service.
func NewMCPService(cfg appcfg.MCPConfig) *MCPService {
	manager := mcp.NewManager()

	service := &MCPService{
		manager: manager,
		config:  cfg,
	}

	// Initialize with default mock servers for development
	manager.InitializeDefaultMocks()

	// Load configuration from file if path is set
	if cfg.ConfigPath != "" {
		if err := manager.LoadConfigs(cfg.ConfigPath); err != nil {
			log.Printf("[mcp-service] warning: failed to load mcp config: %v", err)
		}
	}

	log.Println("[mcp-service] MCP service initialized")
	return service
}

// Manager returns the underlying MCP manager.
func (s *MCPService) Manager() *mcp.Manager {
	return s.manager
}

// ListTools returns all registered MCP tools.
func (s *MCPService) ListTools() []mcp.MCPTool {
	return s.manager.ListTools()
}

// ListToolsByServer returns tools for a specific server.
func (s *MCPService) ListToolsByServer(serverName string) []mcp.MCPTool {
	return s.manager.ListToolsByServer(serverName)
}

// ListServers returns all registered MCP servers.
func (s *MCPService) ListServers() []mcp.MCPConfig {
	return s.manager.ListServers()
}

// ListConnectedServers returns all connected servers.
func (s *MCPService) ListConnectedServers() []string {
	return s.manager.ListConnectedServers()
}

// Connect connects to an MCP server.
func (s *MCPService) Connect(ctx context.Context, config mcp.MCPConfig) error {
	if err := s.manager.Connect(ctx, config); err != nil {
		return err
	}
	return nil
}

// Disconnect disconnects from an MCP server.
func (s *MCPService) Disconnect(serverName string) error {
	return s.manager.Disconnect(serverName)
}

// CallTool calls an MCP tool with the given input.
func (s *MCPService) CallTool(ctx context.Context, serverName, toolName string, input map[string]interface{}) (*mcp.CallResponse, error) {
	return s.manager.CallTool(ctx, serverName, toolName, input)
}

// Call executes a structured CallRequest.
func (s *MCPService) Call(ctx context.Context, req mcp.CallRequest) (*mcp.CallResponse, error) {
	return s.manager.Call(ctx, req)
}

// GetStatus returns the current MCP system status.
func (s *MCPService) GetStatus() map[string]interface{} {
	status := s.manager.GetStatus()
	status["config_path"] = s.config.ConfigPath
	return status
}

// AddServer adds a new server configuration.
func (s *MCPService) AddServer(config mcp.MCPConfig) error {
	return s.manager.AddConfig(config)
}

// RemoveServer removes a server configuration.
func (s *MCPService) RemoveServer(name string) error {
	return s.manager.RemoveConfig(name)
}

// SaveConfig saves the current configuration to file.
func (s *MCPService) SaveConfig() error {
	if s.config.ConfigPath == "" {
		return fmt.Errorf("config path not set")
	}
	return s.manager.SaveConfigs(s.config.ConfigPath)
}

// ExportConfigJSON returns the current configuration as JSON.
func (s *MCPService) ExportConfigJSON() ([]byte, error) {
	return s.manager.ExportMCPConfigJSON()
}

// GetAgentMCPTools returns all MCP tools as agent-compatible tools.
// These should be registered with the agent's tool registry.
func (s *MCPService) GetAgentMCPTools() []*mcp.AgentMCPTool {
	return s.manager.GetAgentMCPTools()
}