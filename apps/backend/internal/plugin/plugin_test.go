package plugin

import (
	"context"
	"os"
	"testing"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	// Register
	p := &Plugin{Name: "test-plugin", Version: "1.0", ID: "test-p", Status: StatusInstalled}
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
	if err := r.UpdateEnabled("test-plugin", true); err != nil {
		t.Fatalf("UpdateEnabled() failed: %v", err)
	}
	found, _ = r.Get("test-plugin")
	if !found.Enabled {
		t.Errorf("expected enabled=true")
	}

	// Unregister
	if err := r.Unregister("test-plugin"); err != nil {
		t.Fatalf("Unregister() failed: %v", err)
	}
	if r.Count() != 0 {
		t.Errorf("expected 0 plugins, got %d", r.Count())
	}
}

func TestLoadManifestV2(t *testing.T) {
	// Test that ManifestV2 can be parsed correctly
	mv := &ManifestV2{
		ID:               "test-algo",
		Name:             "Test Algorithm",
		Version:          "2.0.0",
		MinSchemaVersion: "2.0.0",
		Kind:             ManifestKindAlgorithm,
		Description:      "A test algorithm plugin",
		Author:           "AIStudio",
		Nodes: []PluginNode{
			{
				Type:        "model_trainer.test",
				Name:        "Test Node",
				Description: "Test description",
				Inputs: []PortInfo{
					{ID: "input", Name: "Input", Type: "tensor", Required: true},
				},
				Outputs: []PortInfo{
					{ID: "output", Name: "Output", Type: "json"},
				},
			},
		},
		RuntimeBundle:    "test-bundle",
		SupportedTargets: []string{"python"},
	}

	if mv.ID != "test-algo" {
		t.Errorf("expected test-algo, got %s", mv.ID)
	}
	if len(mv.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(mv.Nodes))
	}
	if mv.Nodes[0].Type != "model_trainer.test" {
		t.Errorf("expected model_trainer.test, got %s", mv.Nodes[0].Type)
	}
}

func TestPluginSummary(t *testing.T) {
	p := &Plugin{
		ID:      "test",
		Name:    "Test",
		Version: "2.0.0",
		Author:  "AIStudio",
		Type:    PluginTypeVision,
		Status:  StatusEnabled,
		Enabled: true,
		Manifest: &ManifestV2{
			Kind: "algorithm",
		},
		Nodes: []PluginNode{
			{Type: "test.node", Name: "Test Node"},
		},
	}

	s := p.ToSummary()
	if s.ID != "test" {
		t.Errorf("expected test, got %s", s.ID)
	}
	if s.NodeCount != 1 {
		t.Errorf("expected 1, got %d", s.NodeCount)
	}
	if s.Kind != "algorithm" {
		t.Errorf("expected algorithm, got %s", s.Kind)
	}
}

func TestListEnabled(t *testing.T) {
	r := NewRegistry()

	r.Register(&Plugin{Name: "p1", ID: "p1", Enabled: true})
	r.Register(&Plugin{Name: "p2", ID: "p2", Enabled: false})
	r.Register(&Plugin{Name: "p3", ID: "p3", Enabled: true})

	enabled := r.ListEnabled()
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled, got %d", len(enabled))
	}
}

func TestRegistryListByType(t *testing.T) {
	r := NewRegistry()

	r.Register(&Plugin{Name: "p1", ID: "p1", Type: PluginTypeVision, Enabled: true})
	r.Register(&Plugin{Name: "p2", ID: "p2", Type: PluginTypeNLP, Enabled: true})
	r.Register(&Plugin{Name: "p3", ID: "p3", Type: PluginTypeVision, Enabled: false})

	vision := r.ListByType(PluginTypeVision)
	if len(vision) != 2 {
		t.Errorf("expected 2 vision plugins, got %d", len(vision))
	}

	nlp := r.ListByType(PluginTypeNLP)
	if len(nlp) != 1 {
		t.Errorf("expected 1 NLP plugin, got %d", len(nlp))
	}

	sim := r.ListByType(PluginTypeSimulation)
	if len(sim) != 0 {
		t.Errorf("expected 0 simulation plugins, got %d", len(sim))
	}
}

func TestRegistryGetByID(t *testing.T) {
	r := NewRegistry()
	r.Register(&Plugin{Name: "test", ID: "abc-123", Version: "1.0", Status: StatusInstalled})

	p, ok := r.GetByID("abc-123")
	if !ok {
		t.Fatal("expected to find plugin by ID")
	}
	if p.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", p.Name)
	}

	_, ok = r.GetByID("nonexistent")
	if ok {
		t.Fatal("expected not found for nonexistent ID")
	}
}

func TestRegistryListSummaries(t *testing.T) {
	r := NewRegistry()
	r.Register(&Plugin{
		Name:    "summary-test",
		ID:      "summary-1",
		Version: "2.0",
		Author:  "AIStudio",
		Type:    PluginTypeVision,
		Status:  StatusEnabled,
		Enabled: true,
		Manifest: &ManifestV2{
			Kind: "algorithm",
		},
		Nodes: []PluginNode{
			{Type: "test.node", Name: "Test Node"},
		},
	})

	summaries := r.ListSummaries()
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].Name != "summary-test" {
		t.Errorf("expected name 'summary-test', got '%s'", summaries[0].Name)
	}
	if summaries[0].NodeCount != 1 {
		t.Errorf("expected node count 1, got %d", summaries[0].NodeCount)
	}
	if summaries[0].Kind != "algorithm" {
		t.Errorf("expected kind 'algorithm', got '%s'", summaries[0].Kind)
	}
}

func TestRegistryUpdateEnabledToDisabled(t *testing.T) {
	r := NewRegistry()
	r.Register(&Plugin{Name: "toggle", ID: "toggle-1", Status: StatusEnabled, Enabled: true})

	if err := r.UpdateEnabled("toggle", false); err != nil {
		t.Fatalf("UpdateEnabled(false) failed: %v", err)
	}

	p, _ := r.Get("toggle")
	if p.Enabled {
		t.Error("expected plugin to be disabled")
	}
	if p.Status != StatusDisabled {
		t.Errorf("expected StatusDisabled, got %s", p.Status)
	}
}

func TestRegistryCount(t *testing.T) {
	r := NewRegistry()
	if c := r.Count(); c != 0 {
		t.Errorf("expected 0, got %d", c)
	}

	r.Register(&Plugin{Name: "a", ID: "a"})
	r.Register(&Plugin{Name: "b", ID: "b"})
	if c := r.Count(); c != 2 {
		t.Errorf("expected 2, got %d", c)
	}

	r.Unregister("a")
	if c := r.Count(); c != 1 {
		t.Errorf("expected 1, got %d", c)
	}
}

func TestManagerBasicOperations(t *testing.T) {
	pluginsDir := t.TempDir()
	mgr := NewManager(pluginsDir)

	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}

	if count := mgr.PluginCount(); count != 0 {
		t.Errorf("expected 0 plugins initially, got %d", count)
	}

	if reg := mgr.GetRegistry(); reg == nil {
		t.Error("GetRegistry() returned nil")
	}
}

func TestManagerRegistryIntegration(t *testing.T) {
	pluginsDir := t.TempDir()
	mgr := NewManager(pluginsDir)
	reg := mgr.GetRegistry()

	p := &Plugin{Name: "integrated", ID: "int-1", Version: "1.0", Status: StatusInstalled}
	if err := reg.Register(p); err != nil {
		t.Fatalf("Register() via manager failed: %v", err)
	}

	if count := mgr.PluginCount(); count != 1 {
		t.Errorf("expected 1 plugin via manager, got %d", count)
	}

	if err := mgr.EnablePlugin("integrated"); err != nil {
		t.Fatalf("EnablePlugin() failed: %v", err)
	}

	enabled := mgr.ListEnabled()
	if len(enabled) != 1 {
		t.Errorf("expected 1 enabled plugin, got %d", len(enabled))
	}

	all := mgr.ListPlugins()
	if len(all) != 1 {
		t.Errorf("expected 1 total plugin, got %d", len(all))
	}

	summaries := mgr.ListPluginSummaries()
	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
	}
}

func TestManagerEnableDisablePlugin(t *testing.T) {
	pluginsDir := t.TempDir()
	mgr := NewManager(pluginsDir)
	reg := mgr.GetRegistry()

	reg.Register(&Plugin{Name: "toggle-me", ID: "toggle-me", Version: "1.0", Status: StatusInstalled})

	if err := mgr.EnablePlugin("toggle-me"); err != nil {
		t.Fatalf("EnablePlugin() failed: %v", err)
	}

	p, _ := reg.Get("toggle-me")
	if !p.Enabled {
		t.Error("expected plugin to be enabled after EnablePlugin()")
	}

	if err := mgr.DisablePlugin("toggle-me"); err != nil {
		t.Fatalf("DisablePlugin() failed: %v", err)
	}

	p, _ = reg.Get("toggle-me")
	if p.Enabled {
		t.Error("expected plugin to be disabled after DisablePlugin()")
	}
}

func TestManagerEnableNonExistent(t *testing.T) {
	mgr := NewManager(t.TempDir())
	if err := mgr.EnablePlugin("ghost"); err == nil {
		t.Fatal("expected error enabling non-existent plugin")
	}
}

func TestManagerGetPluginByID(t *testing.T) {
	mgr := NewManager(t.TempDir())
	reg := mgr.GetRegistry()
	reg.Register(&Plugin{Name: "by-id-test", ID: "abc-456", Version: "1.0"})

	p, err := mgr.GetPluginByID("abc-456")
	if err != nil {
		t.Fatalf("GetPluginByID() failed: %v", err)
	}
	if p.Name != "by-id-test" {
		t.Errorf("expected 'by-id-test', got '%s'", p.Name)
	}

	_, err = mgr.GetPluginByID("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent ID")
	}
}

func TestManagerGetPlugin(t *testing.T) {
	mgr := NewManager(t.TempDir())
	reg := mgr.GetRegistry()
	reg.Register(&Plugin{Name: "fetch-me", ID: "fetch-1", Version: "1.0"})

	p, err := mgr.GetPlugin("fetch-me")
	if err != nil {
		t.Fatalf("GetPlugin() failed: %v", err)
	}
	if p.Name != "fetch-me" {
		t.Errorf("expected 'fetch-me', got '%s'", p.Name)
	}

	_, err = mgr.GetPlugin("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent plugin")
	}
}

func TestManagerListByType(t *testing.T) {
	mgr := NewManager(t.TempDir())
	reg := mgr.GetRegistry()
	reg.Register(&Plugin{Name: "v1", ID: "v1", Type: PluginTypeVision})
	reg.Register(&Plugin{Name: "v2", ID: "v2", Type: PluginTypeVision})
	reg.Register(&Plugin{Name: "n1", ID: "n1", Type: PluginTypeNLP})

	filtered := mgr.ListPluginsByType(PluginTypeVision)
	if len(filtered) != 2 {
		t.Errorf("expected 2 vision plugins, got %d", len(filtered))
	}
}

func TestManagerGetNodeTypes(t *testing.T) {
	mgr := NewManager(t.TempDir())
	reg := mgr.GetRegistry()

	reg.Register(&Plugin{
		Name:    "node-provider",
		ID:      "np-1",
		Status:  StatusInstalled,
		Enabled: true,
		Nodes: []PluginNode{
			{Type: "vision.detect", Name: "Detect"},
			{Type: "vision.classify", Name: "Classify"},
		},
	})

	reg.Register(&Plugin{
		Name:    "disabled-provider",
		ID:      "dp-1",
		Status:  StatusInstalled,
		Enabled: false,
		Nodes: []PluginNode{
			{Type: "vision.segment", Name: "Segment"},
		},
	})

	nodes := mgr.GetNodeTypes()
	if len(nodes) != 2 {
		t.Errorf("expected 2 node types from enabled plugins, got %d", len(nodes))
	}
}

func TestManagerRegistry(t *testing.T) {
	mgr := NewManager(t.TempDir())
	reg := mgr.GetRegistry()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestLoadManifestFromFile(t *testing.T) {
	pluginsDir := t.TempDir()

	manifestContent := `{
		"id": "fs-plugin",
		"name": "FS Plugin",
		"version": "1.0.0",
		"kind": "algorithm",
		"nodes": [
			{
				"type": "fs.read",
				"name": "Read File",
				"inputs": [{"id": "path", "name": "Path", "type": "string", "required": true}],
				"outputs": [{"id": "data", "name": "Data", "type": "string"}]
			}
		]
	}`

	pluginDir := pluginsDir + "/fs-plugin"
	if err := os.Mkdir(pluginDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := os.WriteFile(pluginDir+"/plugin.json", []byte(manifestContent), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	mgr := NewManager(pluginsDir)
	if err := mgr.DiscoverPlugins(); err != nil {
		t.Fatalf("DiscoverPlugins() failed: %v", err)
	}

	p, err := mgr.GetPlugin("FS Plugin")
	if err != nil {
		t.Fatalf("GetPlugin() failed: %v", err)
	}
	if p.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got '%s'", p.Version)
	}
	if len(p.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(p.Nodes))
	}
}

func TestInstallerNew(t *testing.T) {
	reg := NewRegistry()
	inst := NewInstaller(t.TempDir(), reg)
	if inst == nil {
		t.Fatal("NewInstaller() returned nil")
	}
}

func TestInstallerUninstallNonExistent(t *testing.T) {
	reg := NewRegistry()
	inst := NewInstaller(t.TempDir(), reg)

	err := inst.Uninstall(context.Background(), "nonexistent-plugin")
	if err == nil {
		t.Fatal("expected error when uninstalling non-existent plugin")
	}
}

func TestPluginDoubleRegister(t *testing.T) {
	r := NewRegistry()
	p := &Plugin{Name: "double", ID: "double", Version: "1.0", Status: StatusEnabled}

	if err := r.Register(p); err != nil {
		t.Fatalf("first Register() failed: %v", err)
	}

	if err := r.Register(p); err == nil {
		t.Fatal("expected error on duplicate register")
	}
}

func TestPluginGetNonExistent(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("does-not-exist")
	if ok {
		t.Fatal("expected not found for non-existent plugin")
	}
}

func TestPluginUpdateNonExistent(t *testing.T) {
	r := NewRegistry()
	err := r.UpdateEnabled("ghost", true)
	if err == nil {
		t.Fatal("expected error when updating non-existent plugin")
	}
}

func TestPluginUnregisterNonExistent(t *testing.T) {
	r := NewRegistry()
	err := r.Unregister("ghost")
	if err == nil {
		t.Fatal("expected error when unregistering non-existent plugin")
	}
}

func TestManagerSetPluginEventCallback(t *testing.T) {
	mgr := NewManager(t.TempDir())
	reg := mgr.GetRegistry()
	reg.Register(&Plugin{Name: "event-test", ID: "evt-1", Version: "1.0", Status: StatusInstalled})

	events := make([]string, 0)
	mgr.SetPluginEventCallback(func(event string, data map[string]interface{}) {
		events = append(events, event)
	})

	if err := mgr.EnablePlugin("event-test"); err != nil {
		t.Fatalf("EnablePlugin() failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0] != "enabled" {
		t.Errorf("expected 'enabled' event, got '%s'", events[0])
	}

	if err := mgr.DisablePlugin("event-test"); err != nil {
		t.Fatalf("DisablePlugin() failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[1] != "disabled" {
		t.Errorf("expected 'disabled' event, got '%s'", events[1])
	}
}

func TestManagerExecutorRegistration(t *testing.T) {
	mgr := NewManager(t.TempDir())

	mgr.RegisterExecutor("python", &mockExecutor{lang: "python"})
	mgr.RegisterExecutor("node", &mockExecutor{lang: "node"})
}

type mockExecutor struct {
	lang string
}

func (m *mockExecutor) Execute(ctx context.Context, plugin *Plugin, input map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"executed": true}, nil
}

func (m *mockExecutor) Language() string {
	return m.lang
}

func TestManagerExecuteWithExecutor(t *testing.T) {
	mgr := NewManager(t.TempDir())
	reg := mgr.GetRegistry()
	reg.Register(&Plugin{
		Name:    "exec-test",
		ID:      "exec-1",
		Version: "1.0",
		Status:  StatusEnabled,
		Enabled: true,
		Manifest: &ManifestV2{
			Language: "python",
		},
	})

	mgr.RegisterExecutor("python", &mockExecutor{lang: "python"})

	result, err := mgr.Execute(context.Background(), "exec-test", map[string]interface{}{"input": "data"}, nil)
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}
	if result["executed"] != true {
		t.Errorf("expected executed=true, got %v", result["executed"])
	}
}

func TestManagerExecutePluginNotFound(t *testing.T) {
	mgr := NewManager(t.TempDir())
	_, err := mgr.Execute(context.Background(), "nonexistent", nil, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent plugin")
	}
}

func TestManagerExecutePluginDisabled(t *testing.T) {
	mgr := NewManager(t.TempDir())
	reg := mgr.GetRegistry()
	reg.Register(&Plugin{
		Name:    "disabled-p",
		ID:      "dp-1",
		Version: "1.0",
		Status:  StatusDisabled,
		Enabled: false,
	})

	_, err := mgr.Execute(context.Background(), "disabled-p", nil, nil)
	if err == nil {
		t.Fatal("expected error for disabled plugin")
	}
}

func TestManagerExecuteNoExecutor(t *testing.T) {
	mgr := NewManager(t.TempDir())
	reg := mgr.GetRegistry()
	reg.Register(&Plugin{
		Name:    "no-exec",
		ID:      "ne-1",
		Version: "1.0",
		Status:  StatusEnabled,
		Enabled: true,
	})

	_, err := mgr.Execute(context.Background(), "no-exec", nil, nil)
	if err == nil {
		t.Fatal("expected error for missing executor")
	}
}

func TestInstallerGetInstallStatus(t *testing.T) {
	reg := NewRegistry()
	inst := NewInstaller(t.TempDir(), reg)

	status := inst.GetInstallStatus("nonexistent-task")
	if status != nil {
		t.Fatal("expected nil for nonexistent task")
	}

	status = inst.GetInstallStatusByName("nonexistent-plugin")
	if status != nil {
		t.Fatal("expected nil for nonexistent plugin status")
	}
}

func TestInstallerInstallSyncCancelledContext(t *testing.T) {
	reg := NewRegistry()
	inst := NewInstaller(t.TempDir(), reg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := inst.InstallSync(ctx, "http://example.com/plugin.json")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestManifestKindValidation(t *testing.T) {
	valid := ValidManifestKinds()
	expected := []string{"algorithm", "runtime", "system", "adapter", "tool"}

	if len(valid) != len(expected) {
		t.Fatalf("expected %d kinds, got %d", len(expected), len(valid))
	}
	for i, v := range valid {
		if v != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, v)
		}
	}
}
