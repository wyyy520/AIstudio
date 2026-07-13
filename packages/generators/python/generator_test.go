package python

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
	if g.ID() != common.Target("python") {
		t.Errorf("expected ID python, got %s", g.ID())
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
		Target: common.Target("python"),
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
		Target: common.Target("python"),
		Nodes:  []common.Node{},
		Edges:  []common.Edge{},
	}

	rr, err := g.RuntimeRequirement(wf)
	if err != nil {
		t.Fatalf("RuntimeRequirement failed: %v", err)
	}
	if rr.Name != "python" {
		t.Errorf("expected python, got %s", rr.Name)
	}
	if len(rr.Packages) == 0 {
		t.Error("expected at least one package")
	}
}

func TestGeneratorEstimateResources(t *testing.T) {
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

func TestGenerateBasic(t *testing.T) {
	g := NewGenerator()
	outputDir := t.TempDir()

	wf := &common.Workflow{
		ID:      "test-wf",
		Name:    "Test Workflow",
		Target:  common.Target("python"),
		Version: 1,
		Author:  "tester",
		Nodes: []common.Node{
			{
				ID:      "loader",
				Type:    "data_loader",
				Name:    "Load Data",
				Outputs: []common.Port{{ID: "dataset", Name: "Dataset", Type: "dataset"}},
				Config:  map[string]any{"source": "data/coco128"},
			},
			{
				ID:      "trainer",
				Type:    "model_trainer",
				Name:    "Train Model",
				Inputs:  []common.Port{{ID: "train_data", Name: "Training Data", Type: "dataset", Required: true}},
				Outputs: []common.Port{{ID: "model", Name: "Model", Type: "model"}},
				Config:  map[string]any{"epochs": 10, "device": "cpu"},
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

	// Verify essential files exist on disk
	expectedFiles := []string{
		"main.py",
		"pyproject.toml",
		"requirements.txt",
		".gitignore",
		"config/config.yaml",
		"src/__init__.py",
		"src/utils.py",
		"src/load_data.py",
		"src/train_model.py",
		"tests/__init__.py",
		"tests/test_main.py",
	}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(outputDir, f)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file does not exist: %s", f)
		}
	}

	// Verify main.py exists and has expected content
	mainPath := filepath.Join(outputDir, "main.py")
	mainData, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("failed to read main.py: %v", err)
	}
	mainContent := string(mainData)
	if !strings.Contains(mainContent, "Test Workflow") {
		t.Error("main.py should contain workflow name")
	}
	if !strings.Contains(mainContent, "LoadData(") {
		t.Error("main.py should import LoadData class")
	}
	if !strings.Contains(mainContent, "TrainModel(") {
		t.Error("main.py should import TrainModel class")
	}
}

func TestGenerateWithYOLO(t *testing.T) {
	g := NewGenerator()
	outputDir := t.TempDir()

	wf := &common.Workflow{
		ID:      "yolo-wf",
		Name:    "YOLO Training",
		Target:  common.Target("python"),
		Version: 1,
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
				Name:    "Train YOLO",
				Inputs:  []common.Port{{ID: "data", Name: "Data", Type: "dataset", Required: true}},
				Outputs: []common.Port{{ID: "model", Name: "Model", Type: "model"}},
				Config:  map[string]any{"framework": "yolo", "model": "yolo11n.pt"},
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
		ProjectName: "yolo-project",
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(result.Files) == 0 {
		t.Fatal("expected at least one generated file")
	}

	// Verify YOLO-specific files
	trainerPath := filepath.Join(outputDir, "src", "train_yolo.py")
	if _, err := os.Stat(trainerPath); os.IsNotExist(err) {
		t.Error("expected train_yolo.py to exist (YOLO template)")
	}

	// Verify YOLO trainer content has YOLO-specific imports
	trainerData, err := os.ReadFile(trainerPath)
	if err != nil {
		t.Fatalf("failed to read train_yolo.py: %v", err)
	}
	if !strings.Contains(string(trainerData), "from ultralytics import YOLO") {
		t.Error("YOLO trainer should import ultralytics")
	}

	// Check YOLO packages in requirements.txt
	reqPath := filepath.Join(outputDir, "requirements.txt")
	reqData, err := os.ReadFile(reqPath)
	if err != nil {
		t.Fatalf("failed to read requirements.txt: %v", err)
	}
	if !strings.Contains(string(reqData), "ultralytics") {
		t.Error("YOLO workflow should include ultralytics in requirements")
	}
}

func TestGenerateDryRun(t *testing.T) {
	g := NewGenerator()
	outputDir := t.TempDir()

	wf := &common.Workflow{
		ID:     "dry-run-test",
		Name:   "Dry Run",
		Target: common.Target("python"),
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

	// In dry run mode, no files should be written
	mainPath := filepath.Join(outputDir, "main.py")
	if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
		t.Error("dry run should not write files")
	}

	// But the files should still be in the result
	if len(result.Files) == 0 {
		t.Error("dry run should still report generated files")
	}
}

func TestRequiresGPU(t *testing.T) {
	g := NewGenerator()

	wfCPU := &common.Workflow{
		Nodes: []common.Node{
			{ID: "n1", Type: "model_trainer", Name: "Train", Config: map[string]any{"device": "cpu"}},
		},
	}
	if g.requiresGPU(wfCPU) {
		t.Error("cpu workflow should not require GPU")
	}

	wfGPU := &common.Workflow{
		Nodes: []common.Node{
			{ID: "n1", Type: "model_trainer", Name: "Train", Config: map[string]any{"device": "cuda"}},
		},
	}
	if !g.requiresGPU(wfGPU) {
		t.Error("cuda workflow should require GPU")
	}
}

func TestIsYOLOWorkflow(t *testing.T) {
	g := NewGenerator()

	wfNoYOLO := &common.Workflow{
		Nodes: []common.Node{
			{ID: "n1", Type: "model_trainer", Name: "Train", Config: map[string]any{"device": "cpu"}},
		},
	}
	if g.isYOLOWorkflow(wfNoYOLO) {
		t.Error("non-YOLO workflow should not be detected as YOLO")
	}

	wfYOLO := &common.Workflow{
		Nodes: []common.Node{
			{ID: "n1", Type: "model_trainer", Name: "Train", Config: map[string]any{"framework": "yolo"}},
		},
	}
	if !g.isYOLOWorkflow(wfYOLO) {
		t.Error("YOLO workflow should be detected")
	}

	wfYOLOModel := &common.Workflow{
		Nodes: []common.Node{
			{ID: "n1", Type: "model_trainer", Name: "Train", Config: map[string]any{"model": "yolo11n.pt"}},
		},
	}
	if !g.isYOLOWorkflow(wfYOLOModel) {
		t.Error("YOLO model name should be detected")
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
		{"yolo training pipeline", "YoloTrainingPipeline"},
		{"test", "Test"},
	}

	for _, tt := range tests {
		result := toClassName(tt.input)
		if result != tt.expected {
			t.Errorf("toClassName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestCollectPackages(t *testing.T) {
	tests := []struct {
		name     string
		wf       *common.Workflow
		contains string
	}{
		{
			name: "basic packages",
			wf: &common.Workflow{
				Nodes: []common.Node{
					{ID: "n1", Type: "data_loader", Name: "Load"},
				},
			},
			contains: "numpy",
		},
		{
			name: "YOLO packages",
			wf: &common.Workflow{
				Nodes: []common.Node{
					{ID: "n1", Type: "model_trainer", Name: "Train", Config: map[string]any{"framework": "yolo"}},
				},
			},
			contains: "ultralytics",
		},
		{
			name: "visualization packages",
			wf: &common.Workflow{
				Nodes: []common.Node{
					{ID: "n1", Type: "visualization", Name: "Vis"},
				},
			},
			contains: "matplotlib",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := collectPackages(tt.wf)
			found := false
			for _, p := range pkgs {
				if strings.Contains(p, tt.contains) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected package containing %q, got %v", tt.contains, pkgs)
			}
		})
	}
}

func TestSortNodesByDAG(t *testing.T) {
	g := NewGenerator()
	wf := &common.Workflow{
		Nodes: []common.Node{
			{ID: "c", Type: "model_trainer", Name: "C"},
			{ID: "a", Type: "data_loader", Name: "A"},
			{ID: "b", Type: "data_preprocessor", Name: "B"},
		},
		Edges: []common.Edge{
			{ID: "e1", Source: common.EdgeEndpoint{NodeID: "a", PortID: "out"}, Target: common.EdgeEndpoint{NodeID: "b", PortID: "in"}},
			{ID: "e2", Source: common.EdgeEndpoint{NodeID: "b", PortID: "out"}, Target: common.EdgeEndpoint{NodeID: "c", PortID: "in"}},
		},
	}

	sorted := g.sortNodesByDAG(wf)
	if len(sorted) != 3 {
		t.Fatalf("expected 3 sorted nodes, got %d", len(sorted))
	}

	if sorted[0].ID != "a" || sorted[1].ID != "b" || sorted[2].ID != "c" {
		t.Errorf("unexpected sort order: got %v", sorted)
	}
}

func TestNodeTypeToTemplate(t *testing.T) {
	g := NewGenerator()
	tests := []struct {
		nodeType string
		isYOLO   bool
		expected string
	}{
		{"data_loader", false, "data_loader.py.tmpl"},
		{"data_preprocessor", false, "data_preprocessor.py.tmpl"},
		{"data_split", false, "data_split.py.tmpl"},
		{"model_trainer", false, "model_trainer.py.tmpl"},
		{"model_inference", false, "model_inference.py.tmpl"},
		{"custom", false, "src_node.py.tmpl"},
		{"model_trainer", true, "yolo_train.py.tmpl"},
		{"model_inference", true, "yolo_predict.py.tmpl"},
	}

	for _, tt := range tests {
		result := g.nodeTypeToTemplate(tt.nodeType, tt.isYOLO)
		if result != tt.expected {
			t.Errorf("nodeTypeToTemplate(%q, %v) = %q, want %q", tt.nodeType, tt.isYOLO, result, tt.expected)
		}
	}
}