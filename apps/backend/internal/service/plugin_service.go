package service

import (
	"context"
	"fmt"

	"github.com/aistudio/backend/internal/plugin"
	"gorm.io/gorm"
)

// PluginService handles plugin business logic for V2 (pure declaration).
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

// GetNodeTypes returns all workflow node types from registered plugins.
func (s *PluginService) GetNodeTypes() []plugin.PluginNode {
	return s.manager.GetNodeTypes()
}

// ListEnabled returns only enabled plugins.
func (s *PluginService) ListEnabled() []*plugin.Plugin {
	return s.manager.ListEnabled()
}

// Refresh rescans the plugins directory for new/updated manifests.
func (s *PluginService) Refresh() error {
	return s.manager.DiscoverPlugins()
}

// Install installs a plugin from a manifest URL.
func (s *PluginService) Install(ctx context.Context, manifestURL string) (*plugin.InstallTask, error) {
	return s.manager.Installer().Install(ctx, manifestURL)
}

// Uninstall removes a plugin by name.
func (s *PluginService) Uninstall(ctx context.Context, name string) error {
	return s.manager.Installer().Uninstall(ctx, name)
}

// InstallStatus returns the install status for a plugin name.
func (s *PluginService) InstallStatus(name string) *plugin.InstallStatus {
	return s.manager.Installer().GetInstallStatusByName(name)
}

// Execute runs a plugin with the given input.
func (s *PluginService) Execute(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error) {
	return s.manager.Execute(ctx, name, input, nil)
}
