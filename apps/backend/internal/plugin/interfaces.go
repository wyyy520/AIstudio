package plugin

import "context"

// ============================================================================
// Plugin System V2 — Pure Declaration
//
// Plugin manifests are pure declarations of capabilities.
// They contain NO executable code.
// The Generator reads plugin manifests to know what code to generate.
// ============================================================================

// Manifest is the interface for reading plugin declarations.
type Manifest interface {
	// ID returns the unique plugin identifier.
	ID() string
	// GetManifest returns the parsed manifest v2 data.
	GetManifest() *ManifestV2
}

// RegistryInterface is the interface for plugin registration and lookup.
type RegistryInterface interface {
	// Register adds a plugin manifest to the registry.
	Register(p *Plugin) error
	// Unregister removes a plugin.
	Unregister(name string) error
	// Get returns a plugin by name.
	Get(name string) (*Plugin, bool)
	// GetByID returns a plugin by ID.
	GetByID(id string) (*Plugin, bool)
	// List returns all registered plugins.
	List() []*Plugin
	// ListByType returns plugins filtered by type.
	ListByType(pluginType PluginType) []*Plugin
	// ListEnabled returns only enabled plugins.
	ListEnabled() []*Plugin
	// UpdateEnabled sets the enabled state of a plugin.
	UpdateEnabled(name string, enabled bool) error
	// Count returns the number of registered plugins.
	Count() int
}

// Discovery is the interface for dynamic plugin discovery from directory.
type Discovery interface {
	// Discover scans the Plugins/ directory and returns all found plugin manifests.
	Discover() ([]*Plugin, error)
}

// PluginExecutor is the interface for executing a plugin with input/output JSON protocol.
type PluginExecutor interface {
	Execute(ctx context.Context, plugin *Plugin, input map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error)
	Language() string
}