package plugin

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Manager is the central plugin management component.
// It coordinates discovery, registration, and execution.
type Manager struct {
	registry *Registry
	loader   *FileLoader
	executor *SimpleExecutor
	mu       sync.RWMutex
}

// NewManager creates a new plugin manager.
func NewManager(pluginsDir string) *Manager {
	registry := NewRegistry()
	loader := NewFileLoader(pluginsDir)
	executor := NewSimpleExecutor(registry)

	return &Manager{
		registry: registry,
		loader:   loader,
		executor: executor,
	}
}

// DiscoverPlugins scans the plugins directory and registers all found plugins.
func (m *Manager) DiscoverPlugins() error {
	log.Println("[plugin-manager] discovering plugins...")

	plugins, err := m.loader.ScanDir()
	if err != nil {
		return fmt.Errorf("plugin discovery failed: %w", err)
	}

	for _, p := range plugins {
		if err := m.registry.Register(p); err != nil {
			log.Printf("[plugin-manager] warning: failed to register plugin %s: %v", p.Name, err)
			continue
		}
		log.Printf("[plugin-manager] discovered plugin: %s@%s (%s)", p.Name, p.Version, p.Path)
	}

	log.Printf("[plugin-manager] discovery complete: %d plugins found", len(plugins))
	return nil
}

// GetPlugin returns a plugin by name.
func (m *Manager) GetPlugin(name string) (*Plugin, error) {
	p, ok := m.registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return p, nil
}

// ListPlugins returns all registered plugins.
func (m *Manager) ListPlugins() []*Plugin {
	return m.registry.List()
}

// EnablePlugin enables a plugin.
func (m *Manager) EnablePlugin(name string) error {
	return m.executor.SetPluginStatus(name, true)
}

// DisablePlugin disables a plugin.
func (m *Manager) DisablePlugin(name string) error {
	return m.executor.SetPluginStatus(name, false)
}

// ExecutePlugin runs a plugin with the given input.
func (m *Manager) ExecutePlugin(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error) {
	return m.executor.Execute(ctx, name, input)
}

// RegisterPluginFromDir loads and registers a plugin from a specific directory.
func (m *Manager) RegisterPluginFromDir(dirPath string) error {
	plugin, err := m.loader.LoadPlugin(dirPath)
	if err != nil {
		return err
	}
	plugin.CreatedAt = time.Now()
	plugin.UpdatedAt = time.Now()

	return m.registry.Register(plugin)
}

// InstallPlugin installs a plugin by scanning the plugins directory.
// This triggers a re-discovery of plugins.
func (m *Manager) InstallPlugin(name string) error {
	// Re-discover all plugins to pick up new ones
	plugins, err := m.loader.ScanDir()
	if err != nil {
		return fmt.Errorf("install plugin: %w", err)
	}

	for _, p := range plugins {
		if p.Name == name {
			if err := m.registry.Register(p); err != nil {
				return fmt.Errorf("install plugin %s: %w", name, err)
			}
			log.Printf("[plugin-manager] installed plugin: %s@%s", p.Name, p.Version)
			return nil
		}
	}

	return fmt.Errorf("plugin not found in directory: %s", name)
}

// RemovePlugin unregisters a plugin.
func (m *Manager) RemovePlugin(name string) error {
	return m.registry.Unregister(name)
}

// PluginCount returns the total number of registered plugins.
func (m *Manager) PluginCount() int {
	return m.registry.Count()
}