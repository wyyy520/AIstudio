package plugin

import (
	"fmt"
	"log"
	"sync"

	"github.com/aistudio/backend/internal/workflow"
)

// Registry maintains the list of registered plugins and bridges
// plugin node declarations with the workflow engine's NodeRegistry.
type Registry struct {
	mu          sync.RWMutex
	plugins     map[string]*Plugin
	workflowReg *workflow.NodeRegistry
}

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*Plugin),
	}
}

// SetWorkflowRegistry connects the plugin registry to the workflow engine's node registry.
// Also registers nodes from any already-registered plugins.
func (r *Registry) SetWorkflowRegistry(wr *workflow.NodeRegistry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workflowReg = wr

	// Register nodes from already-registered plugins with the new workflow registry
	if wr != nil {
		for name, p := range r.plugins {
			for _, node := range p.Nodes {
				wr.Register(workflow.NodeDefinition{
					Type:        workflow.NodeType(node.Type),
					Plugin:      name,
					Name:        node.Name,
					Description: node.Description,
					Inputs:      convertPortsToWorkflow(node.Inputs),
					Outputs:     convertPortsToWorkflow(node.Outputs),
				})
				log.Printf("[registry] registered node %s from plugin %s (backfill)", node.Type, name)
			}
		}
	}
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
				Type:        workflow.NodeType(node.Type),
				Plugin:      p.Name,
				Name:        node.Name,
				Description: node.Description,
				Inputs:      convertPortsToWorkflow(node.Inputs),
				Outputs:     convertPortsToWorkflow(node.Outputs),
			})
			log.Printf("[registry] registered workflow node %s from plugin %s", node.Type, p.Name)
		}
	}

	return nil
}

// Unregister removes a plugin from the registry.
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

// ListEnabled returns only enabled plugins.
func (r *Registry) ListEnabled() []*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Plugin, 0)
	for _, p := range r.plugins {
		if p.Enabled {
			result = append(result, p)
		}
	}
	return result
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
	if enabled {
		p.Status = StatusEnabled
	} else {
		p.Status = StatusDisabled
	}
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
			ID:          p.ID,
			Name:        p.Name,
			Type:        workflow.DataType(p.Type),
			Description: "", // not in v2 PortInfo yet
			Required:    p.Required,
		})
	}
	return result
}