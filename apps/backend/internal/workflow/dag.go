package workflow

import (
	"sort"
)

// ============================================================================
// DAG (Directed Acyclic Graph) Operations
// ============================================================================

// TopologicalSort performs a topological sort of the workflow nodes.
// Returns the nodes in execution order (dependencies first).
func TopologicalSort(wf *Workflow) ([]Node, error) {
	// Build adjacency list and in-degree count
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	for _, node := range wf.Nodes {
		inDegree[node.ID] = 0
	}

	for _, edge := range wf.Edges {
		adjList[edge.Source.NodeID] = append(adjList[edge.Source.NodeID], edge.Target.NodeID)
		inDegree[edge.Target.NodeID]++
	}

	// Kahn's algorithm
	var queue []string
	for _, node := range wf.Nodes {
		if inDegree[node.ID] == 0 {
			queue = append(queue, node.ID)
		}
	}

	// Sort queue for deterministic output
	sort.Strings(queue)

	var sorted []Node
	nodeMap := makeNodeMap(wf)

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]

		if node, ok := nodeMap[nodeID]; ok {
			sorted = append(sorted, node)
		}

		for _, neighbor := range adjList[nodeID] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(sorted) != len(wf.Nodes) {
		return nil, &ValidationError{
			Code:    "CYCLE_DETECTED",
			Message: "graph contains a cycle — topological sort impossible",
		}
	}

	return sorted, nil
}

// GetUpstreamNodes returns all nodes that are upstream of the given node.
func GetUpstreamNodes(wf *Workflow, nodeID string) []Node {
	// Build reverse adjacency list (target → source)
	reverseAdj := make(map[string][]string)
	for _, edge := range wf.Edges {
		reverseAdj[edge.Target.NodeID] = append(reverseAdj[edge.Target.NodeID], edge.Source.NodeID)
	}

	visited := make(map[string]bool)
	var upstream []string

	var dfs func(id string)
	dfs = func(id string) {
		for _, sourceID := range reverseAdj[id] {
			if !visited[sourceID] {
				visited[sourceID] = true
				upstream = append(upstream, sourceID)
				dfs(sourceID)
			}
		}
	}

	dfs(nodeID)

	nodeMap := makeNodeMap(wf)
	var result []Node
	for _, id := range upstream {
		if node, ok := nodeMap[id]; ok {
			result = append(result, node)
		}
	}

	return result
}

// GetDownstreamNodes returns all nodes that are downstream of the given node.
func GetDownstreamNodes(wf *Workflow, nodeID string) []Node {
	adjList := make(map[string][]string)
	for _, edge := range wf.Edges {
		adjList[edge.Source.NodeID] = append(adjList[edge.Source.NodeID], edge.Target.NodeID)
	}

	visited := make(map[string]bool)
	var downstream []string

	var dfs func(id string)
	dfs = func(id string) {
		for _, targetID := range adjList[id] {
			if !visited[targetID] {
				visited[targetID] = true
				downstream = append(downstream, targetID)
				dfs(targetID)
			}
		}
	}

	dfs(nodeID)

	nodeMap := makeNodeMap(wf)
	var result []Node
	for _, id := range downstream {
		if node, ok := nodeMap[id]; ok {
			result = append(result, node)
		}
	}

	return result
}

// GetSourceNodes returns all nodes with no incoming edges (root nodes).
func GetSourceNodes(wf *Workflow) []Node {
	inDegree := make(map[string]int)
	for _, node := range wf.Nodes {
		inDegree[node.ID] = 0
	}
	for _, edge := range wf.Edges {
		inDegree[edge.Target.NodeID]++
	}

	var sources []Node
	for _, node := range wf.Nodes {
		if inDegree[node.ID] == 0 {
			sources = append(sources, node)
		}
	}
	return sources
}

// GetSinkNodes returns all nodes with no outgoing edges (leaf nodes).
func GetSinkNodes(wf *Workflow) []Node {
	outDegree := make(map[string]int)
	for _, node := range wf.Nodes {
		outDegree[node.ID] = 0
	}
	for _, edge := range wf.Edges {
		outDegree[edge.Source.NodeID]++
	}

	var sinks []Node
	for _, node := range wf.Nodes {
		if outDegree[node.ID] == 0 {
			sinks = append(sinks, node)
		}
	}
	return sinks
}

// GetNodeByID returns a node by its ID.
func GetNodeByID(wf *Workflow, nodeID string) (Node, bool) {
	for _, node := range wf.Nodes {
		if node.ID == nodeID {
			return node, true
		}
	}
	return Node{}, false
}

// GetEdgesForNode returns all edges connected to a node.
func GetEdgesForNode(wf *Workflow, nodeID string) []Edge {
	var edges []Edge
	for _, edge := range wf.Edges {
		if edge.Source.NodeID == nodeID || edge.Target.NodeID == nodeID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// GetInputEdges returns all edges targeting a node.
func GetInputEdges(wf *Workflow, nodeID string) []Edge {
	var edges []Edge
	for _, edge := range wf.Edges {
		if edge.Target.NodeID == nodeID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// GetOutputEdges returns all edges originating from a node.
func GetOutputEdges(wf *Workflow, nodeID string) []Edge {
	var edges []Edge
	for _, edge := range wf.Edges {
		if edge.Source.NodeID == nodeID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// makeNodeMap creates a map from node ID to Node.
func makeNodeMap(wf *Workflow) map[string]Node {
	nodeMap := make(map[string]Node, len(wf.Nodes))
	for _, node := range wf.Nodes {
		nodeMap[node.ID] = node
	}
	return nodeMap
}

// IsValidDAG checks if the workflow graph is a valid DAG.
func IsValidDAG(wf *Workflow) bool {
	_, err := TopologicalSort(wf)
	return err == nil
}

// GetNodeDepth returns the depth of a node in the DAG (distance from source).
func GetNodeDepth(wf *Workflow, nodeID string) int {
	upstream := GetUpstreamNodes(wf, nodeID)
	if len(upstream) == 0 {
		return 0
	}

	// Find max depth among upstream nodes
	maxDepth := 0
	for _, u := range upstream {
		d := GetNodeDepth(wf, u.ID) + 1
		if d > maxDepth {
			maxDepth = d
		}
	}
	return maxDepth
}

// GetExecutionLevels returns nodes grouped by execution level.
// Level 0: source nodes (no dependencies)
// Level 1: nodes that depend only on level 0
// Level 2: nodes that depend on level 0 and 1
// etc.
func GetExecutionLevels(wf *Workflow) [][]Node {
	sorted, err := TopologicalSort(wf)
	if err != nil {
		return nil
	}

	level := make(map[string]int)
	for _, node := range sorted {
		inputEdges := GetInputEdges(wf, node.ID)
		if len(inputEdges) == 0 {
			level[node.ID] = 0
		} else {
			maxLevel := 0
			for _, edge := range inputEdges {
				if l, ok := level[edge.Source.NodeID]; ok && l >= maxLevel {
					maxLevel = l + 1
				}
			}
			level[node.ID] = maxLevel
		}
	}

	// Group by level
	maxLevel := 0
	for _, l := range level {
		if l > maxLevel {
			maxLevel = l
		}
	}

	levels := make([][]Node, maxLevel+1)
	for _, node := range wf.Nodes {
		l := level[node.ID]
		levels[l] = append(levels[l], node)
	}

	return levels
}