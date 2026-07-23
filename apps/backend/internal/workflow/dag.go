// Package workflow provides DAG operations via type aliases to packages/workflow.
//
// All DAG functions (TopologicalSort, GetUpstreamNodes, GetDownstreamNodes,
// GetSourceNodes, GetSinkNodes, GetNodeByID, GetEdgesForNode, GetInputEdges,
// GetOutputEdges, IsValidDAG, GetNodeDepth, GetExecutionLevels, makeNodeMap)
// have been moved to packages/workflow and are re-exported from types.go.
package workflow
