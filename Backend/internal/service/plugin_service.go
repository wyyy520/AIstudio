package service

import (
	"context"
	"fmt"

	"github.com/aistudio/backend/internal/plugin"
	"gorm.io/gorm"
)

// PluginService handles plugin business logic.
type PluginService struct {
	db      *gorm.DB
	manager *plugin.Manager
}

// NewPluginService creates a new PluginService.
func NewPluginService(db *gorm.DB, manager *plugin.Manager) *PluginService {
	return &PluginService{db: db, manager: manager}
}

// List returns all registered plugins as summaries.
func (s *PluginService) List() []plugin.PluginSummary {
	return s.manager.ListPluginSummaries()
}

// ListFull returns all registered plugins with full details.
func (s *PluginService) ListFull() []*plugin.Plugin {
	return s.manager.ListPlugins()
}

// Get returns a single plugin by name.
func (s *PluginService) Get(name string) (*plugin.Plugin, error) {
	return s.manager.GetPlugin(name)
}

// GetByID returns a single plugin by ID.
func (s *PluginService) GetByID(id string) (*plugin.Plugin, error) {
	return s.manager.GetPluginByID(id)
}

// Install installs a new plugin from its manifest path.
func (s *PluginService) Install(manifestPath string) (*plugin.InstallResult, error) {
	return s.manager.InstallPlugin(manifestPath)
}

// InstallFromURL installs a plugin from a remote source (mock).
func (s *PluginService) InstallFromURL(name, url string) (*plugin.InstallResult, error) {
	return s.manager.InstallPluginFromURL(name, url)
}

// Remove removes a plugin by name.
func (s *PluginService) Remove(name string) error {
	return s.manager.RemovePlugin(name)
}

// Update updates a plugin to its latest version.
func (s *PluginService) Update(name string) (*plugin.InstallResult, error) {
	return s.manager.UpdatePlugin(name)
}

// UpdateStatus enables or disables a plugin.
func (s *PluginService) UpdateStatus(name, status string) error {
	switch status {
	case "enabled":
		return s.manager.EnablePlugin(name)
	case "disabled":
		return s.manager.DisablePlugin(name)
	default:
		return fmt.Errorf("invalid plugin status: %s (allowed: enabled, disabled)", status)
	}
}

// Uninstall removes a plugin by name.
func (s *PluginService) Uninstall(name string) error {
	return s.manager.RemovePlugin(name)
}

// Execute runs a plugin with the given input.
func (s *PluginService) Execute(ctx context.Context, name string, input map[string]interface{}) (interface{}, error) {
	return s.manager.ExecutePlugin(ctx, name, input)
}

// CheckDependencies checks plugin dependencies.
func (s *PluginService) CheckDependencies(name string) ([]plugin.DependencyCheckResult, bool, error) {
	return s.manager.CheckDependencies(name)
}

// GetNodeTypes returns all workflow node types from registered plugins.
func (s *PluginService) GetNodeTypes() []plugin.NodeRegistration {
	return s.manager.GetNodeTypes()
}