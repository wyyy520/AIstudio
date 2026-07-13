package matlab

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aistudio/packages/generators/common"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	if g.ID() != common.Target("matlab") {
		t.Errorf("expected ID matlab, got %s", g.ID())
	}
	if g.Name() == "" {
		t.Error("expected non-empty name")
	}
	if g.Description() == "" {
		t.Error("expected non-empty description")
	}
	if g.Version() == "" {
		t.Error("expected non-empty version")
	}
}

func TestGeneratorValidate(t *testing.T) {
	g := NewGenerator()
	wf := &common.Workflow{
		ID:     "test",
		Name:   "Test",
		Target: common.Target("matlab"),
		Nodes:  []common.Node{},
		Edges:  []common.Edge{},
	}
	if err := g.Validate(wf); err != nil {
		t.Errorf("expected valid workflow, got: %v", err)
	}
}

func TestGeneratorRuntimeRequirement(t *testing.T) {
	g := NewGenerator()
	wf := &common.Workflow{
		ID:     "test",
		Name:   "Test",
		Target: common.Target("matlab"),
		Nodes:  []common.Node{},
		Edges:  []common.Edge{},
	}
	rr, err := g.RuntimeRequirement(wf)
	if err != nil {
		t.Fatalf("RuntimeRequirement failed: %v", err)
	}
	if rr.Name != "matlab" {
		t.Errorf("expected matlab, got %s", rr.Name)
	}
	if len(rr.Commands) == 0 {
		t.Error("expected at least one command")
	}
}

func TestGenerateBasic(t *testing.T) {
	g := NewGenerator()
	outputDir := t.TempDir()

	wf := &common.Workflow{
		ID:      "test-wf",
		Name:    "Test Workflow",
		Target:  common.Target("matlab"),
		Version: 1,
		Author:  "tester",
		Nodes: []common.Node{
			{
				ID:      "loader",
				Type:    "data_loader",
				Name:    "Load Data",
				Outputs: []common.Port{{ID: "dataset", Name: "Dataset", Type: "dataset"}},
				Config:  map[string]any{"source": "data/input"},
			},
			{
				ID:      "trainer",
				Type:    "model_trainer",
				Name:    "Train Model",
				Inputs:  []common.Port{{ID: "train_data", Name: "Training Data", Type: "dataset", Required: true}},
				Outputs: []common.Port{{ID: "model", Name: "Model", Type: "model"}},
				Config:  map[string]any{"epochs": 10},
			},
		},
		Edges: []common.Edge{
			{
				ID:     "e1",
				Source: common.EdgeEndpoint{NodeID: "loader", PortID: "dataset"},
				Target: common.EdgeEndpoint{NodeID: "trainer", PortID: "train_data"},
			},
		},
	}

	result, err := g.Generate(context.Background(), wf, common.CompileOptions{
		OutputDir:   outputDir,
		Force:       true,
		ProjectName: "test-project",
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result.ProjectRoot != outputDir {
		t.Errorf("expected project root %s, got %s", outputDir, result.ProjectRoot)
	}
	if len(result.Files) == 0 {
		t.Fatal("expected at least one generated file")
	}

	expectedFiles := []string{
		"startup.m",
		"run.m",
		".gitignore",
		"scripts/load_dataset.m",
		"scripts/train_model.m",
	}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(outputDir, f)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file does not exist: %s", f)
		}
	}

	runPath := filepath.Join(outputDir, "run.m")
	runData, err := os.ReadFile(runPath)
	if err != nil {
		t.Fatalf("failed to read run.m: %v", err)
	}
	runContent := string(runData)
	if !strings.Contains(runContent, "Test Workflow") {
		t.Error("run.m should contain workflow name")
	}
	if !strings.Contains(runContent, "load_dataset") {
		t.Error("run.m should call load_dataset")
	}
}

func TestGenerateDryRun(t *testing.T) {
	g := NewGenerator()
	outputDir := t.TempDir()

	wf := &common.Workflow{
		ID:     "dry-run-test",
		Name:   "Dry Run",
		Target: common.Target("matlab"),
		Nodes: []common.Node{
			{ID: "n1", Type: "data_loader", Name: "Load", Outputs: []common.Port{{ID: "out", Name: "Output", Type: "dataset"}}},
		},
	}

	result, err := g.Generate(context.Background(), wf, common.CompileOptions{
		OutputDir:   outputDir,
		DryRun:      true,
		ProjectName: "dry-run-project",
	})
	if err != nil {
		t.Fatalf("Generate (dry run) failed: %v", err)
	}

	startupPath := filepath.Join(outputDir, "startup.m")
	if _, err := os.Stat(startupPath); !os.IsNotExist(err) {
		t.Error("dry run should not write files")
	}

	if len(result.Files) == 0 {
		t.Error("dry run should still report generated files")
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Load Data", "load_data"},
		{"Train-Model", "train_model"},
		{"Test@#$%", "test"},
		{"Normal", "normal"},
		{"", "unnamed"},
	}

	for _, tt := range tests {
		result := sanitizeName(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToClassName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"load data", "LoadData"},
		{"train_model", "TrainModel"},
		{"test", "Test"},
	}

	for _, tt := range tests {
		result := toClassName(tt.input)
		if result != tt.expected {
			t.Errorf("toClassName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNodeTypeToScriptName(t *testing.T) {
	tests := []struct {
		nodeType string
		expected string
	}{
		{"data_loader", "load_dataset"},
		{"data_preprocessor", "preprocess_data"},
		{"model_trainer", "train_model"},
		{"model_inference", "run_inference"},
		{"custom", "custom"},
	}

	for _, tt := range tests {
		result := nodeTypeToScriptName(tt.nodeType)
		if result != tt.expected {
			t.Errorf("nodeTypeToScriptName(%q) = %q, want %q", tt.nodeType, result, tt.expected)
		}
	}
}

func TestEstimateResources(t *testing.T) {
	g := NewGenerator()
	wf := &common.Workflow{
		Nodes: []common.Node{
			{ID: "n1", Type: "data_loader", Name: "Load"},
			{ID: "n2", Type: "model_trainer", Name: "Train"},
		},
	}
	est, err := g.EstimateResources(wf)
	if err != nil {
		t.Fatalf("EstimateResources failed: %v", err)
	}
	if est.EstimatedFiles <= 0 {
		t.Errorf("expected positive file estimate, got %d", est.EstimatedFiles)
	}
}