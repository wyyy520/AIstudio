// Package compiler provides the EWIR (Engineering Workflow Intermediate Representation) Builder.
//
// EWIR Builder separates workflow.json into:
// - ui.json: Editor state (viewport, selection, layout)
// - workflow.ir.json: Engineering intermediate representation (pure data, no UI)
//
// This separation ensures:
// - Generators only read workflow.ir.json (never ui.json)
// - Editor can restore state from ui.json
// - Clean separation of concerns
package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aistudio/backend/internal/workflow"
)

// ============================================================================
// EWIR Types
// ============================================================================

// EWIR (Engineering Workflow Intermediate Representation) is the core data model.
// It contains only engineering-relevant data, no UI information.
type EWIR struct {
	IRVersion    string        `json:"ir_version"`
	Project      EWIRProject   `json:"project"`
	Nodes        []EWIRNode    `json:"nodes"`
	Edges        []EWIREdge    `json:"edges"`
	DataFlow     []DataFlow    `json:"data_flow,omitempty"`
	ControlFlow  []ControlFlow `json:"control_flow,omitempty"`
	Dependencies []Dependency  `json:"dependencies,omitempty"`
}

// EWIRProject contains project metadata.
type EWIRProject struct {
	Name            string `json:"name"`
	GeneratedAt     string `json:"generated_at"`
	CompilerVersion string `json:"compiler_version"`
	WorkflowID      string `json:"workflow_id"`
}

// EWIRNode represents a node in the intermediate representation.
type EWIRNode struct {
	ID         string         `json:"id"`
	Capability string         `json:"capability"`
	Runtime    string         `json:"runtime"`
	Domain     string         `json:"domain"`
	Params     map[string]any `json:"params"`
	Inputs     []EWIRPort     `json:"inputs"`
	Outputs    []EWIRPort     `json:"outputs"`
	Enabled    bool           `json:"enabled"`
}

// EWIRPort represents a port in the intermediate representation.
type EWIRPort struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	DataType  string `json:"data_type"`
	Direction string `json:"direction"`
}

// EWIREdge represents an edge in the intermediate representation.
type EWIREdge struct {
	ID         string `json:"id"`
	SourceNode string `json:"source_node"`
	TargetNode string `json:"target_node"`
	SourcePort string `json:"source_port"`
	TargetPort string `json:"target_port"`
	EdgeType   string `json:"edge_type"` // "data" or "control"
	Condition  string `json:"condition,omitempty"`
}

// DataFlow represents a data dependency between nodes.
type DataFlow struct {
	FromNode string `json:"from_node"`
	ToNode   string `json:"to_node"`
	DataType string `json:"data_type"`
}

// ControlFlow represents a control dependency between nodes.
type ControlFlow struct {
	FromNode  string `json:"from_node"`
	ToNode    string `json:"to_node"`
	Condition string `json:"condition,omitempty"`
}

// Dependency represents a node dependency.
type Dependency struct {
	NodeID     string   `json:"node_id"`
	DependsOn  []string `json:"depends_on"`
	Dependents []string `json:"dependents"`
}

// UIMetadata contains editor state for restoring the editing session.
type UIMetadata struct {
	Version  string   `json:"version"`
	Viewport Viewport `json:"viewport"`
	Nodes    []UINode `json:"nodes"`
	Edges    []UIEdge `json:"edges"`
	SavedAt  string   `json:"saved_at"`
}

// Viewport contains the editor viewport state.
type Viewport struct {
	Zoom       float64 `json:"zoom"`
	OffsetX    float64 `json:"offset_x"`
	OffsetY    float64 `json:"offset_y"`
	ShowGrid   bool    `json:"show_grid"`
	SnapToGrid bool    `json:"snap_to_grid"`
}

// UINode contains UI-specific node data.
type UINode struct {
	ID        string `json:"id"`
	Position  Point  `json:"position"`
	Size      Size   `json:"size"`
	Selected  bool   `json:"selected"`
	Collapsed bool   `json:"collapsed"`
}

// Point represents a 2D position.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Size represents dimensions.
type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// UIEdge contains UI-specific edge data.
type UIEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Color  string `json:"color,omitempty"`
	Label  string `json:"label,omitempty"`
}

// ============================================================================
// EWIR Builder
// ============================================================================

// EWIRBuilder separates workflow.json into ui.json and workflow.ir.json.
type EWIRBuilder struct {
	compilerVersion string
}

// NewEWIRBuilder creates a new EWIRBuilder.
func NewEWIRBuilder(compilerVersion string) *EWIRBuilder {
	return &EWIRBuilder{
		compilerVersion: compilerVersion,
	}
}

// SplitResult contains the result of splitting workflow.json.
type SplitResult struct {
	UIJSON   *UIMetadata `json:"ui_json"`
	EWIR     *EWIR       `json:"ewir"`
	UIPath   string      `json:"ui_path"`
	EWIRPath string      `json:"ewir_path"`
}

// Split separates a workflow into ui.json and workflow.ir.json.
func (b *EWIRBuilder) Split(wf *workflow.Workflow, outputDir string) (*SplitResult, error) {
	// Build EWIR
	ewir := b.buildEWIR(wf)

	// Build UI metadata
	uiMeta := b.buildUIMetadata(wf)

	// Write files if output directory is provided
	uiPath := ""
	ewirPath := ""
	if outputDir != "" {
		var err error
		uiPath, err = b.writeUIJSON(uiMeta, outputDir)
		if err != nil {
			return nil, fmt.Errorf("failed to write ui.json: %w", err)
		}

		ewirPath, err = b.writeEWIRJSON(ewir, outputDir)
		if err != nil {
			return nil, fmt.Errorf("failed to write workflow.ir.json: %w", err)
		}
	}

	return &SplitResult{
		UIJSON:   uiMeta,
		EWIR:     ewir,
		UIPath:   uiPath,
		EWIRPath: ewirPath,
	}, nil
}

// buildEWIR constructs the Engineering Workflow Intermediate Representation.
func (b *EWIRBuilder) buildEWIR(wf *workflow.Workflow) *EWIR {
	ewir := &EWIR{
		IRVersion: "1.0.0",
		Project: EWIRProject{
			Name:            wf.Name,
			GeneratedAt:     time.Now().Format(time.RFC3339),
			CompilerVersion: b.compilerVersion,
			WorkflowID:      wf.ID,
		},
		Nodes: make([]EWIRNode, 0, len(wf.Nodes)),
		Edges: make([]EWIREdge, 0, len(wf.Edges)),
	}

	// Convert nodes
	for _, node := range wf.Nodes {
		ewirNode := EWIRNode{
			ID:         node.ID,
			Capability: string(node.Type),
			Runtime:    node.Domain,
			Domain:     node.Domain,
			Params:     node.Config,
			Enabled:    true,
		}

		// Convert inputs
		for _, port := range node.Inputs {
			ewirNode.Inputs = append(ewirNode.Inputs, EWIRPort{
				ID:        port.ID,
				Name:      port.Name,
				DataType:  string(port.Type),
				Direction: "input",
			})
		}

		// Convert outputs
		for _, port := range node.Outputs {
			ewirNode.Outputs = append(ewirNode.Outputs, EWIRPort{
				ID:        port.ID,
				Name:      port.Name,
				DataType:  string(port.Type),
				Direction: "output",
			})
		}

		ewir.Nodes = append(ewir.Nodes, ewirNode)
	}

	// Convert edges
	for _, edge := range wf.Edges {
		ewirEdge := EWIREdge{
			ID:         edge.ID,
			SourceNode: edge.Source.NodeID,
			TargetNode: edge.Target.NodeID,
			SourcePort: edge.Source.PortID,
			TargetPort: edge.Target.PortID,
			EdgeType:   "data", // default to data edge
		}

		// Check if edge has condition (control edge)
		if edge.Condition != nil {
			ewirEdge.EdgeType = "control"
			ewirEdge.Condition = edge.Condition.Expression
		}

		ewir.Edges = append(ewir.Edges, ewirEdge)
	}

	// Build data flow
	ewir.DataFlow = b.buildDataFlow(wf)

	// Build control flow
	ewir.ControlFlow = b.buildControlFlow(wf)

	// Build dependencies
	ewir.Dependencies = b.buildDependencies(wf)

	return ewir
}

// buildDataFlow extracts data dependencies from edges.
func (b *EWIRBuilder) buildDataFlow(wf *workflow.Workflow) []DataFlow {
	dataFlow := []DataFlow{}
	for _, edge := range wf.Edges {
		if edge.Condition == nil {
			dataFlow = append(dataFlow, DataFlow{
				FromNode: edge.Source.NodeID,
				ToNode:   edge.Target.NodeID,
				DataType: "data",
			})
		}
	}
	return dataFlow
}

// buildControlFlow extracts control dependencies from edges.
func (b *EWIRBuilder) buildControlFlow(wf *workflow.Workflow) []ControlFlow {
	controlFlow := []ControlFlow{}
	for _, edge := range wf.Edges {
		if edge.Condition != nil {
			controlFlow = append(controlFlow, ControlFlow{
				FromNode:  edge.Source.NodeID,
				ToNode:    edge.Target.NodeID,
				Condition: edge.Condition.Expression,
			})
		}
	}
	return controlFlow
}

// buildDependencies builds the dependency graph for all nodes.
func (b *EWIRBuilder) buildDependencies(wf *workflow.Workflow) []Dependency {
	// Build adjacency maps
	dependsOn := make(map[string][]string)
	dependents := make(map[string][]string)

	for _, edge := range wf.Edges {
		dependsOn[edge.Target.NodeID] = append(dependsOn[edge.Target.NodeID], edge.Source.NodeID)
		dependents[edge.Source.NodeID] = append(dependents[edge.Source.NodeID], edge.Target.NodeID)
	}

	deps := []Dependency{}
	for _, node := range wf.Nodes {
		deps = append(deps, Dependency{
			NodeID:     node.ID,
			DependsOn:  dependsOn[node.ID],
			Dependents: dependents[node.ID],
		})
	}

	return deps
}

// buildUIMetadata constructs the UI metadata for editor state restoration.
func (b *EWIRBuilder) buildUIMetadata(wf *workflow.Workflow) *UIMetadata {
	uiMeta := &UIMetadata{
		Version: "1.0.0",
		Viewport: Viewport{
			Zoom:       1.0,
			OffsetX:    0,
			OffsetY:    0,
			ShowGrid:   true,
			SnapToGrid: true,
		},
		Nodes:   make([]UINode, 0, len(wf.Nodes)),
		Edges:   make([]UIEdge, 0, len(wf.Edges)),
		SavedAt: time.Now().Format(time.RFC3339),
	}

	// Extract viewport from workflow metadata if available
	if wf.Metadata != nil {
		if viewport, ok := wf.Metadata["viewport"].(map[string]any); ok {
			if zoom, ok := viewport["zoom"].(float64); ok {
				uiMeta.Viewport.Zoom = zoom
			}
			if offsetX, ok := viewport["offsetX"].(float64); ok {
				uiMeta.Viewport.OffsetX = offsetX
			}
			if offsetY, ok := viewport["offsetY"].(float64); ok {
				uiMeta.Viewport.OffsetY = offsetY
			}
		}
	}

	// Convert nodes to UI nodes
	for _, node := range wf.Nodes {
		uiNode := UINode{
			ID: node.ID,
			Position: Point{
				X: 0,
				Y: 0,
			},
			Size: Size{
				Width:  250,
				Height: 100,
			},
		}

		// Extract position from node config if available
		if node.Config != nil {
			if pos, ok := node.Config["position"].(map[string]any); ok {
				if x, ok := pos["x"].(float64); ok {
					uiNode.Position.X = x
				}
				if y, ok := pos["y"].(float64); ok {
					uiNode.Position.Y = y
				}
			}
		}

		uiMeta.Nodes = append(uiMeta.Nodes, uiNode)
	}

	// Convert edges to UI edges
	for _, edge := range wf.Edges {
		uiEdge := UIEdge{
			ID:     edge.ID,
			Source: edge.Source.NodeID,
			Target: edge.Target.NodeID,
		}

		if edge.Label != "" {
			uiEdge.Label = edge.Label
		}

		uiMeta.Edges = append(uiMeta.Edges, uiEdge)
	}

	return uiMeta
}

// writeUIJSON writes the UI metadata to ui.json.
func (b *EWIRBuilder) writeUIJSON(meta *UIMetadata, outputDir string) (string, error) {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal ui.json: %w", err)
	}

	path := filepath.Join(outputDir, "ui.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write ui.json: %w", err)
	}

	return path, nil
}

// writeEWIRJSON writes the EWIR to workflow.ir.json.
func (b *EWIRBuilder) writeEWIRJSON(ewir *EWIR, outputDir string) (string, error) {
	data, err := json.MarshalIndent(ewir, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal workflow.ir.json: %w", err)
	}

	path := filepath.Join(outputDir, "workflow.ir.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write workflow.ir.json: %w", err)
	}

	return path, nil
}
