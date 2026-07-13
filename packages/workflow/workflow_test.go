package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSerializeDeserialize(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "test-001",
		Name:          "Test Workflow",
		Description:   "A test workflow",
		Version:       1,
		Author:        "tester",
		Tags:          []string{"test", "example"},
		Target:        TargetPython,
		Nodes: []Node{
			{
				ID:   "node-1",
				Type: NodeTypeDataLoader,
				Name: "Load Data",
				Position: Point{X: 100, Y: 200},
				Inputs: []Port{
					{ID: "in-1", Name: "Input", Type: DataTypeFile, Required: true},
				},
				Outputs: []Port{
					{ID: "out-1", Name: "Output", Type: DataTypeDataset},
				},
			},
			{
				ID:   "node-2",
				Type: NodeTypeModelTrainer,
				Name: "Train Model",
				Position: Point{X: 300, Y: 200},
				Inputs: []Port{
					{ID: "in-1", Name: "Data", Type: DataTypeDataset, Required: true},
				},
				Outputs: []Port{
					{ID: "out-1", Name: "Model", Type: DataTypeModel},
				},
			},
		},
		Edges: []Edge{
			{
				ID: "edge-1",
				Source: EdgeEndpoint{NodeID: "node-1", PortID: "out-1"},
				Target: EdgeEndpoint{NodeID: "node-2", PortID: "in-1"},
			},
		},
	}

	data, err := ToJSON(wf)
	if err != nil {
		t.Fatalf("failed to marshal workflow: %v", err)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("failed to parse workflow: %v", err)
	}

	if parsed.ID != wf.ID {
		t.Errorf("ID mismatch: got %s, want %s", parsed.ID, wf.ID)
	}
	if parsed.Name != wf.Name {
		t.Errorf("Name mismatch: got %s, want %s", parsed.Name, wf.Name)
	}
	if parsed.Target != wf.Target {
		t.Errorf("Target mismatch: got %s, want %s", parsed.Target, wf.Target)
	}
	if len(parsed.Nodes) != len(wf.Nodes) {
		t.Errorf("Node count mismatch: got %d, want %d", len(parsed.Nodes), len(wf.Nodes))
	}
	if len(parsed.Edges) != len(wf.Edges) {
		t.Errorf("Edge count mismatch: got %d, want %d", len(parsed.Edges), len(wf.Edges))
	}

	nodeMap := make(map[string]Node)
	for _, n := range parsed.Nodes {
		nodeMap[n.ID] = n
	}

	if n, ok := nodeMap["node-1"]; ok {
		if n.Type != NodeTypeDataLoader {
			t.Errorf("node-1 type mismatch: got %s, want %s", n.Type, NodeTypeDataLoader)
		}
	} else {
		t.Error("node-1 not found in parsed workflow")
	}
}

func TestClone(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "clone-test",
		Name:          "Clone Test",
		Target:        TargetPython,
		Nodes:         []Node{},
		Edges:         []Edge{},
	}

	clone, err := Clone(wf)
	if err != nil {
		t.Fatalf("failed to clone workflow: %v", err)
	}

	if clone.ID != wf.ID {
		t.Errorf("ID mismatch: got %s, want %s", clone.ID, wf.ID)
	}

	wf.ID = "modified"
	if clone.ID == wf.ID {
		t.Error("clone should be independent of original")
	}
}

func TestValidationValid(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "valid-test",
		Name:          "Valid Workflow",
		Target:        TargetDocker,
		Nodes: []Node{
			{
				ID:   "a",
				Type: NodeTypeDataLoader,
				Name: "A",
				Outputs: []Port{
					{ID: "out", Name: "Output", Type: DataTypeDataset},
				},
			},
			{
				ID:   "b",
				Type: NodeTypeDataPreprocess,
				Name: "B",
				Inputs: []Port{
					{ID: "in", Name: "Input", Type: DataTypeDataset},
				},
				Outputs: []Port{
					{ID: "out", Name: "Output", Type: DataTypeModel},
				},
			},
			{
				ID:   "c",
				Type: NodeTypeModelTrainer,
				Name: "C",
				Inputs: []Port{
					{ID: "in", Name: "Input", Type: DataTypeModel},
				},
				Outputs: []Port{
					{ID: "out", Name: "Output", Type: DataTypeModel},
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a", PortID: "out"}, Target: EdgeEndpoint{NodeID: "b", PortID: "in"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "b", PortID: "out"}, Target: EdgeEndpoint{NodeID: "c", PortID: "in"}},
		},
	}

	result := ValidateWorkflow(wf)
	if !result.Valid {
		t.Errorf("expected valid workflow, got errors: %v", result.Errors)
	}
}

func TestValidationCycle(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "cycle-test",
		Name:          "Cycle Test",
		Target:        TargetPython,
		Nodes: []Node{
			{ID: "a", Type: NodeTypeDataLoader, Name: "A", Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
			{ID: "b", Type: NodeTypeDataPreprocess, Name: "B", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeDataset}}, Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
			{ID: "c", Type: NodeTypeModelTrainer, Name: "C", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeDataset}}, Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeModel}}},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a", PortID: "out"}, Target: EdgeEndpoint{NodeID: "b", PortID: "in"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "b", PortID: "out"}, Target: EdgeEndpoint{NodeID: "c", PortID: "in"}},
			{ID: "e3", Source: EdgeEndpoint{NodeID: "c", PortID: "out"}, Target: EdgeEndpoint{NodeID: "a", PortID: "in"}},
		},
	}

	result := ValidateWorkflow(wf)
	if result.Valid {
		t.Error("expected invalid workflow due to cycle, got valid")
	}
}

func TestValidationSelfLoop(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "self-loop-test",
		Name:          "Self Loop",
		Target:        TargetPython,
		Nodes: []Node{
			{ID: "a", Type: NodeTypeDataLoader, Name: "A", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeFile}}, Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a", PortID: "out"}, Target: EdgeEndpoint{NodeID: "a", PortID: "in"}},
		},
	}

	result := ValidateWorkflow(wf)
	if result.Valid {
		t.Error("expected invalid workflow due to self-loop, got valid")
	}
}

func TestValidationRequiredInput(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "req-input-test",
		Name:          "Required Input",
		Target:        TargetPython,
		Nodes: []Node{
			{
				ID:   "a",
				Type: NodeTypeDataLoader,
				Name: "A",
				Inputs: []Port{
					{ID: "data", Name: "Data", Type: DataTypeFile, Required: true},
				},
			},
		},
		Edges: []Edge{},
	}

	result := ValidateWorkflow(wf)
	if result.Valid {
		t.Error("expected invalid workflow due to missing required input, got valid")
	}
}

func TestFileIO(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.json")

	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "file-io-test",
		Name:          "File I/O Test",
		Target:        TargetPython,
		Nodes:         []Node{},
		Edges:         []Edge{},
	}

	if err := SaveToFile(wf, path); err != nil {
		t.Fatalf("failed to save workflow: %v", err)
	}

	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("failed to load workflow: %v", err)
	}

	if loaded.ID != wf.ID {
		t.Errorf("ID mismatch after file roundtrip: got %s, want %s", loaded.ID, wf.ID)
	}
	if loaded.Name != wf.Name {
		t.Errorf("Name mismatch after file roundtrip: got %s, want %s", loaded.Name, wf.Name)
	}
}

func TestValidateWorkflowFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.json")

	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "validate-file",
		Name:          "Validate File",
		Target:        TargetROS2,
		Nodes:         []Node{},
		Edges:         []Edge{},
	}

	if err := SaveToFile(wf, path); err != nil {
		t.Fatalf("failed to save workflow: %v", err)
	}

	if err := ValidateWorkflowFile(path); err != nil {
		t.Errorf("expected valid workflow file, got error: %v", err)
	}
}

func TestSchemaMigration(t *testing.T) {
	migrator := NewSchemaMigrator()

	wf := &Workflow{
		SchemaVersion: "1.0.0",
		ID:            "migration-test",
		Name:          "Migration Test",
		Target:        TargetPython,
	}

	migrated, err := migrator.Migrate(wf, CurrentSchemaVersion)
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	if !migrated {
		t.Error("expected migration to occur")
	}
	if wf.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("version mismatch after migration: got %s, want %s", wf.SchemaVersion, CurrentSchemaVersion)
	}
}

func TestSchemaMigrationAlreadyCurrent(t *testing.T) {
	migrator := NewSchemaMigrator()

	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "already-current",
		Name:          "Already Current",
		Target:        TargetMATLAB,
	}

	migrated, err := migrator.Migrate(wf, CurrentSchemaVersion)
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	if migrated {
		t.Error("expected no migration for already-current schema")
	}
}

func TestSchemaMigrationNoPath(t *testing.T) {
	migrator := NewSchemaMigrator()

	wf := &Workflow{
		SchemaVersion: "0.0.1",
		ID:            "no-path",
		Name:          "No Path",
		Target:        TargetPython,
	}

	_, err := migrator.Migrate(wf, CurrentSchemaVersion)
	if err == nil {
		t.Error("expected error for unsupported migration path")
	}
}

func TestTopologicalSort(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "topo-test",
		Name:          "Topo Test",
		Target:        TargetPython,
		Nodes: []Node{
			{ID: "a", Type: NodeTypeDataLoader, Name: "A", Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
			{ID: "b", Type: NodeTypeDataPreprocess, Name: "B", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeDataset}}, Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
			{ID: "c", Type: NodeTypeModelTrainer, Name: "C", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeDataset}}, Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeModel}}},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a", PortID: "out"}, Target: EdgeEndpoint{NodeID: "b", PortID: "in"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "b", PortID: "out"}, Target: EdgeEndpoint{NodeID: "c", PortID: "in"}},
		},
	}

	sorted, err := TopologicalSort(wf)
	if err != nil {
		t.Fatalf("topological sort failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(sorted))
	}

	if sorted[0].ID != "a" || sorted[1].ID != "b" || sorted[2].ID != "c" {
		t.Errorf("unexpected sort order: %v", sorted)
	}
}

func TestTopologicalSortCycle(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "topo-cycle",
		Name:          "Topo Cycle",
		Target:        TargetPython,
		Nodes: []Node{
			{ID: "a", Type: NodeTypeDataLoader, Name: "A", Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
			{ID: "b", Type: NodeTypeDataPreprocess, Name: "B", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeDataset}}, Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
			{ID: "c", Type: NodeTypeModelTrainer, Name: "C", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeDataset}}, Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeModel}}},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a", PortID: "out"}, Target: EdgeEndpoint{NodeID: "b", PortID: "in"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "b", PortID: "out"}, Target: EdgeEndpoint{NodeID: "c", PortID: "in"}},
			{ID: "e3", Source: EdgeEndpoint{NodeID: "c", PortID: "out"}, Target: EdgeEndpoint{NodeID: "a", PortID: "in"}},
		},
	}

	_, err := TopologicalSort(wf)
	if err == nil {
		t.Error("expected error for cyclic graph")
	}
}

func TestIsValidDAG(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "dag-test",
		Name:          "DAG Test",
		Target:        TargetPython,
		Nodes: []Node{
			{ID: "a", Type: NodeTypeDataLoader, Name: "A", Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
			{ID: "b", Type: NodeTypeDataPreprocess, Name: "B", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeDataset}}, Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a", PortID: "out"}, Target: EdgeEndpoint{NodeID: "b", PortID: "in"}},
		},
	}

	if !IsValidDAG(wf) {
		t.Error("expected valid DAG")
	}
}

func TestWorkflowManager(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wm.json")
	mgr := NewWorkflowManager()

	wf := mgr.CreateDefault("proj-1", "My Workflow", "python")
	if wf.ID != "proj-1" {
		t.Errorf("expected project ID proj-1, got %s", wf.ID)
	}
	if wf.Target != TargetPython {
		t.Errorf("expected target python, got %s", wf.Target)
	}

	if err := mgr.Write(wf, path); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if !mgr.Exists(path) {
		t.Error("expected file to exist")
	}

	loaded, err := mgr.Read(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if loaded.Name != "My Workflow" {
		t.Errorf("name mismatch: got %s", loaded.Name)
	}

	if err := mgr.Delete(path); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if mgr.Exists(path) {
		t.Error("expected file to be deleted")
	}
}

func TestParseNonExistentFile(t *testing.T) {
	_, err := ParseFile(filepath.Join(os.TempDir(), "nonexistent-"+randUUID()+".json"))
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestSchemaVersionConstant(t *testing.T) {
	if CurrentSchemaVersion != "2.0.0" {
		t.Errorf("CurrentSchemaVersion should be 2.0.0, got %s", CurrentSchemaVersion)
	}
}

func TestCreatedAtUpdatedAt(t *testing.T) {
	now := time.Now()
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "time-test",
		Name:          "Time Test",
		Target:        TargetPython,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	data, err := ToJSON(wf)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if parsed.CreatedAt.IsZero() {
		t.Error("created_at should not be zero")
	}
	if parsed.UpdatedAt.IsZero() {
		t.Error("updated_at should not be zero")
	}
}

func TestValidNodeTypes(t *testing.T) {
	types := ValidNodeTypes()
	if len(types) == 0 {
		t.Error("ValidNodeTypes should not be empty")
	}

	typeSet := make(map[NodeType]bool)
	for _, nt := range types {
		typeSet[nt] = true
	}

	if !typeSet[NodeTypeCondition] {
		t.Error("NodeTypeCondition should be in valid types")
	}
	if !typeSet[NodeTypeDataLoader] {
		t.Error("NodeTypeDataLoader should be in valid types")
	}
}

func TestValidTargets(t *testing.T) {
	targets := ValidTargets()
	if len(targets) == 0 {
		t.Error("ValidTargets should not be empty")
	}

	targetSet := make(map[Target]bool)
	for _, tr := range targets {
		targetSet[tr] = true
	}

	if !targetSet[TargetPython] {
		t.Error("TargetPython should be in valid targets")
	}
	if !targetSet[TargetJava] {
		t.Error("TargetJava should be in valid targets")
	}
}

func TestValidDataTypes(t *testing.T) {
	types := ValidDataTypes()
	if len(types) == 0 {
		t.Error("ValidDataTypes should not be empty")
	}

	typeSet := make(map[DataType]bool)
	for _, dt := range types {
		typeSet[dt] = true
	}

	if !typeSet[DataTypeAny] {
		t.Error("DataTypeAny should be in valid types")
	}
	if !typeSet[DataTypeImage] {
		t.Error("DataTypeImage should be in valid types")
	}
}

func TestControlConfigTypes(t *testing.T) {
	cc := ConditionConfig{
		Expression:   "x > 5",
		TrueBranch:   "branch-true",
		FalseBranch:  "branch-false",
	}
	if cc.Expression != "x > 5" {
		t.Error("ConditionConfig Expression not set correctly")
	}

	lc := LoopConfig{
		Iterations:  10,
		IteratorVar: "i",
	}
	if lc.Iterations != 10 {
		t.Error("LoopConfig Iterations not set correctly")
	}

	sc := SwitchConfig{
		Expression: "color",
		Cases: []SwitchCase{
			{Value: "red", BranchID: "branch-red"},
			{Value: "blue", BranchID: "branch-blue"},
		},
		DefaultCase: "branch-default",
	}
	if len(sc.Cases) != 2 {
		t.Error("SwitchConfig should have 2 cases")
	}

	rc := RetryConfig{
		MaxRetries: 3,
		BackoffMS:  1000,
		RetryOnAny: true,
	}
	if rc.MaxRetries != 3 {
		t.Error("RetryConfig MaxRetries not set correctly")
	}
}

func TestParseEmptyNodes(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "empty-nodes",
		Name:          "Empty Nodes",
		Target:        TargetPython,
		Nodes:         []Node{},
		Edges:         []Edge{},
	}

	data, err := ToJSON(wf)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(parsed.Nodes) != 0 {
		t.Error("expected 0 nodes")
	}
}

func TestDAGOperations(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "dag-ops",
		Name:          "DAG Ops",
		Target:        TargetPython,
		Nodes: []Node{
			{ID: "a", Type: NodeTypeDataLoader, Name: "A", Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
			{ID: "b", Type: NodeTypeDataPreprocess, Name: "B", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeDataset}}, Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
			{ID: "c", Type: NodeTypeModelTrainer, Name: "C", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeDataset}}, Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeModel}}},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a", PortID: "out"}, Target: EdgeEndpoint{NodeID: "b", PortID: "in"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "b", PortID: "out"}, Target: EdgeEndpoint{NodeID: "c", PortID: "in"}},
		},
	}

	sources := GetSourceNodes(wf)
	if len(sources) != 1 || sources[0].ID != "a" {
		t.Errorf("expected 1 source node 'a', got %v", sources)
	}

	sinks := GetSinkNodes(wf)
	if len(sinks) != 1 || sinks[0].ID != "c" {
		t.Errorf("expected 1 sink node 'c', got %v", sinks)
	}

	upstream := GetUpstreamNodes(wf, "c")
	if len(upstream) != 2 {
		t.Errorf("expected 2 upstream nodes for 'c', got %d", len(upstream))
	}

	downstream := GetDownstreamNodes(wf, "a")
	if len(downstream) != 2 {
		t.Errorf("expected 2 downstream nodes for 'a', got %d", len(downstream))
	}

	node, ok := GetNodeByID(wf, "b")
	if !ok || node.ID != "b" {
		t.Error("GetNodeByID failed to find node 'b'")
	}

	edges := GetEdgesForNode(wf, "b")
	if len(edges) != 2 {
		t.Errorf("expected 2 edges for 'b', got %d", len(edges))
	}

	inputEdges := GetInputEdges(wf, "b")
	if len(inputEdges) != 1 {
		t.Errorf("expected 1 input edge for 'b', got %d", len(inputEdges))
	}

	outputEdges := GetOutputEdges(wf, "b")
	if len(outputEdges) != 1 {
		t.Errorf("expected 1 output edge for 'b', got %d", len(outputEdges))
	}

	levels := GetExecutionLevels(wf)
	if levels == nil || len(levels) != 3 {
		t.Errorf("expected 3 execution levels, got %d", len(levels))
	}

	depth := GetNodeDepth(wf, "c")
	if depth != 2 {
		t.Errorf("expected depth 2 for 'c', got %d", depth)
	}
}

func TestGenerateJSONSchema(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "schema-test",
		Name:          "Schema Test",
		Target:        TargetPython,
	}

	schema := wf.GenerateJSONSchema()
	if schema == nil {
		t.Fatal("GenerateJSONSchema returned nil")
	}

	if schema["$schema"] != "http://json-schema.org/draft-07/schema#" {
		t.Errorf("unexpected $schema value: %v", schema["$schema"])
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be a map")
	}

	if _, ok := props["nodes"]; !ok {
		t.Error("schema should have nodes property")
	}
	if _, ok := props["edges"]; !ok {
		t.Error("schema should have edges property")
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required should be a string slice")
	}
	hasID := false
	for _, r := range required {
		if r == "id" {
			hasID = true
		}
	}
	if !hasID {
		t.Error("required should include 'id'")
	}
}

func TestSaveSchema(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema_v2.json")

	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "save-schema-test",
		Name:          "Save Schema Test",
		Target:        TargetPython,
	}

	if err := SaveSchema(wf, path); err != nil {
		t.Fatalf("SaveSchema failed: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("schema file was not created")
	}
}

func TestValidateNode(t *testing.T) {
	tests := []struct {
		name    string
		node    Node
		wantErr bool
	}{
		{
			name:    "valid data_loader",
			node:    Node{ID: "n1", Type: NodeTypeDataLoader, Name: "Loader", Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
			wantErr: false,
		},
		{
			name:    "missing ID",
			node:    Node{Type: NodeTypeDataLoader, Name: "Loader"},
			wantErr: true,
		},
		{
			name:    "missing type",
			node:    Node{ID: "n1", Name: "Loader"},
			wantErr: true,
		},
		{
			name:    "missing name",
			node:    Node{ID: "n1", Type: NodeTypeDataLoader},
			wantErr: true,
		},
		{
			name:    "invalid type",
			node:    Node{ID: "n1", Type: "invalid_type", Name: "Bad"},
			wantErr: true,
		},
		{
			name:    "model_trainer missing outputs",
			node:    Node{ID: "n1", Type: NodeTypeModelTrainer, Name: "Trainer", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeDataset}}},
			wantErr: true,
		},
		{
			name:    "condition missing expression",
			node:    Node{ID: "n1", Type: NodeTypeCondition, Name: "Cond"},
			wantErr: true,
		},
		{
			name:    "valid condition",
			node:    Node{ID: "n1", Type: NodeTypeCondition, Name: "Cond", Config: map[string]any{"expression": "x > 5"}},
			wantErr: false,
		},
		{
			name:    "valid model_trainer",
			node:    Node{ID: "n1", Type: NodeTypeModelTrainer, Name: "Trainer", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeDataset}}, Outputs: []Port{{ID: "out", Name: "Model", Type: DataTypeModel}}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNode(tt.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNode() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationDuplicateEdge(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "dup-edge-test",
		Name:          "Duplicate Edge Test",
		Target:        TargetPython,
		Nodes: []Node{
			{ID: "a", Type: NodeTypeDataLoader, Name: "A", Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
			{ID: "b", Type: NodeTypeDataPreprocess, Name: "B", Inputs: []Port{{ID: "in", Name: "Input", Type: DataTypeDataset}}, Outputs: []Port{{ID: "out", Name: "Output", Type: DataTypeDataset}}},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a", PortID: "out"}, Target: EdgeEndpoint{NodeID: "b", PortID: "in"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "a", PortID: "out"}, Target: EdgeEndpoint{NodeID: "b", PortID: "in"}},
		},
	}

	result := ValidateWorkflow(wf)
	if result.Valid {
		t.Error("expected invalid workflow due to duplicate edge, got valid")
	}

	hasDup := false
	for _, err := range result.Errors {
		if err.Code == "DUPLICATE_EDGE" {
			hasDup = true
			break
		}
	}
	if !hasDup {
		t.Errorf("expected DUPLICATE_EDGE error, got: %v", result.Errors)
	}
}

func TestYOLOWorkflowCreation(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "yolo-training-pipeline",
		Name:          "YOLO Training Pipeline",
		Description:   "End-to-end YOLO object detection training workflow",
		Version:       1,
		Author:        "AIStudio",
		Tags:          []string{"yolo", "object-detection", "training"},
		Target:        TargetPython,
		Variables: map[string]any{
			"epochs": 100,
			"device": "cuda",
		},
		Nodes: []Node{
			{
				ID:   "data-loader",
				Type: NodeTypeDataLoader,
				Name: "Load Dataset",
				Position: Point{X: 100, Y: 100},
				Config: map[string]any{
					"dataset": "coco128",
					"split":   "train",
				},
				Outputs: []Port{
					{ID: "dataset", Name: "Dataset", Type: DataTypeDataset},
				},
			},
			{
				ID:   "data-split",
				Type: NodeTypeDataSplit,
				Name: "Split Dataset",
				Position: Point{X: 300, Y: 100},
				Inputs: []Port{
					{ID: "data", Name: "Data", Type: DataTypeDataset, Required: true},
				},
				Outputs: []Port{
					{ID: "train", Name: "Train Set", Type: DataTypeDataset},
					{ID: "val", Name: "Val Set", Type: DataTypeDataset},
				},
				Config: map[string]any{
					"train_ratio": 0.8,
					"seed":        42,
				},
			},
			{
				ID:   "trainer",
				Type: NodeTypeModelTrainer,
				Name: "Train YOLO Model",
				Position: Point{X: 500, Y: 100},
				Inputs: []Port{
					{ID: "train_data", Name: "Training Data", Type: DataTypeDataset, Required: true},
					{ID: "val_data", Name: "Validation Data", Type: DataTypeDataset},
				},
				Outputs: []Port{
					{ID: "model", Name: "Trained Model", Type: DataTypeModel},
				},
				Config: map[string]any{
					"framework": "yolo",
					"model":     "yolo11n.pt",
					"epochs":    "${epochs}",
					"device":    "${device}",
					"imgsz":     640,
					"batch":     16,
				},
			},
			{
				ID:   "evaluator",
				Type: NodeTypeModelEvaluator,
				Name: "Evaluate Model",
				Position: Point{X: 700, Y: 100},
				Inputs: []Port{
					{ID: "model", Name: "Model", Type: DataTypeModel, Required: true},
					{ID: "val_data", Name: "Validation Data", Type: DataTypeDataset},
				},
				Outputs: []Port{
					{ID: "metrics", Name: "Metrics", Type: DataTypeJSON},
				},
				Config: map[string]any{
					"metrics": []string{"precision", "recall", "mAP"},
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "data-loader", PortID: "dataset"}, Target: EdgeEndpoint{NodeID: "data-split", PortID: "data"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "data-split", PortID: "train"}, Target: EdgeEndpoint{NodeID: "trainer", PortID: "train_data"}},
			{ID: "e3", Source: EdgeEndpoint{NodeID: "data-split", PortID: "val"}, Target: EdgeEndpoint{NodeID: "trainer", PortID: "val_data"}},
			{ID: "e4", Source: EdgeEndpoint{NodeID: "trainer", PortID: "model"}, Target: EdgeEndpoint{NodeID: "evaluator", PortID: "model"}},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := ValidateWorkflow(wf)
	if !result.Valid {
		t.Errorf("YOLO workflow should be valid, got errors: %v", result.Errors)
	}

	sorted, err := TopologicalSort(wf)
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}
	if len(sorted) != 4 {
		t.Errorf("expected 4 sorted nodes, got %d", len(sorted))
	}

	if sorted[0].ID != "data-loader" {
		t.Errorf("expected first node to be data-loader, got %s", sorted[0].ID)
	}

	if !IsValidDAG(wf) {
		t.Error("YOLO workflow should be a valid DAG")
	}

	sources := GetSourceNodes(wf)
	if len(sources) != 1 || sources[0].ID != "data-loader" {
		t.Errorf("expected 1 source node (data-loader), got %v", sources)
	}

	sinks := GetSinkNodes(wf)
	if len(sinks) != 1 || sinks[0].ID != "evaluator" {
		t.Errorf("expected 1 sink node (evaluator), got %v", sinks)
	}

	data, err := ToJSON(wf)
	if err != nil {
		t.Fatalf("failed to marshal YOLO workflow: %v", err)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("failed to parse YOLO workflow: %v", err)
	}
	if len(parsed.Nodes) != 4 {
		t.Errorf("node count mismatch after roundtrip: got %d, want 4", len(parsed.Nodes))
	}
}

// randUUID generates a UUID string for test isolation.
func randUUID() string {
	b := make([]byte, 16)
	// Simple random UUID v4-like without importing crypto/rand for test helper
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
