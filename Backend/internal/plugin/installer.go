package plugin

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Installer handles the plugin installation workflow.
type Installer struct {
	pluginsDir string
	loader     *FileLoader
	depMgr     *DependencyManager
	registry   *Registry
}

// NewInstaller creates a new Installer.
func NewInstaller(pluginsDir string, loader *FileLoader, depMgr *DependencyManager, registry *Registry) *Installer {
	return &Installer{
		pluginsDir: pluginsDir,
		loader:     loader,
		depMgr:     depMgr,
		registry:   registry,
	}
}

// InstallResult holds the outcome of an installation attempt.
type InstallResult struct {
	Success      bool                    `json:"success"`
	Message      string                  `json:"message"`
	Plugin       *Plugin                 `json:"plugin,omitempty"`
	Dependencies []DependencyCheckResult `json:"dependency_results,omitempty"`
}

// InstallPlugin performs the full installation workflow:
// 1. Read plugin.json (manifest)
// 2. Check environment / dependencies
// 3. Download plugin (mock)
// 4. Install dependencies
// 5. Register nodes
func (i *Installer) InstallPlugin(manifestPath string) (*InstallResult, error) {
	log.Printf("[installer] installing plugin from %s", manifestPath)

	// Step 1: Read manifest
	manifest, err := i.loader.LoadManifest(manifestPath)
	if err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("failed to read manifest: %v", err)}, err
	}

	pluginDir := filepath.Dir(manifestPath)

	// Step 2: Check dependencies
	depResults, allOK := i.depMgr.CheckDependencies(manifest.Dependencies)
	if !allOK {
		return &InstallResult{
			Success:      false,
			Message:      "dependency check failed",
			Dependencies: depResults,
		}, fmt.Errorf("dependency check failed for plugin %s", manifest.Name)
	}
	log.Printf("[installer] dependency check passed for %s", manifest.Name)

	// Step 3: Mock download (in production: git clone, download zip, etc.)
	if err := i.mockDownload(manifest.Name, pluginDir); err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("download failed: %v", err)}, err
	}

	// Step 4: Install dependencies (mock)
	for _, dep := range manifest.Dependencies {
		if err := i.depMgr.InstallDependency(dep); err != nil {
			return &InstallResult{
				Success: false,
				Message: fmt.Sprintf("failed to install dependency %s: %v", dep.Name, err),
			}, err
		}
	}

	// Step 5: Build plugin object
	plugin := &Plugin{
		ID:           manifest.ID,
		Name:         manifest.Name,
		Version:      manifest.Version,
		Author:       manifest.Author,
		Type:         PluginType(manifest.Type),
		Description:  manifest.Description,
		Entry:        manifest.Entry,
		Source:       PluginSource(manifest.Source),
		Path:         pluginDir,
		Dependencies: manifest.Dependencies,
		Status:       StatusInstalled,
		Enabled:      true,
		Nodes:        manifest.Nodes,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if plugin.ID == "" {
		plugin.ID = plugin.Name
	}

	// Step 6: Register in the plugin registry
	if err := i.registry.Register(plugin); err != nil {
		return &InstallResult{
			Success: false,
			Message: fmt.Sprintf("registration failed: %v", err),
		}, err
	}

	log.Printf("[installer] plugin %s@%s installed successfully", plugin.Name, plugin.Version)
	return &InstallResult{
		Success:      true,
		Message:      fmt.Sprintf("plugin %s installed successfully", plugin.Name),
		Plugin:       plugin,
		Dependencies: depResults,
	}, nil
}

// InstallFromURL simulates installing a plugin from a remote source.
// In production, this would download the plugin from a URL or git repo.
func (i *Installer) InstallFromURL(name, url string) (*InstallResult, error) {
	log.Printf("[installer] installing plugin %s from URL: %s", name, url)

	pluginDir := filepath.Join(i.pluginsDir, name)
	manifestPath := filepath.Join(pluginDir, "plugin.json")

	// Check if already installed
	if _, err := os.Stat(manifestPath); err == nil {
		return i.InstallPlugin(manifestPath)
	}

	// Mock: create a plugin directory with a default manifest
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("failed to create plugin directory: %v", err)}, err
	}

	// Create a default plugin.json for the mock installation
	defaultManifest := PluginManifest{
		ID:           name,
		Name:         name,
		Version:      "1.0.0",
		Author:       "AIStudio",
		Type:         "system",
		Description:  fmt.Sprintf("Plugin %s installed from URL", name),
		Entry:        "main.py",
		Source:       "market",
		Dependencies: []Dependency{},
		Nodes:        []NodeRegistration{},
	}

	manifestData, err := json.MarshalIndent(defaultManifest, "", "  ")
	if err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("failed to create manifest: %v", err)}, err
	}

	entryPath := filepath.Join(pluginDir, "main.py")
	if err := os.WriteFile(entryPath, []byte("# Plugin entry point\n"), 0644); err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("failed to create entry file: %v", err)}, err
	}

	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("failed to write manifest: %v", err)}, err
	}

	return i.InstallPlugin(manifestPath)
}

// RemovePlugin removes an installed plugin.
func (i *Installer) RemovePlugin(name string) error {
	log.Printf("[installer] removing plugin %s", name)

	plugin, ok := i.registry.Get(name)
	if !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	if err := i.registry.Unregister(name); err != nil {
		return fmt.Errorf("failed to unregister plugin %s: %w", name, err)
	}

	_ = plugin
	log.Printf("[installer] plugin %s removed", name)
	return nil
}

// UpdatePlugin updates an installed plugin to a newer version.
func (i *Installer) UpdatePlugin(name string) (*InstallResult, error) {
	log.Printf("[installer] updating plugin %s", name)

	plugin, ok := i.registry.Get(name)
	if !ok {
		return &InstallResult{Success: false, Message: fmt.Sprintf("plugin not found: %s", name)}, fmt.Errorf("plugin not found: %s", name)
	}

	plugin.Status = StatusUpdating
	plugin.UpdatedAt = time.Now()

	manifestPath := filepath.Join(plugin.Path, "plugin.json")
	result, err := i.InstallPlugin(manifestPath)
	if err != nil {
		plugin.Status = StatusError
		return result, err
	}

	plugin.Status = StatusInstalled
	plugin.UpdatedAt = time.Now()
	return result, nil
}

// mockDownload simulates downloading a plugin.
func (i *Installer) mockDownload(name, targetDir string) error {
	log.Printf("[installer] downloading plugin %s (mock)", name)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin directory does not exist: %s", targetDir)
	}
	return nil
}