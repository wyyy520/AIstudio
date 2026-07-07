package workflow

import (
	"encoding/json"
	"fmt"
)

func ParseWorkflow(data []byte) (*Workflow, error) {
	var wf Workflow
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("failed to parse workflow JSON: %w", err)
	}
	return &wf, nil
}

func ValidateWorkflow(wf *Workflow) error {
	if wf.ID == "" {
		return fmt.Errorf("workflow id is required")
	}
	if wf.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	if len(wf.Nodes) == 0 {
		return fmt.Errorf("workflow must contain at least one node")
	}

	nodeIDs := make(map[string]bool)
	for _, node := range wf.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node id is required")
		}
		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node id: %s", node.ID)
		}
		nodeIDs[node.ID] = true
	}

	for _, node := range wf.Nodes {
		if !DefaultRegistry.Has(node.Type, node.Plugin) {
			return fmt.Errorf("node %s: type %q with plugin %q is not registered", node.ID, node.Type, node.Plugin)
		}
	}

	if err := ValidateEdges(wf.Nodes, wf.Edges); err != nil {
		return err
	}

	if _, err := TopologicalSort(wf.Nodes, wf.Edges); err != nil {
		return fmt.Errorf("DAG validation failed: %w", err)
	}

	return nil
}

func ParseAndValidate(data []byte) (*Workflow, error) {
	wf, err := ParseWorkflow(data)
	if err != nil {
		return nil, err
	}
	if err := ValidateWorkflow(wf); err != nil {
		return nil, err
	}
	return wf, nil
}
