package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	// Register
	p := &Plugin{Name: "test-plugin", Version: "1.0", Status: StatusInstalled}
	if err := r.Register(p); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Duplicate should fail
	if err := r.Register(p); err == nil {
		t.Fatal("expected error for duplicate registration")
	}

	// Get
	found, ok := r.Get("test-plugin")
	if !ok {
		t.Fatal("expected to find plugin")
	}
	if found.Name != "test-plugin" {
		t.Errorf("expected test-plugin, got %s", found.Name)
	}

	// Not found
	_, ok = r.Get("nonexistent")
	if ok {
		t.Fatal("expected not found")
	}

	// List
	plugins := r.List()
	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}

	// Update status
	if err := r.UpdateStatus("test-plugin", StatusEnabled); err != nil {
		t.Fatalf("UpdateStatus() failed: %v", err)
	}
	found, _ = r.Get("test-plugin")
	if found.Status != StatusEnabled {
		t.Errorf("expected enabled, got %s", found.Status)
	}

	// Unregister
	if err := r.Unregister("test-plugin"); err != nil {
		t.Fatalf("Unregister() failed: %v", err)
	}
	if r.Count() != 0 {
		t.Errorf("expected 0 plugins, got %d", r.Count())
	}
}

func TestFileLoader(t *testing.T) {
	// Create a temp plugin directory
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "my-plugin")
	os.MkdirAll(pluginDir, 0755)

	// Create plugin.json
	manifest := `{
		"name": "my-plugin",
		"version": "1.0.0",
		"description": "My test plugin",
		"entry": "main.py"
	}`
	os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(manifest), 0644)
	os.WriteFile(filepath.Join(pluginDir, "main.py"), []byte("print('hello')"), 0644)

	loader := NewFileLoader(tmpDir)
	plugin, err := loader.LoadPlugin(pluginDir)
	if err != nil {
		t.Fatalf("LoadPlugin() failed: %v", err)
	}

	if plugin.Name != "my-plugin" {
		t.Errorf("expected my-plugin, got %s", plugin.Name)
	}
	if plugin.Version != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", plugin.Version)
	}

	// Scan dir
	plugins, err := loader.ScanDir()
	if err != nil {
		t.Fatalf("ScanDir() failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}
}

func TestSimpleExecutor(t *testing.T) {
	r := NewRegistry()
	r.Register(&Plugin{Name: "exec-plugin", Version: "1.0", Description: "exec test", Entry: "main.py", Status: StatusEnabled})

	e := NewSimpleExecutor(r)

	// Execute enabled plugin
	result, err := e.Execute(context.Background(), "exec-plugin", map[string]interface{}{"foo": "bar"})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Execute nonexistent plugin
	_, err = e.Execute(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent plugin")
	}
}

func TestManager(t *testing.T) {
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "test-p")
	os.MkdirAll(pluginDir, 0755)

	manifest := `{"name": "test-p", "version": "1.0", "description": "test", "entry": "run.py"}`
	os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(manifest), 0644)
	os.WriteFile(filepath.Join(pluginDir, "run.py"), []byte("print('ok')"), 0644)

	m := NewManager(tmpDir)
	if err := m.DiscoverPlugins(); err != nil {
		t.Fatalf("DiscoverPlugins() failed: %v", err)
	}

	if m.PluginCount() != 1 {
		t.Errorf("expected 1 plugin, got %d", m.PluginCount())
	}

	// Enable/disable
	if err := m.EnablePlugin("test-p"); err != nil {
		t.Fatalf("EnablePlugin() failed: %v", err)
	}
	p, _ := m.GetPlugin("test-p")
	if p.Status != StatusEnabled {
		t.Errorf("expected enabled, got %s", p.Status)
	}

	if err := m.DisablePlugin("test-p"); err != nil {
		t.Fatalf("DisablePlugin() failed: %v", err)
	}
	p, _ = m.GetPlugin("test-p")
	if p.Status != StatusDisabled {
		t.Errorf("expected disabled, got %s", p.Status)
	}
}