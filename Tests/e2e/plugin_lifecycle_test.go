package e2e

import (
	"testing"

	"github.com/aistudio/backend/internal/plugin"
)

func TestPluginRegistration(t *testing.T) {
	r := plugin.NewRegistry()

	p := &plugin.Plugin{
		ID:      "test-algo",
		Name:    "test-algo",
		Version: "2.0.0",
		Author:  "AIStudio",
		Type:    plugin.PluginTypeVision,
		Status:  plugin.StatusInstalled,
		Enabled: false,
		Nodes: []plugin.PluginNode{
			{
				Type:        "model_trainer.test",
				Name:        "Test Node",
				Description: "Test algorithm node",
				Inputs: []plugin.PortInfo{
					{ID: "input", Name: "Input", Type: "tensor", Required: true},
				},
				Outputs: []plugin.PortInfo{
					{ID: "output", Name: "Output", Type: "json"},
				},
			},
		},
	}

	if err := r.Register(p); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	if r.Count() != 1 {
		t.Errorf("expected 1 plugin, got %d", r.Count())
	}

	found, ok := r.Get("test-algo")
	if !ok {
		t.Fatal("expected to find plugin 'test-algo'")
	}
	if found.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %s", found.Version)
	}
	if len(found.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(found.Nodes))
	}
}

func TestPluginEnableDisable(t *testing.T) {
	r := plugin.NewRegistry()
	r.Register(&plugin.Plugin{Name: "p1", ID: "p1", Enabled: false, Status: plugin.StatusInstalled})

	if err := r.UpdateEnabled("p1", true); err != nil {
		t.Fatalf("UpdateEnabled() failed: %v", err)
	}

	p, _ := r.Get("p1")
	if !p.Enabled {
		t.Error("expected plugin to be enabled")
	}
	if p.Status != plugin.StatusEnabled {
		t.Errorf("expected status enabled, got %s", p.Status)
	}

	if err := r.UpdateEnabled("p1", false); err != nil {
		t.Fatalf("UpdateEnabled(false) failed: %v", err)
	}

	p, _ = r.Get("p1")
	if p.Enabled {
		t.Error("expected plugin to be disabled")
	}
	if p.Status != plugin.StatusDisabled {
		t.Errorf("expected status disabled, got %s", p.Status)
	}
}

func TestPluginListEnabled(t *testing.T) {
	r := plugin.NewRegistry()
	r.Register(&plugin.Plugin{Name: "a", ID: "a", Enabled: true, Status: plugin.StatusEnabled})
	r.Register(&plugin.Plugin{Name: "b", ID: "b", Enabled: false, Status: plugin.StatusDisabled})
	r.Register(&plugin.Plugin{Name: "c", ID: "c", Enabled: true, Status: plugin.StatusEnabled})

	enabled := r.ListEnabled()
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled plugins, got %d", len(enabled))
	}
}

func TestPluginUnregister(t *testing.T) {
	r := plugin.NewRegistry()
	r.Register(&plugin.Plugin{Name: "to-remove", ID: "to-remove", Enabled: false})

	if err := r.Unregister("to-remove"); err != nil {
		t.Fatalf("Unregister() failed: %v", err)
	}
	if r.Count() != 0 {
		t.Errorf("expected 0 plugins after unregister, got %d", r.Count())
	}

	if err := r.Unregister("nonexistent"); err == nil {
		t.Error("expected error when unregistering nonexistent plugin")
	}
}

func TestPluginDuplicateRegistration(t *testing.T) {
	r := plugin.NewRegistry()
	r.Register(&plugin.Plugin{Name: "dup", ID: "dup"})

	err := r.Register(&plugin.Plugin{Name: "dup", ID: "dup"})
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestPluginSummary(t *testing.T) {
	p := &plugin.Plugin{
		ID:      "test",
		Name:    "Test",
		Version: "2.0.0",
		Author:  "AIStudio",
		Type:    plugin.PluginTypeVision,
		Status:  plugin.StatusEnabled,
		Enabled: true,
		Manifest: &plugin.ManifestV2{
			Kind: "algorithm",
		},
		Nodes: []plugin.PluginNode{
			{Type: "test.node", Name: "Test Node"},
		},
	}

	s := p.ToSummary()
	if s.ID != "test" {
		t.Errorf("expected id 'test', got '%s'", s.ID)
	}
	if s.NodeCount != 1 {
		t.Errorf("expected node count 1, got %d", s.NodeCount)
	}
	if s.Kind != "algorithm" {
		t.Errorf("expected kind 'algorithm', got '%s'", s.Kind)
	}
}
