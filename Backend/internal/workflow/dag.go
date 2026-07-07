package workflow

import (
	"fmt"
)

func TopologicalSort(nodes []Node, edges []Edge) ([]string, error) {
	inDegree := make(map[string]int)
	adjacency := make(map[string][]string)

	for _, node := range nodes {
		inDegree[node.ID] = 0
	}

	for _, edge := range edges {
		src := edge.Source.NodeID
		tgt := edge.Target.NodeID
		if _, ok := inDegree[src]; !ok {
			return nil, fmt.Errorf("edge references unknown source node: %s", src)
		}
		if _, ok := inDegree[tgt]; !ok {
			return nil, fmt.Errorf("edge references unknown target node: %s", tgt)
		}
		adjacency[src] = append(adjacency[src], tgt)
		inDegree[tgt]++
	}

	queue := make([]string, 0)
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	result := make([]string, 0, len(nodes))
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)
		for _, neighbor := range adjacency[nodeID] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(result) != len(nodes) {
		return nil, fmt.Errorf("workflow contains a cycle: %d of %d nodes are reachable", len(result), len(nodes))
	}

	return result, nil
}

func DetectCycle(nodes []Node, edges []Edge) error {
	_, err := TopologicalSort(nodes, edges)
	return err
}
