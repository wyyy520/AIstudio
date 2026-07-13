package mcp

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Client is the MCP client that connects to MCP servers and invokes tools.
// In the first version, servers are in-process (local or mock).
// Future versions will support remote connections via TCP/HTTP.
type Client struct {
	mu       sync.RWMutex
	servers  map[string]*Server
	registry *Registry
}

// NewClient creates a new MCP client.
func NewClient(registry *Registry) *Client {
	return &Client{
		servers:  make(map[string]*Server),
		registry: registry,
	}
}

// Connect establishes a connection to an MCP server.
// In the first version, this registers a local/mock server instance.
func (c *Client) Connect(config MCPConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.servers[config.Name]; exists {
		return fmt.Errorf("already connected to server: %s", config.Name)
	}

	// Register the server in the registry
	if err := c.registry.RegisterServer(config); err != nil {
		return fmt.Errorf("failed to register server: %w", err)
	}

	// Create a local server instance (for mock/process transports)
	server := NewServer(config, c.registry)
	c.servers[config.Name] = server

	log.Printf("[mcp-client] connected to server: %s (transport=%s)", config.Name, config.Transport)
	return nil
}

// ConnectWithServer connects to a server with pre-registered tool handlers.
func (c *Client) ConnectWithServer(server *Server) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	config := server.Config()
	if _, exists := c.servers[config.Name]; exists {
		return fmt.Errorf("already connected to server: %s", config.Name)
	}

	// Register the server config
	if err := c.registry.RegisterServer(config); err != nil {
		return fmt.Errorf("failed to register server: %w", err)
	}

	// Register tools from the server
	for name := range server.tools {
		if tool, ok := c.registry.GetTool(config.Name, name); !ok {
			// Register the tool if not already registered
			c.registry.RegisterTool(MCPTool{
				Name:        name,
				Description: "Tool from " + config.Name,
				ServerName:  config.Name,
				IsMock:      config.Transport == TransportMock,
			})
		} else {
			_ = tool
		}
	}

	c.servers[config.Name] = server
	log.Printf("[mcp-client] connected with server: %s (transport=%s)", config.Name, config.Transport)
	return nil
}

// Disconnect removes a connection to an MCP server.
func (c *Client) Disconnect(serverName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.servers[serverName]; !exists {
		return fmt.Errorf("not connected to server: %s", serverName)
	}

	delete(c.servers, serverName)

	// Also remove from registry
	if err := c.registry.RemoveServer(serverName); err != nil {
		log.Printf("[mcp-client] warning: failed to remove server from registry: %v", err)
	}

	log.Printf("[mcp-client] disconnected from server: %s", serverName)
	return nil
}

// CallTool invokes a specific tool on a connected server.
func (c *Client) CallTool(ctx context.Context, serverName, toolName string, input map[string]interface{}) (*CallResponse, error) {
	c.mu.RLock()
	server, ok := c.servers[serverName]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("server not connected: %s", serverName)
	}

	start := time.Now()

	output, err := server.CallTool(ctx, toolName, input)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return &CallResponse{
			Success:    false,
			Error:      err.Error(),
			DurationMs: duration,
		}, nil
	}

	return &CallResponse{
		Success:    true,
		Output:     output,
		DurationMs: duration,
	}, nil
}

// Call sends a structured CallRequest and returns the response.
func (c *Client) Call(ctx context.Context, req CallRequest) (*CallResponse, error) {
	timeout := req.TimeoutMs
	if timeout <= 0 {
		timeout = 30000 // default 30s
	}

	callCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	return c.CallTool(callCtx, req.ServerName, req.ToolName, req.Input)
}

// IsConnected checks if a server is connected.
func (c *Client) IsConnected(serverName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.servers[serverName]
	return ok
}

// ListConnectedServers returns all connected server names.
func (c *Client) ListConnectedServers() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.servers))
	for name := range c.servers {
		names = append(names, name)
	}
	return names
}

// GetServerConnection returns the server instance for a connected server.
func (c *Client) GetServerConnection(serverName string) (*Server, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	server, ok := c.servers[serverName]
	return server, ok
}

// DisconnectAll disconnects from all servers.
func (c *Client) DisconnectAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for name := range c.servers {
		c.registry.RemoveServer(name)
	}
	c.servers = make(map[string]*Server)
	log.Println("[mcp-client] disconnected from all servers")
}