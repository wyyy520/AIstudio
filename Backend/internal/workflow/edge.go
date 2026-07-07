package workflow

import (
	"fmt"
)

var typeCompatibility = map[string][]string{
	"image":   {"image", "tensor", "any"},
	"tensor":  {"tensor", "model", "any"},
	"dataset": {"dataset", "any"},
	"model":   {"model", "any"},
	"text":    {"text", "json", "any"},
	"number":  {"number", "json", "any"},
	"boolean": {"boolean", "any"},
	"json":    {"json", "text", "any"},
	"file":    {"file", "any"},
	"stream":  {"stream", "any"},
	"any":     {"any"},
}

func isTypeCompatible(sourceType, targetType string) bool {
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

func ValidateEdge(edge Edge, nodes []Node) error {
	if edge.Source.NodeID == edge.Target.NodeID {
		return fmt.Errorf("edge %s: self-loop is not allowed (node %s)", edge.ID, edge.Source.NodeID)
	}

	var sourceNode, targetNode *Node
	for i := range nodes {
		if nodes[i].ID == edge.Source.NodeID {
			sourceNode = &nodes[i]
		}
		if nodes[i].ID == edge.Target.NodeID {
			targetNode = &nodes[i]
		}
	}

	if sourceNode == nil {
		return fmt.Errorf("edge %s: source node %s not found", edge.ID, edge.Source.NodeID)
	}
	if targetNode == nil {
		return fmt.Errorf("edge %s: target node %s not found", edge.ID, edge.Target.NodeID)
	}

	var sourcePort, targetPort *Port
	for i := range sourceNode.Outputs {
		if sourceNode.Outputs[i].ID == edge.Source.PortID {
			sourcePort = &sourceNode.Outputs[i]
			break
		}
	}
	for i := range targetNode.Inputs {
		if targetNode.Inputs[i].ID == edge.Target.PortID {
			targetPort = &targetNode.Inputs[i]
			break
		}
	}

	if sourcePort == nil {
		return fmt.Errorf("edge %s: source port %s not found on node %s", edge.ID, edge.Source.PortID, edge.Source.NodeID)
	}
	if targetPort == nil {
		return fmt.Errorf("edge %s: target port %s not found on node %s", edge.ID, edge.Target.PortID, edge.Target.NodeID)
	}

	if !isTypeCompatible(sourcePort.Type, targetPort.Type) {
		return fmt.Errorf("edge %s: type mismatch: source port %s (type=%s) -> target port %s (type=%s)",
			edge.ID, sourcePort.ID, sourcePort.Type, targetPort.ID, targetPort.Type)
	}

	return nil
}

func ValidateEdges(nodes []Node, edges []Edge) error {
	inputPortUsed := make(map[string]int)

	for _, edge := range edges {
		if err := ValidateEdge(edge, nodes); err != nil {
			return err
		}

		key := edge.Target.NodeID + ":" + edge.Target.PortID
		inputPortUsed[key]++
	}

	for _, node := range nodes {
		for _, port := range node.Inputs {
			key := node.ID + ":" + port.ID
			count := inputPortUsed[key]
			if port.Required && count == 0 {
				return fmt.Errorf("node %s: required input port %s has no connection", node.ID, port.ID)
			}
			if !port.Multiple && count > 1 {
				return fmt.Errorf("node %s: input port %s does not allow multiple connections (got %d)", node.ID, port.ID, count)
			}
		}
	}

	return nil
}
