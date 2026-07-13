package sdk

import (
	"github.com/aistudio/packages/workflow"
	"github.com/google/uuid"
)

type Workflow = workflow.Workflow
type Node = workflow.Node
type Edge = workflow.Edge
type NodeType = workflow.NodeType
type Target = workflow.Target
type ValidationResult = workflow.ValidationResult

func NewWorkflow(name string, target Target) *Workflow {
	return &Workflow{
		SchemaVersion: workflow.CurrentSchemaVersion,
		ID:            uuid.New().String(),
		Name:          name,
		Version:       1,
		Target:        target,
		Nodes:         make([]Node, 0),
		Edges:         make([]Edge, 0),
	}
}

func AddNode(wf *Workflow, nodeType NodeType, config map[string]any) string {
	id := uuid.New().String()
	wf.Nodes = append(wf.Nodes, Node{
		ID:     id,
		Type:   nodeType,
		Config: config,
	})
	return id
}

func AddEdge(wf *Workflow, source, target string) string {
	id := uuid.New().String()
	wf.Edges = append(wf.Edges, Edge{
		ID: id,
		Source: workflow.EdgeEndpoint{NodeID: source},
		Target: workflow.EdgeEndpoint{NodeID: target},
	})
	return id
}

func Validate(wf *Workflow) error {
	result := workflow.ValidateWorkflow(wf)
	if !result.Valid {
		if len(result.Errors) > 0 {
			return result.Errors[0]
		}
	}
	return nil
}

func SaveWorkflow(wf *Workflow, path string) error {
	return workflow.Save(wf, path)
}

func LoadWorkflow(path string) (*Workflow, error) {
	return workflow.LoadFromFile(path)
}