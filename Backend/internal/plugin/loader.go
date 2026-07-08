package plugin

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// FileLoader implements Loader using the filesystem.
type FileLoader struct {
	pluginsDir string
}

// NewFileLoader creates a new FileLoader that scans the given directory.
func NewFileLoader(pluginsDir string) *FileLoader {
	return &FileLoader{
		pluginsDir: pluginsDir,
	}
}

// LoadManifest reads and parses plugin.json from the given path.
func (l *FileLoader) LoadManifest(path string) (*PluginManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin manifest %s: %w", path, err)
	}

	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse plugin manifest %s: %w", path, err)
	}

	// Validate required fields
	if manifest.Name == "" {
		return nil, fmt.Errorf("plugin manifest %s: name is required", path)
	}
	if manifest.Version == "" {
		return nil, fmt.Errorf("plugin manifest %s: version is required", path)
	}
	if manifest.Entry == "" {
		return nil, fmt.Errorf("plugin manifest %s: entry is required", path)
	}

	// Set defaults
	if manifest.Type == "" {
		manifest.Type = "system"
	}
	if manifest.ID == "" {
		manifest.ID = manifest.Name
	}
	if manifest.Source == "" {
		manifest.Source = "local"
	}

	return &manifest, nil
}

// LoadPlugin loads a plugin from a directory containing plugin.json.
func (l *FileLoader) LoadPlugin(path string) (*Plugin, error) {
	manifestPath := filepath.Join(path, "plugin.json")

	info, err := os.Stat(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plugin.json not found in %s", path)
		}
		return nil, fmt.Errorf("failed to stat %s: %w", manifestPath, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("plugin.json is a directory in %s", path)
	}

	manifest, err := l.LoadManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	// Verify entry file exists
	entryPath := filepath.Join(path, manifest.Entry)
	if _, err := os.Stat(entryPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin entry file not found: %s", entryPath)
	}

	now := time.Now()
	plugin := &Plugin{
		ID:           manifest.ID,
		Name:         manifest.Name,
		Version:      manifest.Version,
		Author:       manifest.Author,
		Type:         PluginType(manifest.Type),
		Description:  manifest.Description,
		Entry:        manifest.Entry,
		Source:       PluginSource(manifest.Source),
		Path:         path,
		Dependencies: manifest.Dependencies,
		Nodes:        manifest.Nodes,
		Status:       StatusInstalled,
		Enabled:      true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if plugin.ID == "" {
		plugin.ID = plugin.Name
	}

	return plugin, nil
}

// ScanDir scans the plugins directory and discovers all plugins.
func (l *FileLoader) ScanDir() ([]*Plugin, error) {
	entries, err := os.ReadDir(l.pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read plugins directory %s: %w", l.pluginsDir, err)
	}

	var plugins []*Plugin
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(l.pluginsDir, entry.Name())
		plugin, err := l.LoadPlugin(pluginPath)
		if err != nil {
			log.Printf("[plugin-loader] warning: failed to load plugin from %s: %v", pluginPath, err)
			continue
		}

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

// GetPluginsDir returns the plugins directory path.
func (l *FileLoader) GetPluginsDir() string {
	return l.pluginsDir
}