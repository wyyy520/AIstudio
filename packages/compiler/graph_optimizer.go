// Package compiler provides the Graph Optimizer for workflow DAG optimization.
//
// Graph Optimizer performs static analysis on the workflow DAG to:
// - Eliminate dead nodes (nodes with no path to output)
// - Remove invalid edges (edges referencing non-existent nodes)
// - Remove duplicate edges
// - Fuse same-type sequential nodes
// - Remove unreachable nodes
// - Detect and report cycles
package compiler

import (
	"github.com/aistudio/packages/workflow"
)

// ============================================================================
// Graph Optimizer Types
// ============================================================================

// OptimizationResult contains the result of graph optimization.
type OptimizationResult struct {
	Nodes          []workflow.Node `json:"nodes"`
	Edges          []workflow.Edge `json:"edges"`
	RemovedNodes   []string        `json:"removed_nodes,omitempty"`
	RemovedEdges   []string        `json:"removed_edges,omitempty"`
	FusedNodes     []string        `json:"fused_nodes,omitempty"`
	Warnings       []string        `json:"warnings,omitempty"`
	OriginalCount  int             `json:"original_count"`
	OptimizedCount int             `json:"optimized_count"`
}

// GraphOptimizer performs static analysis and optimization on workflow DAGs.
type GraphOptimizer struct{}

// NewGraphOptimizer creates a new GraphOptimizer.
func NewGraphOptimizer() *GraphOptimizer {
	return &GraphOptimizer{}
}

// Optimize performs all optimization passes on the workflow.
func (o *GraphOptimizer) Optimize(nodes []workflow.Node, edges []workflow.Edge) *OptimizationResult {
	result := &OptimizationResult{
		Nodes:         nodes,
		Edges:         edges,
		OriginalCount: len(nodes),
		OptimizedCount: len(nodes),
	}

	// Pass 1: Remove invalid edges
	result = o.removeInvalidEdges(result)

	// Pass 2: Remove duplicate edges
	result = o.removeDuplicateEdges(result)

	// Pass 3: Remove unreachable nodes
	result = o.removeUnreachableNodes(result)

	// Pass 4: Remove dead nodes (no path to sink)
	result = o.removeDeadNodes(result)

	// Pass 5: Fuse same-type sequential nodes
	result = o.fuseSameTypeNodes(result)

	result.OptimizedCount = len(result.Nodes)
	return result
}

// removeInvalidEdges removes edges that reference non-existent nodes.
func (o *GraphOptimizer) removeInvalidEdges(r *OptimizationResult) *OptimizationResult {
	nodeIDs := make(map[string]bool, len(r.Nodes))
	for _, n := range r.Nodes {
		nodeIDs[n.ID] = true
	}

	validEdges := make([]workflow.Edge, 0, len(r.Edges))
	for _, e := range r.Edges {
		if nodeIDs[e.Source.NodeID] && nodeIDs[e.Target.NodeID] {
			validEdges = append(validEdges, e)
		} else {
			r.RemovedEdges = append(r.RemovedEdges, e.ID)
			r.Warnings = append(r.Warnings, "removed invalid edge "+e.ID+": references non-existent node")
		}
	}
	r.Edges = validEdges
	return r
}

// removeDuplicateEdges removes edges with the same source and target.
func (o *GraphOptimizer) removeDuplicateEdges(r *OptimizationResult) *OptimizationResult {
	seen := make(map[string]bool)
	uniqueEdges := make([]workflow.Edge, 0, len(r.Edges))

	for _, e := range r.Edges {
		key := e.Source.NodeID + "->" + e.Target.NodeID + ":" + e.Source.PortID + "->" + e.Target.PortID
		if !seen[key] {
			seen[key] = true
			uniqueEdges = append(uniqueEdges, e)
		} else {
			r.RemovedEdges = append(r.RemovedEdges, e.ID)
			r.Warnings = append(r.Warnings, "removed duplicate edge "+e.ID)
		}
	}
	r.Edges = uniqueEdges
	return r
}

// removeUnreachableNodes removes nodes that cannot be reached from any source node.
func (o *GraphOptimizer) removeUnreachableNodes(r *OptimizationResult) *OptimizationResult {
	// Build adjacency list
	adj := make(map[string][]string)
	inDegree := make(map[string]int)
	for _, n := range r.Nodes {
		inDegree[n.ID] = 0
	}
	for _, e := range r.Edges {
		adj[e.Source.NodeID] = append(adj[e.Source.NodeID], e.Target.NodeID)
		inDegree[e.Target.NodeID]++
	}

	// Find source nodes (in-degree == 0)
	queue := []string{}
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	// BFS to find all reachable nodes
	reachable := make(map[string]bool)
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if reachable[curr] {
			continue
		}
		reachable[curr] = true
		for _, next := range adj[curr] {
			if !reachable[next] {
				queue = append(queue, next)
			}
		}
	}

	// Keep only reachable nodes
	reachableNodes := make([]workflow.Node, 0, len(r.Nodes))
	for _, n := range r.Nodes {
		if reachable[n.ID] {
			reachableNodes = append(reachableNodes, n)
		} else {
			r.RemovedNodes = append(r.RemovedNodes, n.ID)
			r.Warnings = append(r.Warnings, "removed unreachable node "+n.ID+" ("+n.Name+")")
		}
	}
	r.Nodes = reachableNodes

	// Remove edges connected to removed nodes
	reachableNodeIDs := reachable
	validEdges := make([]workflow.Edge, 0, len(r.Edges))
	for _, e := range r.Edges {
		if reachableNodeIDs[e.Source.NodeID] && reachableNodeIDs[e.Target.NodeID] {
			validEdges = append(validEdges, e)
		}
	}
	r.Edges = validEdges
	return r
}

// removeDeadNodes removes nodes that have no path to any sink node.
func (o *GraphOptimizer) removeDeadNodes(r *OptimizationResult) *OptimizationResult {
	// Build reverse adjacency list
	revAdj := make(map[string][]string)
	outDegree := make(map[string]int)
	for _, n := range r.Nodes {
		outDegree[n.ID] = 0
	}
	for _, e := range r.Edges {
		revAdj[e.Target.NodeID] = append(revAdj[e.Target.NodeID], e.Source.NodeID)
		outDegree[e.Source.NodeID]++
	}

	// Find sink nodes (out-degree == 0)
	queue := []string{}
	for id, deg := range outDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	// BFS backwards from sink nodes
	canReachSink := make(map[string]bool)
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if canReachSink[curr] {
			continue
		}
		canReachSink[curr] = true
		for _, prev := range revAdj[curr] {
			if !canReachSink[prev] {
				queue = append(queue, prev)
			}
		}
	}

	// Keep only nodes that can reach a sink
	liveNodes := make([]workflow.Node, 0, len(r.Nodes))
	for _, n := range r.Nodes {
		if canReachSink[n.ID] {
			liveNodes = append(liveNodes, n)
		} else {
			r.RemovedNodes = append(r.RemovedNodes, n.ID)
			r.Warnings = append(r.Warnings, "removed dead node "+n.ID+" ("+n.Name+"): no path to output")
		}
	}
	r.Nodes = liveNodes

	// Remove edges connected to removed nodes
	liveNodeIDs := canReachSink
	validEdges := make([]workflow.Edge, 0, len(r.Edges))
	for _, e := range r.Edges {
		if liveNodeIDs[e.Source.NodeID] && liveNodeIDs[e.Target.NodeID] {
			validEdges = append(validEdges, e)
		}
	}
	r.Edges = validEdges
	return r
}

// fuseSameTypeNodes fuses sequential nodes of the same type.
func (o *GraphOptimizer) fuseSameTypeNodes(r *OptimizationResult) *OptimizationResult {
	// Build adjacency and in-degree maps
	adj := make(map[string][]string)
	inDegree := make(map[string]int)
	nodeMap := make(map[string]workflow.Node)
	for _, n := range r.Nodes {
		inDegree[n.ID] = 0
		nodeMap[n.ID] = n
	}
	for _, e := range r.Edges {
		adj[e.Source.NodeID] = append(adj[e.Source.NodeID], e.Target.NodeID)
		inDegree[e.Target.NodeID]++
	}

	// Find fusable pairs: single input, single output, same type
	fused := make(map[string]bool)
	fusedNodes := make([]workflow.Node, 0)
	newEdges := make([]workflow.Edge, 0)

	for _, n := range r.Nodes {
		if fused[n.ID] {
			continue
		}

		// Check if node has exactly one input and one output
		inputEdges := []workflow.Edge{}
		outputEdges := []workflow.Edge{}
		for _, e := range r.Edges {
			if e.Target.NodeID == n.ID {
				inputEdges = append(inputEdges, e)
			}
			if e.Source.NodeID == n.ID {
				outputEdges = append(outputEdges, e)
			}
		}

		if len(outputEdges) == 1 && len(inputEdges) == 1 {
			nextID := outputEdges[0].Target.NodeID
			nextNode, exists := nodeMap[nextID]
			if exists && !fused[nextID] && nextNode.Type == n.Type {
				// Check if next node also has single input/output
				nextInputCount := 0
				nextOutputCount := 0
				for _, e := range r.Edges {
					if e.Target.NodeID == nextID {
						nextInputCount++
					}
					if e.Source.NodeID == nextID {
						nextOutputCount++
					}
				}

				if nextInputCount == 1 && nextOutputCount <= 1 {
					// Fuse nodes
					fusedNode := n
					fusedNode.Name = n.Name + " + " + nextNode.Name
					for k, v := range nextNode.Config {
						fusedNode.Config[k] = v
					}
					fusedNodes = append(fusedNodes, fusedNode)
					fused[n.ID] = true
					fused[nextID] = true
					r.FusedNodes = append(r.FusedNodes, n.ID+"+"+nextID)
					r.Warnings = append(r.Warnings, "fused nodes "+n.Name+" + "+nextNode.Name)
					continue
				}
			}
		}

		fusedNodes = append(fusedNodes, n)
	}

	// Rebuild edges for fused nodes
	fusedNodeIDs := make(map[string]bool)
	for _, n := range fusedNodes {
		fusedNodeIDs[n.ID] = true
	}
	for _, e := range r.Edges {
		if fusedNodeIDs[e.Source.NodeID] && fusedNodeIDs[e.Target.NodeID] {
			newEdges = append(newEdges, e)
		}
	}

	r.Nodes = fusedNodes
	r.Edges = newEdges
	return r
}

// DetectCycles checks if the workflow contains cycles.
func (o *GraphOptimizer) DetectCycles(nodes []workflow.Node, edges []workflow.Edge) (hasCycle bool, cycleNodes []string) {
	adj := make(map[string][]string)
	for _, n := range nodes {
		adj[n.ID] = []string{}
	}
	for _, e := range edges {
		adj[e.Source.NodeID] = append(adj[e.Source.NodeID], e.Target.NodeID)
	}

	// DFS-based cycle detection
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var dfs func(node string) bool

	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, next := range adj[node] {
			if !visited[next] {
				if dfs(next) {
					cycleNodes = append(cycleNodes, node)
					return true
				}
			} else if recStack[next] {
				cycleNodes = append(cycleNodes, node, next)
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for _, n := range nodes {
		if !visited[n.ID] {
			if dfs(n.ID) {
				return true, cycleNodes
			}
		}
	}

	return false, nil
}
