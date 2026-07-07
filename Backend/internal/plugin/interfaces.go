package plugin

import "context"

// Executor defines the interface for executing a plugin.
type Executor interface {
	// Execute runs the plugin with the given name and input.
	Execute(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error)
}

// Loader defines the interface for loading plugins.
type Loader interface {
	// LoadPlugin loads a plugin from the given path.
	LoadPlugin(path string) (*Plugin, error)
	// LoadManifest reads and parses plugin.json from the given path.
	LoadManifest(path string) (*PluginManifest, error)
}