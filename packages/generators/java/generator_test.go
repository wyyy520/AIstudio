package java_

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
	if g.ID() != common.Target("java") {
		t.Errorf("expected ID java, got %s", g.ID())
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
		Target: common.Target("java"),
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
		Target: common.Target("java"),
		Nodes:  []common.Node{},
		Edges:  []common.Edge{},
	}
	rr, err := g.RuntimeRequirement(wf)
	if err != nil {
		t.Fatalf("RuntimeRequirement failed: %v", err)
	}
	if rr.Name != "java" {
		t.Errorf("expected java, got %s", rr.Name)
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
		Target:  common.Target("java"),
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
				ID:      "trainer",
				Type:    "model_trainer",
				Name:    "Train Model",
				Inputs:  []common.Port{{ID: "data", Name: "Data", Type: "dataset", Required: true}},
				Outputs: []common.Port{{ID: "model", Name: "Model", Type: "model"}},
			},
		},
		Edges: []common.Edge{
			{
				ID:     "e1",
				Source: common.EdgeEndpoint{NodeID: "loader", PortID: "dataset"},
				Target: common.EdgeEndpoint{NodeID: "trainer", PortID: "data"},
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
		"pom.xml",
		".gitignore",
		"src/main/java/com/aistudio/test_project/App.java",
		"src/main/java/com/aistudio/test_project/LoadDataNode.java",
		"src/main/java/com/aistudio/test_project/TrainModelNode.java",
		"src/test/java/com/aistudio/test_project/AppTest.java",
	}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(outputDir, f)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file does not exist: %s", f)
		}
	}

	appPath := filepath.Join(outputDir, "src/main/java/com/aistudio/test_project/App.java")
	appData, err := os.ReadFile(appPath)
	if err != nil {
		t.Fatalf("failed to read App.java: %v", err)
	}
	if !strings.Contains(string(appData), "Test Workflow") {
		t.Error("App.java should contain workflow name")
	}
}

func TestGenerateDryRun(t *testing.T) {
	g := NewGenerator()
	outputDir := t.TempDir()

	wf := &common.Workflow{
		ID:     "dry-run-test",
		Name:   "Dry Run",
		Target: common.Target("java"),
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

	pomPath := filepath.Join(outputDir, "pom.xml")
	if _, err := os.Stat(pomPath); !os.IsNotExist(err) {
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