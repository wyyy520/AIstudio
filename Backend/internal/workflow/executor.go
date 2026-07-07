package workflow

import (
	"context"
	"fmt"
	"time"
)

type Executor struct {
	registry *NodeRegistry
}

func NewExecutor(registry *NodeRegistry) *Executor {
	return &Executor{registry: registry}
}

func (e *Executor) Execute(ctx context.Context, wf *Workflow) (*ExecutionResult, error) {
	order, err := TopologicalSort(wf.Nodes, wf.Edges)
	if err != nil {
		return nil, fmt.Errorf("failed to generate execution plan: %w", err)
	}

	nodeMap := make(map[string]*Node)
	for i := range wf.Nodes {
		nodeMap[wf.Nodes[i].ID] = &wf.Nodes[i]
	}

	outgoingEdges := make(map[string][]Edge)
	for _, edge := range wf.Edges {
		outgoingEdges[edge.Source.NodeID] = append(outgoingEdges[edge.Source.NodeID], edge)
	}

	incomingEdges := make(map[string][]Edge)
	for _, edge := range wf.Edges {
		incomingEdges[edge.Target.NodeID] = append(incomingEdges[edge.Target.NodeID], edge)
	}

	nodeOutputs := make(map[string]map[string]interface{})
	result := &ExecutionResult{
		TaskID:      fmt.Sprintf("task_%d", time.Now().UnixNano()),
		Status:      NodeStatusRunning,
		NodeOutputs: make(map[string]NodeResult),
	}

	progressPerNode := 1.0 / float64(len(order))
	accumulatedProgress := 0.0

	for _, nodeID := range order {
		node, ok := nodeMap[nodeID]
		if !ok {
			return nil, fmt.Errorf("node %s not found in workflow", nodeID)
		}

		if node.Runtime == nil {
			node.Runtime = &NodeRuntime{}
		}

		node.Runtime.Status = NodeStatusRunning
		now := time.Now()
		node.Runtime.StartedAt = &now

		inputs := make(map[string]interface{})
		for _, edge := range incomingEdges[nodeID] {
			srcOutputs, hasOutput := nodeOutputs[edge.Source.NodeID]
			if !hasOutput {
				return nil, fmt.Errorf("node %s: source node %s has no output", nodeID, edge.Source.NodeID)
			}
			outputVal, hasVal := srcOutputs[edge.Source.PortID]
			if !hasVal {
				return nil, fmt.Errorf("node %s: source node %s has no output on port %s", nodeID, edge.Source.NodeID, edge.Source.PortID)
			}
			inputs[edge.Target.PortID] = outputVal
		}

		execNode, err := e.registry.CreateExecutable(node.Type, node.Plugin)
		if err != nil {
			node.Runtime.Status = NodeStatusError
			errMsg := err.Error()
			node.Runtime.Error = errMsg
			result.NodeOutputs[nodeID] = NodeResult{
				Status: NodeStatusError,
				Error:  errMsg,
			}
			result.Status = NodeStatusError
			result.Error = fmt.Sprintf("node %s: %s", nodeID, errMsg)
			return result, nil
		}

		startTime := time.Now()

		outputs, execErr := execNode.Execute(ctx, inputs, node.Parameters)

		duration := time.Since(startTime).Milliseconds()
		finishTime := time.Now()
		node.Runtime.FinishedAt = &finishTime
		node.Runtime.DurationMs = &duration

		if execErr != nil {
			node.Runtime.Status = NodeStatusError
			errMsg := execErr.Error()
			node.Runtime.Error = errMsg
			result.NodeOutputs[nodeID] = NodeResult{
				Status:   NodeStatusError,
				Duration: duration,
				Error:    errMsg,
			}
			result.Status = NodeStatusError
			result.Error = fmt.Sprintf("node %s: %s", nodeID, errMsg)
			return result, nil
		}

		node.Runtime.Status = NodeStatusSuccess
		node.Runtime.OutputSnapshot = outputs

		for _, port := range node.Outputs {
			if _, exists := outputs[port.ID]; !exists {
				node.Runtime.Status = NodeStatusError
				errMsg := fmt.Sprintf("missing output port %s (expected from node %s)", port.ID, nodeID)
				node.Runtime.Error = errMsg
				result.NodeOutputs[nodeID] = NodeResult{
					Status:   NodeStatusError,
					Duration: duration,
					Error:    errMsg,
				}
				result.Status = NodeStatusError
				result.Error = errMsg
				return result, nil
			}
		}

		nodeOutputs[nodeID] = outputs

		accumulatedProgress += progressPerNode
		result.Progress = accumulatedProgress

		result.NodeOutputs[nodeID] = NodeResult{
			Status:   NodeStatusSuccess,
			Outputs:  outputs,
			Duration: duration,
		}
	}

	result.Status = NodeStatusSuccess
	result.Progress = 1.0

	return result, nil
}
