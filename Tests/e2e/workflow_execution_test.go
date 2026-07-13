package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/aistudio/backend/internal/workflow"
)

func TestWorkflowExecutionSimple(t *testing.T) {
	wf := &workflow.Workflow{
		SchemaVersion: "2.0.0",
		ID:            "exec-test-001",
		Name:          "exec-test",
		Version:       1,
		Target:        "python",
		Nodes: []workflow.Node{
			{ID: "n1", Type: "data.loader", Name: "load", Position: workflow.Point{X: 0, Y: 0}, Outputs: []workflow.Port{{ID: "dataset", Name: "out", Type: "any"}}},
			{ID: "n2", Type: "training.train", Name: "train", Position: workflow.Point{X: 200, Y: 0}, Inputs: []workflow.Port{{ID: "data", Name: "in", Type: "any", Required: true}}},
		},
		Edges: []workflow.Edge{
			{ID: "e1", Source: workflow.EdgeEndpoint{NodeID: "n1", PortID: "dataset"}, Target: workflow.EdgeEndpoint{NodeID: "n2", PortID: "data"}},
		},
	}

	engine := workflow.NewEngineWithBuiltIns()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := engine.RunWorkflow(ctx, wf)
	if err != nil {
		t.Fatalf("workflow execution failed: %v", err)
	}

	if result.Status != "completed" {
		t.Errorf("expected status completed, got %s", result.Status)
	}
	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}
	if result.NodeResults == nil {
		t.Fatal("expected node results, got nil")
	}
	if len(result.NodeResults) != 2 {
		t.Errorf("expected 2 node results, got %d", len(result.NodeResults))
	}
}

func TestWorkflowExecutionOrder(t *testing.T) {
	wf := &workflow.Workflow{
		SchemaVersion: "2.0.0",
		ID:            "exec-order-001",
		Name:          "exec-order",
		Version:       1,
		Target:        "python",
		Nodes: []workflow.Node{
			{ID: "n1", Type: "data.loader", Name: "first", Position: workflow.Point{X: 0, Y: 0}},
			{ID: "n2", Type: "data.loader", Name: "second", Position: workflow.Point{X: 200, Y: 0}},
			{ID: "n3", Type: "data.loader", Name: "third", Position: workflow.Point{X: 400, Y: 0}},
		},
		Edges: []workflow.Edge{
			{ID: "e1", Source: workflow.EdgeEndpoint{NodeID: "n1", PortID: "dataset"}, Target: workflow.EdgeEndpoint{NodeID: "n2", PortID: "data"}},
			{ID: "e2", Source: workflow.EdgeEndpoint{NodeID: "n2", PortID: "dataset"}, Target: workflow.EdgeEndpoint{NodeID: "n3", PortID: "data"}},
		},
	}

	sorted, err := workflow.TopologicalSort(wf)
	if err != nil {
		t.Fatalf("topological sort failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Fatalf("expected 3 sorted nodes, got %d", len(sorted))
	}

	order := make([]string, len(sorted))
	for i, n := range sorted {
		order[i] = n.ID
	}

	if order[0] != "n1" {
		t.Errorf("expected first node n1, got %s", order[0])
	}
	if order[1] != "n2" {
		t.Errorf("expected second node n2, got %s", order[1])
	}
	if order[2] != "n3" {
		t.Errorf("expected third node n3, got %s", order[2])
	}
}

func TestWorkflowExecutionCycle(t *testing.T) {
	wf := &workflow.Workflow{
		SchemaVersion: "2.0.0",
		ID:            "cycle-test-001",
		Name:          "cycle-test",
		Version:       1,
		Target:        "python",
		Nodes: []workflow.Node{
			{ID: "n1", Type: "data.loader", Name: "a", Position: workflow.Point{X: 0, Y: 0}},
			{ID: "n2", Type: "data.loader", Name: "b", Position: workflow.Point{X: 200, Y: 0}},
			{ID: "n3", Type: "data.loader", Name: "c", Position: workflow.Point{X: 400, Y: 0}},
		},
		Edges: []workflow.Edge{
			{ID: "e1", Source: workflow.EdgeEndpoint{NodeID: "n1", PortID: "dataset"}, Target: workflow.EdgeEndpoint{NodeID: "n2", PortID: "data"}},
			{ID: "e2", Source: workflow.EdgeEndpoint{NodeID: "n2", PortID: "dataset"}, Target: workflow.EdgeEndpoint{NodeID: "n3", PortID: "data"}},
			{ID: "e3", Source: workflow.EdgeEndpoint{NodeID: "n3", PortID: "dataset"}, Target: workflow.EdgeEndpoint{NodeID: "n1", PortID: "data"}},
		},
	}

	_, err := workflow.TopologicalSort(wf)
	if err == nil {
		t.Fatal("expected cycle detection error, got nil")
	}
}
