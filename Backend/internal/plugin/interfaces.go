package plugin

import "context"

// Executor defines the interface for executing a plugin.
type Executor interface {
	// Execute runs the plugin with the given name and input.
	Execute(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error)
	// ExecuteNode runs a specific node from a plugin.
	ExecuteNode(ctx context.Context, pluginName, nodeType string, input map[string]interface{}) (map[string]interface{}, error)
}

// Loader defines the interface for loading plugins.
type Loader interface {
	// LoadPlugin loads a plugin from the given path.
	LoadPlugin(path string) (*Plugin, error)
	// LoadManifest reads and parses plugin.json from the given path.
	LoadManifest(path string) (*PluginManifest, error)
}

// PluginInstaller defines the interface for installing plugins.
type PluginInstaller interface {
	// InstallPlugin installs a plugin from a manifest path.
	InstallPlugin(manifestPath string) (*InstallResult, error)
	// InstallFromURL installs a plugin from a remote source.
	InstallFromURL(name, url string) (*InstallResult, error)
	// RemovePlugin removes an installed plugin.
	RemovePlugin(name string) error
	// UpdatePlugin updates a plugin to its latest version.
	UpdatePlugin(name string) (*InstallResult, error)
}

// PluginDependencyManager defines the interface for dependency management.
type PluginDependencyManager interface {
	// CheckDependencies validates all dependencies for a plugin.
	CheckDependencies(deps []Dependency) ([]DependencyCheckResult, bool)
	// InstallDependency installs a single dependency.
	InstallDependency(dep Dependency) error
}