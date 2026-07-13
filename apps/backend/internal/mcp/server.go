package mcp

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Server represents an MCP server that exposes tools.
// In the first version, this is a local in-process server.
// Future versions will support the standard MCP protocol over stdio/HTTP.
type Server struct {
	config   MCPConfig
	tools    map[string]ToolHandler
	registry *Registry
}

// ToolHandler is a function that handles an MCP tool call.
type ToolHandler func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

// NewServer creates a new MCP server instance.
func NewServer(config MCPConfig, registry *Registry) *Server {
	return &Server{
		config:   config,
		tools:    make(map[string]ToolHandler),
		registry: registry,
	}
}

// RegisterHandler registers a tool handler for this server.
func (s *Server) RegisterHandler(toolName string, handler ToolHandler) {
	s.tools[toolName] = handler
}

// CallTool invokes a tool on this server by name.
func (s *Server) CallTool(ctx context.Context, toolName string, input map[string]interface{}) (map[string]interface{}, error) {
	handler, ok := s.tools[toolName]
	if !ok {
		return nil, fmt.Errorf("tool %q not found on server %q", toolName, s.config.Name)
	}

	log.Printf("[mcp-server] %s.%s called with input: %v", s.config.Name, toolName, input)

	// Add timeout context if configured
	if s.config.TimeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(s.config.TimeoutMs)*time.Millisecond)
		defer cancel()
	}

	output, err := handler(ctx, input)
	if err != nil {
		log.Printf("[mcp-server] %s.%s error: %v", s.config.Name, toolName, err)
		return nil, err
	}

	log.Printf("[mcp-server] %s.%s completed", s.config.Name, toolName)
	return output, nil
}

// ListTools returns all tools registered on this server.
func (s *Server) ListTools() []MCPTool {
	var tools []MCPTool
	for name := range s.tools {
		if tool, ok := s.registry.GetTool(s.config.Name, name); ok {
			tools = append(tools, tool)
		}
	}
	return tools
}

// Config returns the server configuration.
func (s *Server) Config() MCPConfig {
	return s.config
}

