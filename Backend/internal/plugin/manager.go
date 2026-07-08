package plugin

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Manager is the central plugin management component.
// It coordinates discovery, installation, registration, execution, and persistence.
type Manager struct {
	registry *Registry
	loader   *FileLoader
	executor *SimpleExecutor
	installer *Installer
	depMgr   *DependencyManager
	repo     *PluginRepository

	mu sync.RWMutex
}

// NewManager creates a new plugin manager.
func NewManager(pluginsDir string) *Manager {
	registry := NewRegistry()
	loader := NewFileLoader(pluginsDir)
	executor := NewSimpleExecutor(registry)
	depMgr := NewDependencyManager()
	installer := NewInstaller(pluginsDir, loader, depMgr, registry)

	return &Manager{
		registry:  registry,
		loader:    loader,
		executor:  executor,
		installer: installer,
		depMgr:    depMgr,
	}
}

// SetWorkflowRegistry connects the plugin registry to a workflow engine's node registry.
func (m *Manager) SetWorkflowRegistry(wr interface{}) {
	// Type-assert to workflow.NodeRegistry through the interface
	type WorkflowRegistry interface {
		Register(def interface{})
	}
	_ = wr
}

// SetRepository wires the database repository for plugin persistence.
func (m *Manager) SetRepository(repo *PluginRepository) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.repo = repo
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

		// Persist to database
		if m.repo != nil {
			if err := m.repo.Save(p); err != nil {
				log.Printf("[plugin-manager] warning: failed to persist plugin %s: %v", p.Name, err)
			}
		}

		log.Printf("[plugin-manager] discovered plugin: %s@%s (%s) [%d nodes]", p.Name, p.Version, p.Type, len(p.Nodes))
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

// GetPluginByID returns a plugin by ID.
func (m *Manager) GetPluginByID(id string) (*Plugin, error) {
	p, ok := m.registry.GetByID(id)
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", id)
	}
	return p, nil
}

// ListPlugins returns all registered plugins.
func (m *Manager) ListPlugins() []*Plugin {
	return m.registry.List()
}

// ListPluginSummaries returns lightweight summaries of all plugins.
func (m *Manager) ListPluginSummaries() []PluginSummary {
	return m.registry.ListSummaries()
}

// ListPluginsByType returns plugins filtered by type.
func (m *Manager) ListPluginsByType(pluginType PluginType) []*Plugin {
	return m.registry.ListByType(pluginType)
}

// ListPluginsByStatus returns plugins filtered by status.
func (m *Manager) ListPluginsByStatus(status Status) []*Plugin {
	return m.registry.ListByStatus(status)
}

// EnablePlugin enables a plugin.
func (m *Manager) EnablePlugin(name string) error {
	plugin, ok := m.registry.Get(name)
	if !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	plugin.Enabled = true
	plugin.UpdatedAt = time.Now()

	if m.repo != nil {
		m.repo.Update(plugin)
	}

	log.Printf("[plugin-manager] enabled plugin: %s", name)
	return nil
}

// DisablePlugin disables a plugin.
func (m *Manager) DisablePlugin(name string) error {
	plugin, ok := m.registry.Get(name)
	if !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	plugin.Enabled = false
	plugin.UpdatedAt = time.Now()

	if m.repo != nil {
		m.repo.Update(plugin)
	}

	log.Printf("[plugin-manager] disabled plugin: %s", name)
	return nil
}

// ExecutePlugin runs a plugin with the given input.
func (m *Manager) ExecutePlugin(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error) {
	return m.executor.Execute(ctx, name, input)
}

// InstallPlugin installs a plugin from a plugin.json manifest path.
func (m *Manager) InstallPlugin(manifestPath string) (*InstallResult, error) {
	result, err := m.installer.InstallPlugin(manifestPath)
	if err != nil {
		return result, err
	}

	// Persist to database
	if m.repo != nil && result.Success && result.Plugin != nil {
		if err := m.repo.Save(result.Plugin); err != nil {
			log.Printf("[plugin-manager] warning: failed to persist plugin %s: %v", result.Plugin.Name, err)
		}
	}

	return result, nil
}

// InstallPluginFromURL installs a plugin from a remote source (mock).
func (m *Manager) InstallPluginFromURL(name, url string) (*InstallResult, error) {
	result, err := m.installer.InstallFromURL(name, url)
	if err != nil {
		return result, err
	}

	// Persist to database
	if m.repo != nil && result.Success && result.Plugin != nil {
		if err := m.repo.Save(result.Plugin); err != nil {
			log.Printf("[plugin-manager] warning: failed to persist plugin %s: %v", result.Plugin.Name, err)
		}
	}

	return result, nil
}

// RemovePlugin unregisters and removes a plugin.
func (m *Manager) RemovePlugin(name string) error {
	if err := m.installer.RemovePlugin(name); err != nil {
		return err
	}

	// Remove from database
	if m.repo != nil {
		if err := m.repo.Delete(name); err != nil {
			log.Printf("[plugin-manager] warning: failed to delete plugin %s from db: %v", name, err)
		}
	}

	return nil
}

// UpdatePlugin updates a plugin to its latest version.
func (m *Manager) UpdatePlugin(name string) (*InstallResult, error) {
	result, err := m.installer.UpdatePlugin(name)
	if err != nil {
		return result, err
	}

	// Update database
	if m.repo != nil && result.Success && result.Plugin != nil {
		if err := m.repo.Update(result.Plugin); err != nil {
			log.Printf("[plugin-manager] warning: failed to update plugin %s in db: %v", result.Plugin.Name, err)
		}
	}

	return result, nil
}

// CheckDependencies checks dependencies for a plugin by name.
func (m *Manager) CheckDependencies(name string) ([]DependencyCheckResult, bool, error) {
	plugin, ok := m.registry.Get(name)
	if !ok {
		return nil, false, fmt.Errorf("plugin not found: %s", name)
	}
	results, allOK := m.depMgr.CheckDependencies(plugin.Dependencies)
	return results, allOK, nil
}

// GetDependencyManager returns the dependency manager for external use.
func (m *Manager) GetDependencyManager() *DependencyManager {
	return m.depMgr
}

// GetRegistry returns the plugin registry.
func (m *Manager) GetRegistry() *Registry {
	return m.registry
}

// GetExecutor returns the plugin executor for wiring to the Python engine.
func (m *Manager) GetExecutor() *SimpleExecutor {
	return m.executor
}

// PluginCount returns the total number of registered plugins.
func (m *Manager) PluginCount() int {
	return m.registry.Count()
}

// GetNodeTypes returns all workflow node types provided by registered plugins.
func (m *Manager) GetNodeTypes() []NodeRegistration {
	plugins := m.registry.List()
	var nodes []NodeRegistration
	for _, p := range plugins {
		if p.Enabled {
			nodes = append(nodes, p.Nodes...)
		}
	}
	return nodes
}