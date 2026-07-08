package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Manager is the central MCP component that orchestrates clients, servers, and tools.
// It provides the unified API for the rest of the system.
type Manager struct {
	mu       sync.RWMutex
	registry *Registry
	client   *Client
	servers  map[string]*Server // all server instances (connected or not)
	configs  []MCPConfig        // configuration from mcp.json
}

// NewManager creates a new MCP manager.
func NewManager() *Manager {
	registry := NewRegistry()
	client := NewClient(registry)

	return &Manager{
		registry: registry,
		client:   client,
		servers:  make(map[string]*Server),
		configs:  make([]MCPConfig, 0),
	}
}

// Registry returns the tool registry for external access.
func (m *Manager) Registry() *Registry {
	return m.registry
}

// Client returns the MCP client for external access.
func (m *Manager) Client() *Client {
	return m.client
}

// ---- Connection Management ----

// Connect establishes a connection to a server by config.
func (m *Manager) Connect(ctx context.Context, config MCPConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create server instance
	server := NewServer(config, m.registry)
	m.servers[config.Name] = server

	// Connect via client
	if err := m.client.ConnectWithServer(server); err != nil {
		delete(m.servers, config.Name)
		return fmt.Errorf("failed to connect to %s: %w", config.Name, err)
	}

	// Track config
	found := false
	for i, c := range m.configs {
		if c.Name == config.Name {
			m.configs[i] = config
			found = true
			break
		}
	}
	if !found {
		m.configs = append(m.configs, config)
	}

	log.Printf("[mcp-manager] connected to server: %s", config.Name)
	return nil
}

// ConnectWithServer connects using a pre-built server instance.
func (m *Manager) ConnectWithServer(server *Server) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	config := server.Config()
	m.servers[config.Name] = server

	if err := m.client.ConnectWithServer(server); err != nil {
		delete(m.servers, config.Name)
		return fmt.Errorf("failed to connect to %s: %w", config.Name, err)
	}

	// Track config
	found := false
	for i, c := range m.configs {
		if c.Name == config.Name {
			m.configs[i] = config
			found = true
			break
		}
	}
	if !found {
		m.configs = append(m.configs, config)
	}

	log.Printf("[mcp-manager] connected with server: %s", config.Name)
	return nil
}

// Disconnect removes a connection to a server.
func (m *Manager) Disconnect(serverName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.client.Disconnect(serverName); err != nil {
		return err
	}

	delete(m.servers, serverName)

	// Remove from configs
	for i, c := range m.configs {
		if c.Name == serverName {
			m.configs = append(m.configs[:i], m.configs[i+1:]...)
			break
		}
	}

	log.Printf("[mcp-manager] disconnected from server: %s", serverName)
	return nil
}

// IsConnected checks if a server is connected.
func (m *Manager) IsConnected(serverName string) bool {
	return m.client.IsConnected(serverName)
}

// ListConnectedServers returns all connected server names.
func (m *Manager) ListConnectedServers() []string {
	return m.client.ListConnectedServers()
}

// ---- Tool Calling ----

// CallTool invokes a tool on a connected server.
func (m *Manager) CallTool(ctx context.Context, serverName, toolName string, input map[string]interface{}) (*CallResponse, error) {
	return m.client.CallTool(ctx, serverName, toolName, input)
}

// Call sends a structured CallRequest and returns the response.
func (m *Manager) Call(ctx context.Context, req CallRequest) (*CallResponse, error) {
	return m.client.Call(ctx, req)
}

// ---- Tool Management ----

// RegisterTool registers a tool on a server.
func (m *Manager) RegisterTool(serverName string, tool MCPTool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, ok := m.servers[serverName]
	if !ok {
		return fmt.Errorf("server not found: %s", serverName)
	}

	tool.ServerName = serverName
	if err := m.registry.RegisterTool(tool); err != nil {
		return err
	}

	// If the tool has a mock response, register the handler
	if tool.MockResponse != nil {
		server.RegisterHandler(tool.Name, func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
			return tool.MockResponse, nil
		})
	}

	return nil
}

// RegisterHandler registers a handler for a tool on a server.
func (m *Manager) RegisterHandler(serverName, toolName string, handler ToolHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, ok := m.servers[serverName]
	if !ok {
		return fmt.Errorf("server not found: %s", serverName)
	}

	server.RegisterHandler(toolName, handler)

	// Also register the tool in the registry
	m.registry.RegisterTool(MCPTool{
		Name:        toolName,
		Description: toolName + " on " + serverName,
		ServerName:  serverName,
	})

	return nil
}

// ListTools returns all registered MCP tools.
func (m *Manager) ListTools() []MCPTool {
	return m.registry.ListTools()
}

// ListToolsByServer returns tools for a specific server.
func (m *Manager) ListToolsByServer(serverName string) []MCPTool {
	return m.registry.ListToolsByServer(serverName)
}

// ListServers returns all registered server configurations.
func (m *Manager) ListServers() []MCPConfig {
	return m.registry.ListServers()
}

// GetServerConfig returns a server's configuration.
func (m *Manager) GetServerConfig(name string) (MCPConfig, bool) {
	return m.registry.GetServer(name)
}

// ---- Config File Management ----

// LoadConfigs loads MCP server configurations from a JSON file.
func (m *Manager) LoadConfigs(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[mcp-manager] config file not found: %s, using defaults", path)
			return nil
		}
		return fmt.Errorf("failed to read mcp config: %w", err)
	}

	configs, err := LoadMCPConfigs(data)
	if err != nil {
		return fmt.Errorf("failed to parse mcp config: %w", err)
	}

	m.configs = configs

	// Register all server configs in the registry
	for _, config := range configs {
		if config.Enabled {
			m.registry.RegisterServer(config)
		}
	}

	log.Printf("[mcp-manager] loaded %d MCP server configs from %s", len(configs), path)
	return nil
}

// SaveConfigs saves MCP server configurations to a JSON file.
func (m *Manager) SaveConfigs(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := SaveMCPConfigs(m.configs)
	if err != nil {
		return fmt.Errorf("failed to serialize configs: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Printf("[mcp-manager] saved %d MCP server configs to %s", len(m.configs), path)
	return nil
}

// GetConfigs returns all loaded configurations.
func (m *Manager) GetConfigs() []MCPConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]MCPConfig, len(m.configs))
	copy(result, m.configs)
	return result
}

// AddConfig adds a new server configuration without connecting.
func (m *Manager) AddConfig(config MCPConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range m.configs {
		if c.Name == config.Name {
			return fmt.Errorf("server config already exists: %s", config.Name)
		}
	}

	m.configs = append(m.configs, config)
	m.registry.RegisterServer(config)
	return nil
}

// RemoveConfig removes a server configuration.
func (m *Manager) RemoveConfig(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Disconnect first if connected
	if m.client.IsConnected(name) {
		m.client.Disconnect(name)
	}

	delete(m.servers, name)
	m.registry.RemoveServer(name)

	for i, c := range m.configs {
		if c.Name == name {
			m.configs = append(m.configs[:i], m.configs[i+1:]...)
			break
		}
	}

	return nil
}

// ---- Initialization ----

// InitializeDefaultMocks connects the default mock servers (SUMO, MATLAB, VISSIM).
func (m *Manager) InitializeDefaultMocks() error {
	log.Println("[mcp-manager] initializing default mock servers...")

	// SUMO
	sumoServer := CreateMockSUMOServer(m.registry)
	if err := m.ConnectWithServer(sumoServer); err != nil {
		log.Printf("[mcp-manager] warning: failed to initialize SUMO: %v", err)
	}

	// MATLAB
	matlabServer := CreateMockMATLABServer(m.registry)
	if err := m.ConnectWithServer(matlabServer); err != nil {
		log.Printf("[mcp-manager] warning: failed to initialize MATLAB: %v", err)
	}

	// VISSIM
	vissimServer := CreateMockVISSIMServer(m.registry)
	if err := m.ConnectWithServer(vissimServer); err != nil {
		log.Printf("[mcp-manager] warning: failed to initialize VISSIM: %v", err)
	}

	log.Println("[mcp-manager] default mock servers initialized")
	return nil
}

// GetAgentMCPTools returns all MCP tools as agent-compatible Tool instances.
// These can be registered with the agent's tool registry.
func (m *Manager) GetAgentMCPTools() []*AgentMCPTool {
	tools := m.registry.ListTools()
	var agentTools []*AgentMCPTool

	for _, tool := range tools {
		agentTools = append(agentTools, &AgentMCPTool{
			MCPTool: tool,
			CallFunc: func(ctx context.Context, serverName, toolName string, input map[string]interface{}) (map[string]interface{}, error) {
				resp, err := m.client.CallTool(ctx, serverName, toolName, input)
				if err != nil {
					return nil, err
				}
				if !resp.Success {
					return nil, fmt.Errorf(resp.Error)
				}
				return resp.Output, nil
			},
		})
	}

	return agentTools
}

// GetStatus returns a summary of the MCP system status.
func (m *Manager) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	connected := m.client.ListConnectedServers()
	serverDetails := make([]map[string]interface{}, 0)

	for _, name := range connected {
		config, _ := m.registry.GetServer(name)
		tools := m.registry.ListToolsByServer(name)
		serverDetails = append(serverDetails, map[string]interface{}{
			"name":        name,
			"type":        config.Type,
			"description": config.Description,
			"transport":   config.Transport,
			"tool_count":  len(tools),
		})
	}

	return map[string]interface{}{
		"connected_servers": connected,
		"total_servers":     len(m.servers),
		"total_tools":       len(m.registry.ListTools()),
		"servers":           serverDetails,
	}
}

// ExportMCPConfigJSON exports the current configs as JSON bytes.
func (m *Manager) ExportMCPConfigJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(MCPConfigsFile{Servers: m.configs}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal configs: %w", err)
	}
	return data, nil
}