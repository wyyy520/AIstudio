package mcp

import (
	"fmt"
	"sync"
)

// Registry manages all MCP tool registrations.
// It stores tool definitions and server configurations.
type Registry struct {
	mu      sync.RWMutex
	tools   map[string]MCPTool   // key: server_name.tool_name
	servers map[string]MCPConfig // key: server_name
}

// NewRegistry creates a new MCP tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools:   make(map[string]MCPTool),
		servers: make(map[string]MCPConfig),
	}
}

func toolKey(serverName, toolName string) string {
	return serverName + "." + toolName
}

// RegisterTool registers a single MCP tool.
func (r *Registry) RegisterTool(tool MCPTool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := toolKey(tool.ServerName, tool.Name)
	if _, exists := r.tools[key]; exists {
		return fmt.Errorf("tool already registered: %s", key)
	}

	r.tools[key] = tool
	return nil
}

// RemoveTool removes a registered tool.
func (r *Registry) RemoveTool(serverName, toolName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := toolKey(serverName, toolName)
	if _, exists := r.tools[key]; !exists {
		return fmt.Errorf("tool not found: %s", key)
	}

	delete(r.tools, key)
	return nil
}

// GetTool retrieves a tool by server and tool name.
func (r *Registry) GetTool(serverName, toolName string) (MCPTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[toolKey(serverName, toolName)]
	return tool, ok
}

// ListTools returns all registered tools.
func (r *Registry) ListTools() []MCPTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]MCPTool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

// ListToolsByServer returns tools for a specific server.
func (r *Registry) ListToolsByServer(serverName string) []MCPTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []MCPTool
	for _, t := range r.tools {
		if t.ServerName == serverName {
			result = append(result, t)
		}
	}
	return result
}

// RegisterServer stores a server configuration.
func (r *Registry) RegisterServer(config MCPConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.servers[config.Name]; exists {
		return fmt.Errorf("server already registered: %s", config.Name)
	}

	r.servers[config.Name] = config
	return nil
}

// RemoveServer removes a server configuration and all its tools.
func (r *Registry) RemoveServer(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.servers[name]; !exists {
		return fmt.Errorf("server not found: %s", name)
	}

	delete(r.servers, name)

	// Remove all tools for this server
	for key := range r.tools {
		serverName := key[:len(name)]
		if serverName == name && len(key) > len(name) && key[len(name)] == '.' {
			delete(r.tools, key)
		}
	}

	return nil
}

// GetServer retrieves a server configuration.
func (r *Registry) GetServer(name string) (MCPConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, ok := r.servers[name]
	return config, ok
}

// ListServers returns all registered server configurations.
func (r *Registry) ListServers() []MCPConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]MCPConfig, 0, len(r.servers))
	for _, c := range r.servers {
		result = append(result, c)
	}
	return result
}

// RegisterServerWithTools registers a server and all its tools at once.
func (r *Registry) RegisterServerWithTools(config MCPConfig, tools []MCPTool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.servers[config.Name]; exists {
		return fmt.Errorf("server already registered: %s", config.Name)
	}

	r.servers[config.Name] = config
	for _, tool := range tools {
		key := toolKey(tool.ServerName, tool.Name)
		r.tools[key] = tool
	}

	return nil
}