package plugin

import (
	"fmt"
	"sync"
)

// PluginManager is the central plugin management component.
// It handles pure declaration — no execution, no installation.
type PluginManager struct {
	registry *Registry

	mu sync.RWMutex
}

// NewPluginManager creates a new plugin manager.
func NewPluginManager() *PluginManager {
	return &PluginManager{
		registry: NewRegistry(),
	}
}

// Registry returns the underlying plugin registry.
func (m *PluginManager) Registry() *Registry {
	return m.registry
}

// LoadPlugins discovers and registers all plugins from a directory.
func (m *PluginManager) LoadPlugins(pluginsDir string) error {
	plugins, err := DiscoverPlugins(pluginsDir)
	if err != nil {
		return fmt.Errorf("plugin discovery failed: %w", err)
	}

	for _, p := range plugins {
		if err := m.registry.Register(p); err != nil {
			continue
		}
	}

	return nil
}

// Refresh re-discovers all plugins from the given directory.
func (m *PluginManager) Refresh(pluginsDir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.registry = NewRegistry()
	return m.LoadPlugins(pluginsDir)
}

// List returns summaries of all registered plugins.
func (m *PluginManager) List() []PluginSummary {
	return m.registry.ListSummaries()
}

// Get returns a plugin by name.
func (m *PluginManager) Get(name string) (*Plugin, error) {
	p, ok := m.registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return p, nil
}

// GetByID returns a plugin by ID.
func (m *PluginManager) GetByID(id string) (*Plugin, error) {
	p, ok := m.registry.GetByID(id)
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", id)
	}
	return p, nil
}

// Enable enables a plugin.
func (m *PluginManager) Enable(name string) error {
	return m.registry.UpdateEnabled(name, true)
}

// Disable disables a plugin.
func (m *PluginManager) Disable(name string) error {
	return m.registry.UpdateEnabled(name, false)
}

// ListEnabled returns only enabled plugins.
func (m *PluginManager) ListEnabled() []*Plugin {
	return m.registry.ListEnabled()
}

// ListNodeTypes returns all node types from enabled plugins.
func (m *PluginManager) ListNodeTypes() []PluginNode {
	var nodes []PluginNode
	for _, p := range m.registry.List() {
		if p.Enabled {
			nodes = append(nodes, p.Nodes...)
		}
	}
	return nodes
}

// PluginCount returns the total number of registered plugins.
func (m *PluginManager) PluginCount() int {
	return m.registry.Count()
}
