package plugin

import (
	"context"
	"fmt"
	"sync"
)

// PythonEngineRunner is the interface for running Python engine tasks.
// It is implemented by the engine package to avoid circular imports.
type PythonEngineRunner interface {
	RunPluginAction(ctx context.Context, taskID, plugin, action string, params map[string]interface{}) (map[string]interface{}, error)
}

// SimpleExecutor provides a basic plugin execution framework.
// It dispatches to the Python Engine for Python-based plugins.
type SimpleExecutor struct {
	registry      *Registry
	engineRunner  PythonEngineRunner
	mu            sync.RWMutex
}

// NewSimpleExecutor creates a new SimpleExecutor.
func NewSimpleExecutor(registry *Registry) *SimpleExecutor {
	return &SimpleExecutor{
		registry: registry,
	}
}

// SetEngineRunner sets the Python engine runner for real execution.
func (e *SimpleExecutor) SetEngineRunner(runner PythonEngineRunner) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.engineRunner = runner
}

// Execute runs a plugin by name with the given input.
func (e *SimpleExecutor) Execute(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error) {
	p, ok := e.registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	if !p.Enabled {
		return nil, fmt.Errorf("plugin %s is disabled", name)
	}

	// Dispatch to Python Engine if available
	e.mu.RLock()
	runner := e.engineRunner
	e.mu.RUnlock()

	if runner != nil {
		action := "execute"
		taskID := fmt.Sprintf("plugin-%s-%d", name, ctx.Value("request_id"))
		if taskID == "" {
			taskID = fmt.Sprintf("plugin-%s", name)
		}

		result, err := runner.RunPluginAction(ctx, taskID, name, action, input)
		if err != nil {
			return nil, fmt.Errorf("plugin %s execution failed: %w", name, err)
		}
		return result, nil
	}

	// Fallback: mock execution
	result := map[string]interface{}{
		"plugin":     name,
		"version":    p.Version,
		"type":       string(p.Type),
		"status":     "executed",
		"input":      input,
		"node_count": len(p.Nodes),
		"output":     fmt.Sprintf("Plugin %s executed successfully (mock)", name),
	}

	return result, nil
}

// ExecuteNode runs a specific node from a plugin.
func (e *SimpleExecutor) ExecuteNode(ctx context.Context, pluginName, nodeType string, input map[string]interface{}) (map[string]interface{}, error) {
	p, ok := e.registry.Get(pluginName)
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", pluginName)
	}

	if !p.Enabled {
		return nil, fmt.Errorf("plugin %s is disabled", pluginName)
	}

	// Find the node type
	var nodeReg *NodeRegistration
	for i := range p.Nodes {
		if p.Nodes[i].Type == nodeType {
			nodeReg = &p.Nodes[i]
			break
		}
	}

	if nodeReg == nil {
		return nil, fmt.Errorf("node type %s not found in plugin %s", nodeType, pluginName)
	}

	// Dispatch to Python Engine if available
	e.mu.RLock()
	runner := e.engineRunner
	e.mu.RUnlock()

	if runner != nil {
		action := nodeType
		taskID := fmt.Sprintf("plugin-%s-%s", pluginName, nodeType)
		result, err := runner.RunPluginAction(ctx, taskID, pluginName, action, input)
		if err != nil {
			return nil, fmt.Errorf("plugin node %s/%s execution failed: %w", pluginName, nodeType, err)
		}
		return result, nil
	}

	// Fallback: mock execution
	result := map[string]interface{}{
		"plugin":  pluginName,
		"node":    nodeType,
		"version": p.Version,
		"status":  "executed",
		"input":   input,
		"output":  fmt.Sprintf("Plugin %s node %s executed successfully (mock)", pluginName, nodeType),
	}

	return result, nil
}

// SetPluginStatus updates a plugin's enabled/disabled status.
func (e *SimpleExecutor) SetPluginStatus(name string, enabled bool) error {
	return e.registry.UpdateEnabled(name, enabled)
}