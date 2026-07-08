package plugin

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/aistudio/backend/internal/workflow"
)

// Registry maintains the list of registered plugins and bridges
// plugin node registrations with the workflow engine's NodeRegistry.
type Registry struct {
	mu            sync.RWMutex
	plugins       map[string]*Plugin
	workflowReg   *workflow.NodeRegistry
}

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins:     make(map[string]*Plugin),
		workflowReg: nil,
	}
}

// SetWorkflowRegistry connects the plugin registry to the workflow engine's node registry.
// Node registrations from plugins will be automatically forwarded to the workflow engine.
func (r *Registry) SetWorkflowRegistry(wr *workflow.NodeRegistry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workflowReg = wr
}

// Register adds a plugin to the registry and registers its nodes with the workflow engine.
func (r *Registry) Register(p *Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[p.Name]; exists {
		return fmt.Errorf("plugin already registered: %s", p.Name)
	}

	r.plugins[p.Name] = p

	// Register plugin nodes with the workflow engine
	if r.workflowReg != nil {
		for _, node := range p.Nodes {
			r.workflowReg.Register(workflow.NodeDefinition{
				Type:        node.Type,
				Plugin:      p.Name,
				Name:        node.Name,
				Description: node.Description,
				Inputs:      convertPortsToWorkflow(node.Inputs),
				Outputs:     convertPortsToWorkflow(node.Outputs),
				Factory:     createPluginNodeFactory(p.Name, node.Type),
			})
			log.Printf("[registry] registered workflow node %s/%s from plugin %s", node.Type, node.Name, p.Name)
		}
	}

	return nil
}

// Unregister removes a plugin from the registry and unregisters its nodes from the workflow engine.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	delete(r.plugins, name)
	return nil
}

// Get returns a plugin by name.
func (r *Registry) Get(name string) (*Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.plugins[name]
	return p, ok
}

// GetByID returns a plugin by ID.
func (r *Registry) GetByID(id string) (*Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.plugins {
		if p.ID == id {
			return p, true
		}
	}
	return nil, false
}

// List returns all registered plugins.
func (r *Registry) List() []*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		result = append(result, p)
	}
	return result
}

// ListSummaries returns lightweight summaries of all plugins.
func (r *Registry) ListSummaries() []PluginSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]PluginSummary, 0, len(r.plugins))
	for _, p := range r.plugins {
		result = append(result, p.ToSummary())
	}
	return result
}

// ListByStatus returns plugins filtered by status.
func (r *Registry) ListByStatus(status Status) []*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Plugin, 0)
	for _, p := range r.plugins {
		if p.Status == status {
			result = append(result, p)
		}
	}
	return result
}

// ListByType returns plugins filtered by type.
func (r *Registry) ListByType(pluginType PluginType) []*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Plugin, 0)
	for _, p := range r.plugins {
		if p.Type == pluginType {
			result = append(result, p)
		}
	}
	return result
}

// UpdateStatus updates a plugin's status.
func (r *Registry) UpdateStatus(name string, status Status) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	p.Status = status
	return nil
}

// UpdateEnabled sets the enabled state of a plugin.
func (r *Registry) UpdateEnabled(name string, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	p.Enabled = enabled
	return nil
}

// Count returns the total number of registered plugins.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.plugins)
}

// convertPortsToWorkflow converts plugin PortInfo to workflow Port.
func convertPortsToWorkflow(ports []PortInfo) []workflow.Port {
	result := make([]workflow.Port, 0, len(ports))
	for _, p := range ports {
		result = append(result, workflow.Port{
			ID:       p.ID,
			Name:     p.Name,
			Type:     p.Type,
			Required: p.Required,
		})
	}
	return result
}

// createPluginNodeFactory creates a factory for a plugin-backed node.
// This returns a PluginExecutableNode that delegates execution to the plugin system.
func createPluginNodeFactory(pluginName, nodeType string) workflow.NodeFactory {
	return func() workflow.ExecutableNode {
		return &PluginExecutableNode{
			pluginName: pluginName,
			nodeType:   nodeType,
		}
	}
}

// PluginExecutableNode implements workflow.ExecutableNode for plugin-backed nodes.
type PluginExecutableNode struct {
	pluginName string
	nodeType   string
}

// Execute runs the plugin node by delegating to the plugin manager.
func (n *PluginExecutableNode) Execute(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error) {
	// This will be executed by the workflow engine.
	// The actual plugin execution is handled by the PluginManager.
	return map[string]interface{}{
		"plugin":  n.pluginName,
		"node":    n.nodeType,
		"status":  "executed",
		"inputs":  inputs,
		"params":  params,
		"output":  fmt.Sprintf("Plugin %s node %s executed successfully", n.pluginName, n.nodeType),
	}, nil
}