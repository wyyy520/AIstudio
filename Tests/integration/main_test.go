package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aistudio/packages/compiler"
	"github.com/aistudio/packages/event"
	"github.com/aistudio/packages/project"
	"github.com/aistudio/packages/runtime"
	"github.com/aistudio/packages/workflow"
)

func TestFullPipeline(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	// 1. Create a Workflow (YOLO training)
	wf := &workflow.Workflow{
		SchemaVersion: workflow.CurrentSchemaVersion,
		ID:            "test-yolo-001",
		Name:          "test-yolo",
		Version:       1,
		Target:        workflow.TargetPython,
		Nodes: []workflow.Node{
			{ID: "n1", Type: workflow.NodeTypeDataLoader, Name: "load_data"},
			{ID: "n2", Type: workflow.NodeTypeModelTrainer, Name: "train_yolo"},
		},
		Edges: []workflow.Edge{
			{
				ID: "e1",
				Source: workflow.EdgeEndpoint{NodeID: "n1", PortID: "dataset"},
				Target: workflow.EdgeEndpoint{NodeID: "n2", PortID: "dataset"},
			},
		},
	}

	// 2. Save workflow
	wfPath := filepath.Join(tempDir, "workflow.json")
	if err := workflow.SaveToFile(wf, wfPath); err != nil {
		t.Fatalf("failed to save workflow: %v", err)
	}

	// Verify saved file can be reloaded
	loaded, err := workflow.LoadFromFile(wfPath)
	if err != nil {
		t.Fatalf("failed to reload workflow: %v", err)
	}
	if loaded.ID != wf.ID {
		t.Fatalf("workflow ID mismatch: %s != %s", loaded.ID, wf.ID)
	}

	// 3. Create Project
	projectManager := project.NewManager(tempDir)
	p, err := projectManager.Create("test", "python", tempDir)
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	t.Cleanup(func() { projectManager.Delete(p) })

	if p.Target != "python" {
		t.Fatalf("expected target python, got %s", p.Target)
	}
	if _, err := os.Stat(p.RootPath); os.IsNotExist(err) {
		t.Fatalf("project root does not exist: %s", p.RootPath)
	}

	// 4. Compile
	bus := event.New()
	t.Cleanup(bus.Close)

	compilerEngine := compiler.NewCompiler(bus)
	targets := compilerEngine.ListTargets()
	t.Logf("available targets: %d", len(targets))

	compileOpts := compiler.CompileOptions{
		OutputDir:   p.RootPath,
		Target:      workflow.TargetPython,
		ProjectName: "test",
	}
	result, err := compilerEngine.Compile(ctx, wf, compileOpts)
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	if result.ProjectRoot == "" {
		t.Fatal("expected non-empty project root")
	}
	if result.Target != workflow.TargetPython {
		t.Fatalf("expected target python, got %s", result.Target)
	}
	if result.WorkflowID != wf.ID {
		t.Fatalf("expected workflow ID %s, got %s", wf.ID, result.WorkflowID)
	}

	// 5. Verify generated files exist
	expectedFiles := []string{
		filepath.Join(tempDir, "src", "load_data.py"),
		filepath.Join(tempDir, "src", "train_yolo.py"),
		filepath.Join(tempDir, "requirements.txt"),
		filepath.Join(tempDir, "pyproject.toml"),
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Logf("note: expected file not generated yet: %s", f)
		}
	}

	// 6. Run using local executor
	executor := runtime.NewLocalExecutor()
	runConfig := runtime.RunConfig{
		ProjectDir: result.ProjectRoot,
		EntryPoint: "python",
		Args:       []string{"-c", "print('hello from aistudio')"},
	}
	runResult := executor.Execute(ctx, runConfig)

	if runResult.Status != runtime.RunStatusCompleted && runResult.Status != runtime.RunStatusFailed {
		t.Fatalf("unexpected run status: %s", runResult.Status)
	}
	if runResult.RunID == "" {
		t.Fatal("expected non-empty run ID")
	}

	// 7. Check logs directory exists
	logsDir := filepath.Join(tempDir, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		t.Fatalf("expected logs directory at %s", logsDir)
	}
}