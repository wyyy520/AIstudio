package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aistudio/backend/internal/workflow"
)

// Manager is the central plugin management component for V2.
type Manager struct {
	registry      *Registry
	pluginsDir    string
	workflowReg   *workflow.NodeRegistry
	pluginEventFn func(event string, data map[string]interface{})
	executors     map[string]PluginExecutor
	installer     *Installer

	mu sync.RWMutex
}

// NewManager creates a new plugin manager for V2.
// It scans the provided Plugins directory for plugin manifests.
func NewManager(pluginsDir string) *Manager {
	registry := NewRegistry()
	return &Manager{
		registry:   registry,
		pluginsDir: pluginsDir,
		executors:  make(map[string]PluginExecutor),
		installer:  NewInstaller(pluginsDir, registry),
	}
}

// SetWorkflowRegistry connects the plugin registry to a workflow engine's node registry.
func (m *Manager) SetWorkflowRegistry(wr *workflow.NodeRegistry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workflowReg = wr
	m.registry.SetWorkflowRegistry(wr)
}

// SetPluginEventCallback sets a callback for plugin lifecycle events.
func (m *Manager) SetPluginEventCallback(fn func(event string, data map[string]interface{})) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pluginEventFn = fn
}

// DiscoverPlugins scans the plugins directory and registers all found plugins.
func (m *Manager) DiscoverPlugins() error {
	log.Printf("[plugin-manager] discovering plugins from %s...", m.pluginsDir)

	plugins, err := discoverFromDir(m.pluginsDir)
	if err != nil {
		return fmt.Errorf("plugin discovery failed: %w", err)
	}

	for _, p := range plugins {
		if err := m.registry.Register(p); err != nil {
			log.Printf("[plugin-manager] warning: failed to register plugin %s: %v", p.Name, err)
			continue
		}
		log.Printf("[plugin-manager] discovered plugin: %s@%s (%s) [%d nodes]",
			p.Name, p.Version, p.Manifest.Kind, len(p.Nodes))
	}

	log.Printf("[plugin-manager] discovery complete: %d plugins found", len(plugins))
	return nil
}

// discoverFromDir recursively scans a directory for plugin.json files.
func discoverFromDir(rootDir string) ([]*Plugin, error) {
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	var plugins []*Plugin
	err = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) != "plugin.json" {
			return nil
		}

		p, err := loadPluginFromManifest(path)
		if err != nil {
			log.Printf("[plugin-manifest] warning: failed to load %s: %v", path, err)
			return nil // skip invalid manifests
		}
		plugins = append(plugins, p)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk plugins dir failed: %w", err)
	}

	return plugins, nil
}

// loadPluginFromManifest reads a plugin.json file and returns a Plugin.
func loadPluginFromManifest(manifestPath string) (*Plugin, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var mv ManifestV2
	if err := json.Unmarshal(data, &mv); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	// Validate required fields
	if mv.ID == "" {
		return nil, fmt.Errorf("manifest %s: id is required", manifestPath)
	}
	if mv.Name == "" {
		return nil, fmt.Errorf("manifest %s: name is required", manifestPath)
	}
	if mv.Version == "" {
		return nil, fmt.Errorf("manifest %s: version is required", manifestPath)
	}
	if len(mv.Nodes) == 0 {
		return nil, fmt.Errorf("manifest %s: at least one node is required", manifestPath)
	}

	// Infer plugin type from kind
	var pluginType PluginType
	switch mv.Kind {
	case ManifestKindAlgorithm:
		pluginType = PluginTypeVision
	case ManifestKindRuntime:
		pluginType = PluginTypeSystem
	case ManifestKindSystem:
		pluginType = PluginTypeSystem
	default:
		pluginType = PluginTypeSystem
	}

	now := time.Now()
	sourceDir := filepath.Dir(manifestPath)

	return &Plugin{
		ID:          mv.ID,
		Name:        mv.Name,
		Version:     mv.Version,
		Author:      mv.Author,
		Type:        pluginType,
		Description: mv.Description,
		Status:      StatusDisabled,
		Enabled:     false,
		Nodes:       mv.Nodes,
		Manifest:    &mv,
		SourceDir:   sourceDir,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
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

// ListEnabled returns only enabled plugins.
func (m *Manager) ListEnabled() []*Plugin {
	return m.registry.ListEnabled()
}

// EnablePlugin enables a plugin.
func (m *Manager) EnablePlugin(name string) error {
	if err := m.registry.UpdateEnabled(name, true); err != nil {
		return err
	}

	m.mu.RLock()
	fn := m.pluginEventFn
	m.mu.RUnlock()

	if fn != nil {
		fn("enabled", map[string]interface{}{"name": name})
	}

	log.Printf("[plugin-manager] enabled plugin: %s", name)
	return nil
}

// DisablePlugin disables a plugin.
func (m *Manager) DisablePlugin(name string) error {
	if err := m.registry.UpdateEnabled(name, false); err != nil {
		return err
	}

	m.mu.RLock()
	fn := m.pluginEventFn
	m.mu.RUnlock()

	if fn != nil {
		fn("disabled", map[string]interface{}{"name": name})
	}

	log.Printf("[plugin-manager] disabled plugin: %s", name)
	return nil
}

// GetNodeTypes returns all workflow node types from enabled plugins.
func (m *Manager) GetNodeTypes() []PluginNode {
	var nodes []PluginNode
	for _, p := range m.registry.List() {
		if p.Enabled {
			nodes = append(nodes, p.Nodes...)
		}
	}
	return nodes
}

// GetRegistry returns the plugin registry.
func (m *Manager) GetRegistry() *Registry {
	return m.registry
}

// PluginCount returns the total number of registered plugins.
func (m *Manager) PluginCount() int {
	return m.registry.Count()
}

// RegisterExecutor registers an executor for a plugin language.
func (m *Manager) RegisterExecutor(language string, executor PluginExecutor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executors[language] = executor
}

// Execute runs a plugin with the given input and config.
func (m *Manager) Execute(ctx context.Context, pluginName string, input map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	p, ok := m.registry.Get(pluginName)
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", pluginName)
	}
	if !p.Enabled {
		return nil, fmt.Errorf("plugin is disabled: %s", pluginName)
	}

	language := "python"
	if p.Manifest != nil && p.Manifest.Language != "" {
		language = p.Manifest.Language
	}

	m.mu.RLock()
	executor, ok := m.executors[language]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no executor registered for language: %s", language)
	}

	return executor.Execute(ctx, p, input, config)
}

// Installer returns the plugin installer.
func (m *Manager) Installer() *Installer {
	return m.installer
}
