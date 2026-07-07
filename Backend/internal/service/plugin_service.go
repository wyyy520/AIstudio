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

// List returns all registered plugins.
func (s *PluginService) List() []*plugin.Plugin {
	return s.manager.ListPlugins()
}

// Get returns a single plugin by name.
func (s *PluginService) Get(name string) (*plugin.Plugin, error) {
	return s.manager.GetPlugin(name)
}

// Install installs a new plugin by scanning the plugins directory.
func (s *PluginService) Install(name string) error {
	return s.manager.InstallPlugin(name)
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

// Uninstall removes a plugin.
func (s *PluginService) Uninstall(name string) error {
	return s.manager.RemovePlugin(name)
}

// Execute runs a plugin with the given input.
func (s *PluginService) Execute(ctx context.Context, name string, input map[string]interface{}) (interface{}, error) {
	return s.manager.ExecutePlugin(ctx, name, input)
}