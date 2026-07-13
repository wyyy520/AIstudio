package unity

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
	if g.ID() != common.Target("unity") {
		t.Errorf("expected ID unity, got %s", g.ID())
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
		Target: common.Target("unity"),
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
		Target: common.Target("unity"),
		Nodes:  []common.Node{},
		Edges:  []common.Edge{},
	}
	rr, err := g.RuntimeRequirement(wf)
	if err != nil {
		t.Fatalf("RuntimeRequirement failed: %v", err)
	}
	if rr.Name != "unity" {
		t.Errorf("expected unity, got %s", rr.Name)
	}
}

func TestGenerateBasic(t *testing.T) {
	g := NewGenerator()
	outputDir := t.TempDir()

	wf := &common.Workflow{
		ID:      "test-wf",
		Name:    "Test Workflow",
		Target:  common.Target("unity"),
		Version: 1,
		Author:  "tester",
		Nodes: []common.Node{
			{
				ID:      "loader",
				Type:    "data_loader",
				Name:    "Load Data",
				Outputs: []common.Port{{ID: "dataset", Name: "Dataset", Type: "dataset"}},
			},
			{
				ID:      "inference",
				Type:    "model_inference",
				Name:    "Run Inference",
				Inputs:  []common.Port{{ID: "data", Name: "Data", Type: "dataset", Required: true}},
				Outputs: []common.Port{{ID: "result", Name: "Result", Type: "prediction"}},
			},
		},
		Edges: []common.Edge{
			{
				ID:     "e1",
				Source: common.EdgeEndpoint{NodeID: "loader", PortID: "dataset"},
				Target: common.EdgeEndpoint{NodeID: "inference", PortID: "data"},
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
		"package.json",
		"Runtime/AistudioWorkflow.asmdef",
		"Editor/AistudioWorkflow.Editor.asmdef",
		"Runtime/WorkflowRunner.cs",
		"Editor/WorkflowEditor.cs",
		"Runtime/LoadDataNode.cs",
		"Runtime/RunInferenceNode.cs",
		".gitignore",
	}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(outputDir, f)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file does not exist: %s", f)
		}
	}

	pkgPath := filepath.Join(outputDir, "package.json")
	pkgData, err := os.ReadFile(pkgPath)
	if err != nil {
		t.Fatalf("failed to read package.json: %v", err)
	}
	if !strings.Contains(string(pkgData), "test-project") {
		t.Error("package.json should contain project name")
	}
}

func TestGenerateDryRun(t *testing.T) {
	g := NewGenerator()
	outputDir := t.TempDir()

	wf := &common.Workflow{
		ID:     "dry-run-test",
		Name:   "Dry Run",
		Target: common.Target("unity"),
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

	asmdefPath := filepath.Join(outputDir, "Runtime", "AistudioWorkflow.asmdef")
	if _, err := os.Stat(asmdefPath); !os.IsNotExist(err) {
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

func TestEstimateResources(t *testing.T) {
	g := NewGenerator()
	wf := &common.Workflow{
		Nodes: []common.Node{
			{ID: "n1", Type: "data_loader", Name: "Load"},
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