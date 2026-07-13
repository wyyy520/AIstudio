package workflow

import (
	"fmt"
)

// ============================================================================
// Workflow Validator
//
// Validates the workflow DAG structure and data flow consistency.
// ============================================================================

// ValidationError represents a workflow validation error.
type ValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	NodeID  string `json:"nodeId,omitempty"`
	EdgeID  string `json:"edgeId,omitempty"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s (node=%s, edge=%s)", e.Code, e.Message, e.NodeID, e.EdgeID)
}

// ValidationResult contains the result of workflow validation.
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []*ValidationError `json:"errors,omitempty"`
	Warnings []string        `json:"warnings,omitempty"`
}

// Validate performs comprehensive validation of a workflow.
// This includes DAG structure, data types, and configuration.
func Validate(wf *Workflow) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]*ValidationError, 0),
		Warnings: make([]string, 0),
	}

	// 1. Validate DAG (no cycles)
	if err := validateDAG(wf); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// 2. Validate data types
	validateDataTypes(wf, result)

	// 3. Validate required inputs
	validateRequiredInputs(wf, result)

	// 4. Validate node types
	validateNodeTypes(wf, result)

	// 5. Validate variable references
	validateVariables(wf, result)

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}

// validateDAG checks for cycles in the workflow graph.
func validateDAG(wf *Workflow) *ValidationError {
	// Build adjacency list
	graph := make(map[string][]string)
	for _, edge := range wf.Edges {
		graph[edge.Source.NodeID] = append(graph[edge.Source.NodeID], edge.Target.NodeID)
	}

	// Check for self-loops
	for _, edge := range wf.Edges {
		if edge.Source.NodeID == edge.Target.NodeID {
			return &ValidationError{
				Code:    "SELF_LOOP",
				Message: "self-loop detected",
				EdgeID:  edge.ID,
				NodeID:  edge.Source.NodeID,
			}
		}
	}

	// DFS-based cycle detection
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(nodeID string) bool
	dfs = func(nodeID string) bool {
		visited[nodeID] = true
		recStack[nodeID] = true

		for _, neighbor := range graph[nodeID] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}

		recStack[nodeID] = false
		return false
	}

	for _, node := range wf.Nodes {
		if !visited[node.ID] {
			if dfs(node.ID) {
				return &ValidationError{
					Code:    "CYCLE_DETECTED",
					Message: "workflow graph contains a cycle",
				}
			}
		}
	}

	return nil
}

// validateDataTypes checks data type compatibility along edges.
func validateDataTypes(wf *Workflow, result *ValidationResult) {
	// Build port lookup
	nodePorts := make(map[string]map[string]Port)
	for _, node := range wf.Nodes {
		ports := make(map[string]Port)
		for _, p := range node.Inputs {
			ports[p.ID] = p
		}
		for _, p := range node.Outputs {
			ports[p.ID] = p
		}
		nodePorts[node.ID] = ports
	}

	for _, edge := range wf.Edges {
		sourcePorts, ok := nodePorts[edge.Source.NodeID]
		if !ok {
			continue
		}
		targetPorts, ok := nodePorts[edge.Target.NodeID]
		if !ok {
			continue
		}

		sourcePort, ok := sourcePorts[edge.Source.PortID]
		if !ok {
			result.Errors = append(result.Errors, &ValidationError{
				Code:    "PORT_NOT_FOUND",
				Message: fmt.Sprintf("source port %s not found on node %s", edge.Source.PortID, edge.Source.NodeID),
				EdgeID:  edge.ID,
				NodeID:  edge.Source.NodeID,
			})
			continue
		}

		targetPort, ok := targetPorts[edge.Target.PortID]
		if !ok {
			result.Errors = append(result.Errors, &ValidationError{
				Code:    "PORT_NOT_FOUND",
				Message: fmt.Sprintf("target port %s not found on node %s", edge.Target.PortID, edge.Target.NodeID),
				EdgeID:  edge.ID,
				NodeID:  edge.Target.NodeID,
			})
			continue
		}

		if !isTypeCompatible(sourcePort.Type, targetPort.Type) {
			result.Errors = append(result.Errors, &ValidationError{
				Code:    "TYPE_MISMATCH",
				Message: fmt.Sprintf("type mismatch: %s -> %s", sourcePort.Type, targetPort.Type),
				EdgeID:  edge.ID,
			})
		}
	}
}

// validateRequiredInputs checks that all required inputs are connected.
func validateRequiredInputs(wf *Workflow, result *ValidationResult) {
	// Build set of connected input ports
	connectedInputs := make(map[string]bool) // key: "nodeID:portID"
	for _, edge := range wf.Edges {
		key := edge.Target.NodeID + ":" + edge.Target.PortID
		connectedInputs[key] = true
	}

	for _, node := range wf.Nodes {
		for _, port := range node.Inputs {
			if port.Required {
				key := node.ID + ":" + port.ID
				if !connectedInputs[key] {
					result.Errors = append(result.Errors, &ValidationError{
						Code:    "REQUIRED_INPUT_NOT_CONNECTED",
						Message: fmt.Sprintf("required input port %s (%s) is not connected", port.Name, port.ID),
						NodeID:  node.ID,
					})
				}
			}
		}
	}
}

// validateNodeTypes checks that node types are valid.
func validateNodeTypes(wf *Workflow, result *ValidationResult) {
	validTypes := make(map[NodeType]bool)
	for _, t := range ValidNodeTypes() {
		validTypes[t] = true
	}

	for _, node := range wf.Nodes {
		if !validTypes[node.Type] {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("node %s (%s): unknown node type '%s'", node.Name, node.ID, node.Type))
		}
	}
}

// validateVariables checks that variable references are valid.
func validateVariables(wf *Workflow, result *ValidationResult) {
	// Check referenced variables exist
	for _, node := range wf.Nodes {
		for key, val := range node.Config {
			if str, ok := val.(string); ok {
				// Check for ${variable} references
				// Basic check — just look for ${...} pattern
				if len(str) > 3 && str[0] == '$' && str[1] == '{' {
					varName := str[2 : len(str)-1]
					if _, exists := wf.Variables[varName]; !exists {
						result.Warnings = append(result.Warnings,
							fmt.Sprintf("node %s (%s): config '%s' references undefined variable '%s'",
								node.Name, node.ID, key, varName))
					}
				}
			}
		}
	}
}

// typeCompatibility defines which data types are compatible.
var typeCompatibility = map[DataType][]DataType{
	DataTypeImage:   {DataTypeImage, DataTypeTensor, DataTypeAny},
	DataTypeTensor:  {DataTypeTensor, DataTypeModel, DataTypeAny},
	DataTypeDataset: {DataTypeDataset, DataTypeAny},
	DataTypeModel:   {DataTypeModel, DataTypeAny},
	DataTypeText:    {DataTypeText, DataTypeJSON, DataTypeAny},
	DataTypeNumber:  {DataTypeNumber, DataTypeJSON, DataTypeAny},
	DataTypeBoolean: {DataTypeBoolean, DataTypeAny},
	DataTypeJSON:    {DataTypeJSON, DataTypeText, DataTypeAny},
	DataTypeFile:    {DataTypeFile, DataTypeAny},
	DataTypeStream:  {DataTypeStream, DataTypeAny},
	DataTypeAny:     {DataTypeAny},
}

// isTypeCompatible checks if sourceType can be connected to targetType.
func isTypeCompatible(sourceType, targetType DataType) bool {
	compatible, ok := typeCompatibility[sourceType]
	if !ok {
		return sourceType == targetType
	}
	for _, t := range compatible {
		if t == targetType {
			return true
		}
	}
	return sourceType == targetType
}