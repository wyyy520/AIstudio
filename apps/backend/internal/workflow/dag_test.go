package workflow

import (
	"testing"
)

func TestTopologicalSortSimple(t *testing.T) {
	wf := &Workflow{
		SchemaVersion: "2.0.0",
		ID:            "test-sort-001",
		Name:          "test-sort",
		Target:        "python",
		Nodes: []Node{
			{ID: "n1", Type: "data.loader", Name: "a"},
			{ID: "n2", Type: "data.loader", Name: "b"},
			{ID: "n3", Type: "data.loader", Name: "c"},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "n1"}, Target: EdgeEndpoint{NodeID: "n2"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "n2"}, Target: EdgeEndpoint{NodeID: "n3"}},
		},
	}

	sorted, err := TopologicalSort(wf)
	if err != nil {
		t.Fatalf("TopologicalSort() failed: %v", err)
	}
	if len(sorted) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(sorted))
	}

	if sorted[0].ID != "n1" {
		t.Errorf("expected n1 first, got %s", sorted[0].ID)
	}
	if sorted[1].ID != "n2" {
		t.Errorf("expected n2 second, got %s", sorted[1].ID)
	}
	if sorted[2].ID != "n3" {
		t.Errorf("expected n3 third, got %s", sorted[2].ID)
	}
}

func TestTopologicalSortMultipleSources(t *testing.T) {
	wf := &Workflow{
		ID:     "test-multi-source",
		Name:   "multi-source",
		Target: "python",
		Nodes: []Node{
			{ID: "a", Type: "data.loader", Name: "a"},
			{ID: "b", Type: "data.loader", Name: "b"},
			{ID: "c", Type: "data.loader", Name: "c"},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a"}, Target: EdgeEndpoint{NodeID: "c"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "b"}, Target: EdgeEndpoint{NodeID: "c"}},
		},
	}

	sorted, err := TopologicalSort(wf)
	if err != nil {
		t.Fatalf("TopologicalSort() failed: %v", err)
	}
	if len(sorted) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(sorted))
	}

	if sorted[0].ID != "a" && sorted[0].ID != "b" {
		t.Errorf("expected source node (a or b) first, got %s", sorted[0].ID)
	}
	if sorted[2].ID != "c" {
		t.Errorf("expected c last, got %s", sorted[2].ID)
	}
}

func TestTopologicalSortNoEdges(t *testing.T) {
	wf := &Workflow{
		ID:     "test-no-edges",
		Name:   "no-edges",
		Target: "python",
		Nodes: []Node{
			{ID: "x", Type: "data.loader", Name: "x"},
			{ID: "y", Type: "data.loader", Name: "y"},
		},
	}

	sorted, err := TopologicalSort(wf)
	if err != nil {
		t.Fatalf("TopologicalSort() failed: %v", err)
	}
	if len(sorted) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(sorted))
	}
}

func TestTopologicalSortSingleNode(t *testing.T) {
	wf := &Workflow{
		ID:     "test-single",
		Name:   "single",
		Target: "python",
		Nodes: []Node{
			{ID: "only", Type: "data.loader", Name: "only"},
		},
	}

	sorted, err := TopologicalSort(wf)
	if err != nil {
		t.Fatalf("TopologicalSort() failed: %v", err)
	}
	if len(sorted) != 1 {
		t.Fatalf("expected 1 node, got %d", len(sorted))
	}
	if sorted[0].ID != "only" {
		t.Errorf("expected node 'only', got '%s'", sorted[0].ID)
	}
}

func TestCycleDetectionThreeNode(t *testing.T) {
	wf := &Workflow{
		ID:     "test-cycle-3",
		Name:   "cycle-3",
		Target: "python",
		Nodes: []Node{
			{ID: "a", Type: "data.loader", Name: "a"},
			{ID: "b", Type: "data.loader", Name: "b"},
			{ID: "c", Type: "data.loader", Name: "c"},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a"}, Target: EdgeEndpoint{NodeID: "b"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "b"}, Target: EdgeEndpoint{NodeID: "c"}},
			{ID: "e3", Source: EdgeEndpoint{NodeID: "c"}, Target: EdgeEndpoint{NodeID: "a"}},
		},
	}

	_, err := TopologicalSort(wf)
	if err == nil {
		t.Fatal("expected cycle detection error, got nil")
	}

	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	if valErr.Code != "CYCLE_DETECTED" {
		t.Errorf("expected code CYCLE_DETECTED, got %s", valErr.Code)
	}
}

func TestCycleDetectionSelfLoop(t *testing.T) {
	wf := &Workflow{
		ID:     "test-self-loop",
		Name:   "self-loop",
		Target: "python",
		Nodes: []Node{
			{ID: "a", Type: "data.loader", Name: "a"},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a"}, Target: EdgeEndpoint{NodeID: "a"}},
		},
	}

	result := Validate(wf)
	if result.Valid {
		t.Fatal("expected invalid workflow for self-loop")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected at least one validation error")
	}
	foundSelfLoop := false
	for _, e := range result.Errors {
		if e.Code == "SELF_LOOP" {
			foundSelfLoop = true
			break
		}
	}
	if !foundSelfLoop {
		t.Errorf("expected SELF_LOOP error, got: %v", result.Errors)
	}
}

func TestPortValidation(t *testing.T) {
	wf := &Workflow{
		ID:     "test-port-validation",
		Name:   "port-val",
		Target: "python",
		Nodes: []Node{
			{
				ID:   "n1",
				Type: "data.loader",
				Name: "source",
				Outputs: []Port{
					{ID: "out", Name: "Output", Type: DataTypeImage},
				},
			},
			{
				ID:   "n2",
				Type: "training.train",
				Name: "target",
				Inputs: []Port{
					{ID: "in", Name: "Input", Type: DataTypeDataset, Required: true},
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "n1", PortID: "out"}, Target: EdgeEndpoint{NodeID: "n2", PortID: "in"}},
		},
	}

	result := Validate(wf)
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			t.Logf("validation error: %s", e.Message)
		}
	}
}

func TestPortNotFound(t *testing.T) {
	wf := &Workflow{
		ID:     "test-port-not-found",
		Name:   "port-not-found",
		Target: "python",
		Nodes: []Node{
			{
				ID:   "n1",
				Type: "data.loader",
				Name: "source",
				Outputs: []Port{
					{ID: "out", Name: "Output", Type: DataTypeAny},
				},
			},
			{
				ID:   "n2",
				Type: "training.train",
				Name: "target",
				Inputs: []Port{
					{ID: "in", Name: "Input", Type: DataTypeAny},
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "n1", PortID: "nonexistent"}, Target: EdgeEndpoint{NodeID: "n2", PortID: "in"}},
		},
	}

	result := Validate(wf)
	foundPortError := false
	for _, e := range result.Errors {
		if e.Code == "PORT_NOT_FOUND" {
			foundPortError = true
			break
		}
	}
	if !foundPortError {
		t.Error("expected PORT_NOT_FOUND error for nonexistent source port")
	}
}

func TestIsValidDAG(t *testing.T) {
	validWF := &Workflow{
		ID:     "valid-dag",
		Name:   "valid",
		Target: "python",
		Nodes: []Node{
			{ID: "a", Type: "data.loader", Name: "a"},
			{ID: "b", Type: "data.loader", Name: "b"},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a"}, Target: EdgeEndpoint{NodeID: "b"}},
		},
	}
	if !IsValidDAG(validWF) {
		t.Error("expected valid DAG")
	}

	cycleWF := &Workflow{
		ID:     "invalid-dag",
		Name:   "invalid",
		Target: "python",
		Nodes: []Node{
			{ID: "a", Type: "data.loader", Name: "a"},
			{ID: "b", Type: "data.loader", Name: "b"},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a"}, Target: EdgeEndpoint{NodeID: "b"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "b"}, Target: EdgeEndpoint{NodeID: "a"}},
		},
	}
	if IsValidDAG(cycleWF) {
		t.Error("expected invalid DAG")
	}
}

func TestGetUpstreamDownstream(t *testing.T) {
	wf := &Workflow{
		ID:     "test-up-down-stream",
		Name:   "up-down",
		Target: "python",
		Nodes: []Node{
			{ID: "a", Type: "data.loader", Name: "a"},
			{ID: "b", Type: "data.loader", Name: "b"},
			{ID: "c", Type: "data.loader", Name: "c"},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a"}, Target: EdgeEndpoint{NodeID: "b"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "b"}, Target: EdgeEndpoint{NodeID: "c"}},
		},
	}

	upstream := GetUpstreamNodes(wf, "c")
	if len(upstream) != 2 {
		t.Errorf("expected 2 upstream nodes for c, got %d", len(upstream))
	}

	downstream := GetDownstreamNodes(wf, "a")
	if len(downstream) != 2 {
		t.Errorf("expected 2 downstream nodes for a, got %d", len(downstream))
	}
}

func TestGetSourceSinkNodes(t *testing.T) {
	wf := &Workflow{
		ID:     "test-source-sink",
		Name:   "source-sink",
		Target: "python",
		Nodes: []Node{
			{ID: "a", Type: "data.loader", Name: "a"},
			{ID: "b", Type: "data.loader", Name: "b"},
			{ID: "c", Type: "data.loader", Name: "c"},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a"}, Target: EdgeEndpoint{NodeID: "b"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "b"}, Target: EdgeEndpoint{NodeID: "c"}},
		},
	}

	sources := GetSourceNodes(wf)
	if len(sources) != 1 {
		t.Errorf("expected 1 source node, got %d", len(sources))
	} else if sources[0].ID != "a" {
		t.Errorf("expected source 'a', got '%s'", sources[0].ID)
	}

	sinks := GetSinkNodes(wf)
	if len(sinks) != 1 {
		t.Errorf("expected 1 sink node, got %d", len(sinks))
	} else if sinks[0].ID != "c" {
		t.Errorf("expected sink 'c', got '%s'", sinks[0].ID)
	}
}

func TestGetExecutionLevels(t *testing.T) {
	wf := &Workflow{
		ID:     "test-levels",
		Name:   "levels",
		Target: "python",
		Nodes: []Node{
			{ID: "a", Type: "data.loader", Name: "a"},
			{ID: "b", Type: "data.loader", Name: "b"},
			{ID: "c", Type: "data.loader", Name: "c"},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a"}, Target: EdgeEndpoint{NodeID: "b"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "b"}, Target: EdgeEndpoint{NodeID: "c"}},
		},
	}

	levels := GetExecutionLevels(wf)
	if levels == nil {
		t.Fatal("expected non-nil levels")
	}
	if len(levels) != 3 {
		t.Fatalf("expected 3 levels, got %d", len(levels))
	}
	if len(levels[0]) != 1 || levels[0][0].ID != "a" {
		t.Errorf("expected level 0 with node a")
	}
	if len(levels[2]) != 1 || levels[2][0].ID != "c" {
		t.Errorf("expected level 2 with node c")
	}
}

func TestGetNodeDepth(t *testing.T) {
	wf := &Workflow{
		ID:     "test-depth",
		Name:   "depth",
		Target: "python",
		Nodes: []Node{
			{ID: "a", Type: "data.loader", Name: "a"},
			{ID: "b", Type: "data.loader", Name: "b"},
			{ID: "c", Type: "data.loader", Name: "c"},
		},
		Edges: []Edge{
			{ID: "e1", Source: EdgeEndpoint{NodeID: "a"}, Target: EdgeEndpoint{NodeID: "b"}},
			{ID: "e2", Source: EdgeEndpoint{NodeID: "b"}, Target: EdgeEndpoint{NodeID: "c"}},
		},
	}

	if depth := GetNodeDepth(wf, "a"); depth != 0 {
		t.Errorf("expected depth 0 for source, got %d", depth)
	}
	if depth := GetNodeDepth(wf, "b"); depth != 1 {
		t.Errorf("expected depth 1 for node b, got %d", depth)
	}
	if depth := GetNodeDepth(wf, "c"); depth != 2 {
		t.Errorf("expected depth 2 for node c, got %d", depth)
	}
}
