package stm32

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
	if g.ID() != common.Target("stm32") {
		t.Errorf("expected ID stm32, got %s", g.ID())
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
		Target: common.Target("stm32"),
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
		Target: common.Target("stm32"),
		Nodes:  []common.Node{},
		Edges:  []common.Edge{},
	}
	rr, err := g.RuntimeRequirement(wf)
	if err != nil {
		t.Fatalf("RuntimeRequirement failed: %v", err)
	}
	if rr.Name != "stm32" {
		t.Errorf("expected stm32, got %s", rr.Name)
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
		Target:  common.Target("stm32"),
		Version: 1,
		Author:  "tester",
		Nodes: []common.Node{
			{
				ID:      "sensor",
				Type:    "data_loader",
				Name:    "Read Sensor",
				Outputs: []common.Port{{ID: "data", Name: "Data", Type: "sensor_data"}},
			},
			{
				ID:      "process",
				Type:    "data_preprocessor",
				Name:    "Process Data",
				Inputs:  []common.Port{{ID: "input", Name: "Input", Type: "sensor_data", Required: true}},
				Outputs: []common.Port{{ID: "result", Name: "Result", Type: "processed_data"}},
			},
		},
		Edges: []common.Edge{
			{
				ID:     "e1",
				Source: common.EdgeEndpoint{NodeID: "sensor", PortID: "data"},
				Target: common.EdgeEndpoint{NodeID: "process", PortID: "input"},
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
		"Makefile",
		"test-project.ioc",
		"Core/Src/main.c",
		"Core/Inc/main.h",
		"Core/Src/stm32_it.c",
		"Core/Src/aistudio_nodes.c",
		"Core/Inc/aistudio_nodes.h",
		".gitignore",
	}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(outputDir, f)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file does not exist: %s", f)
		}
	}

	mainPath := filepath.Join(outputDir, "Core", "Src", "main.c")
	mainData, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("failed to read main.c: %v", err)
	}
	if !strings.Contains(string(mainData), "Test Workflow") {
		t.Error("main.c should contain workflow name")
	}
}

func TestGenerateDryRun(t *testing.T) {
	g := NewGenerator()
	outputDir := t.TempDir()

	wf := &common.Workflow{
		ID:     "dry-run-test",
		Name:   "Dry Run",
		Target: common.Target("stm32"),
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

	makePath := filepath.Join(outputDir, "Makefile")
	if _, err := os.Stat(makePath); !os.IsNotExist(err) {
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

func TestNodeTypeToFuncName(t *testing.T) {
	tests := []struct {
		nodeType string
		expected string
	}{
		{"data_loader", "AISTUDIO_DataLoad"},
		{"model_trainer", "AISTUDIO_TrainModel"},
		{"model_inference", "AISTUDIO_RunInference"},
		{"custom", "AISTUDIO_Custom"},
	}

	for _, tt := range tests {
		result := nodeTypeToFuncName(tt.nodeType)
		if result != tt.expected {
			t.Errorf("nodeTypeToFuncName(%q) = %q, want %q", tt.nodeType, result, tt.expected)
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
	if est.MinMemoryMB != 512 {
		t.Errorf("expected MinMemoryMB 512, got %d", est.MinMemoryMB)
	}
}