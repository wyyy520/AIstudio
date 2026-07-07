package plugin

import (
	"context"
	"fmt"
	"sync"
)

// SimpleExecutor provides a basic plugin execution framework.
// For now, it looks up the plugin by name and returns a placeholder.
// Future: integrate with MCP Tool Calling, Python Engine, etc.
type SimpleExecutor struct {
	registry *Registry
	mu       sync.RWMutex
}

// NewSimpleExecutor creates a new SimpleExecutor.
func NewSimpleExecutor(registry *Registry) *SimpleExecutor {
	return &SimpleExecutor{
		registry: registry,
	}
}

// Execute runs a plugin by name with the given input.
// Currently returns a placeholder result.
// In the future, this will invoke:
//   - MCP Tool Calling for MCP plugins
//   - Python Engine for Python-based plugins
//   - Direct Go execution for native plugins
func (e *SimpleExecutor) Execute(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error) {
	plugin, ok := e.registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	if plugin.Status != StatusEnabled && plugin.Status != StatusInstalled {
		return nil, fmt.Errorf("plugin %s is not enabled (status: %s)", name, plugin.Status)
	}

	// TODO: In future iterations, dispatch to:
	//   - MCP runtime if plugin is MCP-based
	//   - Python Engine if plugin has .py entry
	//   - Native Go executor if plugin is compiled-in

	result := map[string]interface{}{
		"plugin":  name,
		"version": plugin.Version,
		"status":  "executed",
		"input":   input,
		"output":  "Plugin execution not yet implemented. This is a placeholder.",
	}

	return result, nil
}

// SetPluginStatus updates a plugin's enabled/disabled status.
func (e *SimpleExecutor) SetPluginStatus(name string, enabled bool) error {
	status := StatusEnabled
	if !enabled {
		status = StatusDisabled
	}
	return e.registry.UpdateStatus(name, status)
}