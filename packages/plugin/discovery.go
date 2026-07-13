package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DiscoverPlugins scans a root directory for plugin.json files
// and returns all found plugin manifests.
func DiscoverPlugins(rootDir string) ([]*Plugin, error) {
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

		p, loadErr := LoadPluginFromManifest(path)
		if loadErr != nil {
			return nil
		}
		plugins = append(plugins, p)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk plugins dir failed: %w", err)
	}

	return plugins, nil
}

// LoadPluginFromManifest reads a plugin.json file and returns a Plugin.
func LoadPluginFromManifest(manifestPath string) (*Plugin, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var mv ManifestV2
	if err := json.Unmarshal(data, &mv); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

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

	var pluginType PluginType
	switch mv.Kind {
	case ManifestKindAlgorithm:
		pluginType = PluginTypeVision
	case ManifestKindRuntime:
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
