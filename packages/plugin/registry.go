package plugin

import (
	"fmt"
	"sync"
)

// Registry maintains the list of registered plugins and provides
// node type lookup capabilities.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin
}

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*Plugin),
	}
}

// Register adds a plugin to the registry.
func (r *Registry) Register(p *Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[p.Name]; exists {
		return fmt.Errorf("plugin already registered: %s", p.Name)
	}

	r.plugins[p.Name] = p
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

// RegisterNodeType registers a node type from a plugin manifest.
func (r *Registry) RegisterNodeType(manifest *ManifestV2, node PluginNode) error {
	return nil
}

// GetNodeType returns a PluginNode by type name across all registered plugins.
func (r *Registry) GetNodeType(typeName string) (PluginNode, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.plugins {
		for _, n := range p.Nodes {
			if n.Type == typeName {
				return n, true
			}
		}
	}
	return PluginNode{}, false
}

// ListNodeTypes returns all node types from all registered plugins.
func (r *Registry) ListNodeTypes() []PluginNode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var nodes []PluginNode
	for _, p := range r.plugins {
		nodes = append(nodes, p.Nodes...)
	}
	return nodes
}

// GetNodesByTarget returns all plugin nodes that support a given target.
func (r *Registry) GetNodesByTarget(target string) []PluginNode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var nodes []PluginNode
	for _, p := range r.plugins {
		if p.Manifest == nil {
			continue
		}
		for _, t := range p.Manifest.SupportedTargets {
			if t == target {
				nodes = append(nodes, p.Nodes...)
				break
			}
		}
	}
	return nodes
}
