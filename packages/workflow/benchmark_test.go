package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	wf := createBenchWorkflow()
	data, err := json.Marshal(wf)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFile(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "bench.json")
	wf := createBenchWorkflow()
	if err := SaveDirect(wf, path); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseFile(path)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateWorkflow(b *testing.B) {
	wf := createBenchWorkflow()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateWorkflow(wf)
	}
}

func BenchmarkValidateNode(b *testing.B) {
	node := Node{
		ID:   "n1",
		Type: NodeTypeDataLoader,
		Name: "Loader",
		Outputs: []Port{
			{ID: "out", Name: "Output", Type: DataTypeDataset},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateNode(node)
	}
}

func BenchmarkSerialize(b *testing.B) {
	wf := createBenchWorkflow()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ToJSON(wf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSaveFile(b *testing.B) {
	dir := b.TempDir()
	wf := createBenchWorkflow()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := filepath.Join(dir, "bench_save.json")
		if err := SaveDirect(wf, path); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTopologicalSort(b *testing.B) {
	wf := createBenchWorkflow()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := TopologicalSort(wf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidNodeTypes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ValidNodeTypes()
	}
}

func BenchmarkValidNodeTypesMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ValidNodeTypesMap()
	}
}

func BenchmarkClone(b *testing.B) {
	wf := createBenchWorkflow()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Clone(wf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetExecutionLevels(b *testing.B) {
	wf := createBenchWorkflow()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetExecutionLevels(wf)
	}
}

func BenchmarkIsValidDAG(b *testing.B) {
	wf := createBenchWorkflow()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidDAG(wf)
	}
}

func BenchmarkValidateSchema(b *testing.B) {
	wf := createBenchWorkflow()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateSchema(wf)
	}
}

func BenchmarkValidateDataTypes(b *testing.B) {
	wf := createBenchWorkflow()
	result := &ValidationResult{Valid: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateDataTypes(wf, result)
	}
}

func createBenchWorkflow() *Workflow {
	nodes := make([]Node, 20)
	edges := make([]Edge, 0, 19)
	for i := 0; i < 20; i++ {
		nodes[i] = Node{
			ID:   fmt.Sprintf("node-%d", i),
			Type: NodeTypeDataLoader,
			Name: "Node",
			Inputs: []Port{
				{ID: "in", Name: "Input", Type: DataTypeDataset, Required: true},
			},
			Outputs: []Port{
				{ID: "out", Name: "Output", Type: DataTypeDataset},
			},
		}
	}
	for i := 0; i < 19; i++ {
		edges = append(edges, Edge{
			ID:     fmt.Sprintf("edge-%d", i),
			Source: EdgeEndpoint{NodeID: fmt.Sprintf("node-%d", i), PortID: "out"},
			Target: EdgeEndpoint{NodeID: fmt.Sprintf("node-%d", i+1), PortID: "in"},
		})
	}
	return &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "benchmark-workflow",
		Name:          "Benchmark Workflow",
		Target:        TargetPython,
		Nodes:         nodes,
		Edges:         edges,
	}
}

func BenchmarkAllocations(b *testing.B) {
	b.Run("Parse", func(b *testing.B) {
		wf := createBenchWorkflow()
		data, _ := json.Marshal(wf)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			Parse(data)
		}
	})
	b.Run("Validate", func(b *testing.B) {
		wf := createBenchWorkflow()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ValidateWorkflow(wf)
		}
	})
}

func init() {
	// Ensure test data is accessible
	_ = os.MkdirAll(benchTempDir(), 0755)
}

func benchTempDir() string {
	return filepath.Join(os.TempDir(), "aistudio-bench")
}
