// Package workflow defines the AIStudio Workflow DSL.
//
// Workflow is the Single Source of Truth for the entire platform.
// It contains only declaration data — no runtime state, no execution status.
//
// This package re-exports all type definitions from packages/workflow
// while adding internal-only types for the backend engine (Engine, NodeRegistry, etc.).
//
// Design Principles:
// 1. Pure Declaration — Workflow describes WHAT, not HOW
// 2. No Runtime State — No status, progress, logs, errors in workflow.json
// 3. Self-Describing — Contains all info needed for compilation
// 4. Versioned Schema — Schema version for backward compatibility
// 5. Target-Aware — Declares target platform for compilation
// 6. File-Backed — Workflow is always stored as workflow.json on disk
// 7. Atomic Writes — File writes are atomic to prevent corruption
// 8. Manually Editable — Users can edit workflow.json with any text editor
package workflow

import (
	"context"
	"fmt"
	"time"

	pw "github.com/aistudio/packages/workflow"
)

// ============================================================================
// Type Aliases — all types re-exported from packages/workflow
// ============================================================================

type Workflow = pw.Workflow
type WorkflowFile = pw.WorkflowFile
type Node = pw.Node
type Edge = pw.Edge
type Port = pw.Port
type Viewport = pw.Viewport
type PluginInfo = pw.PluginInfo
type EdgeCondition = pw.EdgeCondition
type NodeType = pw.NodeType
type EdgeEndpoint = pw.EdgeEndpoint
type DataType = pw.DataType
type Target = pw.Target
type Point = pw.Point
type Size = pw.Size
type Constraints = pw.Constraints
type ConditionConfig = pw.ConditionConfig
type LoopConfig = pw.LoopConfig
type SwitchConfig = pw.SwitchConfig
type SwitchCase = pw.SwitchCase
type RetryConfig = pw.RetryConfig
type ValidationError = pw.ValidationError
type ValidationResult = pw.ValidationResult
type SchemaMigrator = pw.SchemaMigrator
type MigrationError = pw.MigrationError
type WorkflowManager = pw.WorkflowManager

// ============================================================================
// Constant Aliases
// ============================================================================

const (
	NodeTypeCondition NodeType = pw.NodeTypeCondition
	NodeTypeLoop      NodeType = pw.NodeTypeLoop
	NodeTypeSwitch    NodeType = pw.NodeTypeSwitch
	NodeTypeRetry     NodeType = pw.NodeTypeRetry
)

const (
	NodeTypeDataLoader     NodeType = pw.NodeTypeDataLoader
	NodeTypeDataPreprocess NodeType = pw.NodeTypeDataPreprocess
	NodeTypeDataAugment    NodeType = pw.NodeTypeDataAugment
	NodeTypeModelTrainer   NodeType = pw.NodeTypeModelTrainer
	NodeTypeModelEvaluator NodeType = pw.NodeTypeModelEvaluator
	NodeTypeModelExporter  NodeType = pw.NodeTypeModelExporter
	NodeTypeModelInference NodeType = pw.NodeTypeModelInference
	NodeTypeDataSplit      NodeType = pw.NodeTypeDataSplit
	NodeTypeFeatureExtract NodeType = pw.NodeTypeFeatureExtract
	NodeTypeHyperparamTune NodeType = pw.NodeTypeHyperparamTune
	NodeTypeVisualization  NodeType = pw.NodeTypeVisualization
	NodeTypeMetricCompute  NodeType = pw.NodeTypeMetricCompute
	NodeTypeCustom         NodeType = pw.NodeTypeCustom
)

const (
	DataTypeImage    DataType = pw.DataTypeImage
	DataTypeTensor   DataType = pw.DataTypeTensor
	DataTypeDataset  DataType = pw.DataTypeDataset
	DataTypeModel    DataType = pw.DataTypeModel
	DataTypeText     DataType = pw.DataTypeText
	DataTypeNumber   DataType = pw.DataTypeNumber
	DataTypeBoolean  DataType = pw.DataTypeBoolean
	DataTypeJSON     DataType = pw.DataTypeJSON
	DataTypeFile     DataType = pw.DataTypeFile
	DataTypeStream   DataType = pw.DataTypeStream
	DataTypeAny      DataType = pw.DataTypeAny
)

const (
	TargetPython Target = pw.TargetPython
	TargetMATLAB Target = pw.TargetMATLAB
	TargetROS2   Target = pw.TargetROS2
	TargetSTM32  Target = pw.TargetSTM32
	TargetDocker Target = pw.TargetDocker
	TargetCPP    Target = pw.TargetCPP
	TargetUnity  Target = pw.TargetUnity
	TargetJava   Target = pw.TargetJava
)

const CurrentSchemaVersion = pw.CurrentSchemaVersion

// ============================================================================
// Function Variable Aliases — delegate to packages/workflow
// ============================================================================

// ValidTargets returns all valid target platforms.
var ValidTargets = pw.ValidTargets

// ValidTargetsMap returns a map of valid target platforms for O(1) lookup.
var ValidTargetsMap = pw.ValidTargetsMap

// ValidNodeTypes returns all valid node types including DSL-native control nodes.
var ValidNodeTypes = pw.ValidNodeTypes

// ValidNodeTypesMap returns a map of valid node types for O(1) lookup.
var ValidNodeTypesMap = pw.ValidNodeTypesMap

// ValidDataTypes returns all valid data types.
var ValidDataTypes = pw.ValidDataTypes

// ValidateWorkflow performs comprehensive validation of a workflow.
var ValidateWorkflow = pw.ValidateWorkflow

// Validate is an alias for ValidateWorkflow for backward compatibility.
var Validate = pw.ValidateWorkflow

// TopologicalSort performs a topological sort of the workflow nodes.
var TopologicalSort = pw.TopologicalSort

// GetUpstreamNodes returns all nodes upstream of the given node.
var GetUpstreamNodes = pw.GetUpstreamNodes

// GetDownstreamNodes returns all nodes downstream of the given node.
var GetDownstreamNodes = pw.GetDownstreamNodes

// GetSourceNodes returns all nodes with no incoming edges (root nodes).
var GetSourceNodes = pw.GetSourceNodes

// GetSinkNodes returns all nodes with no outgoing edges (leaf nodes).
var GetSinkNodes = pw.GetSinkNodes

// GetNodeByID returns a node by its ID.
var GetNodeByID = pw.GetNodeByID

// GetEdgesForNode returns all edges connected to a node.
var GetEdgesForNode = pw.GetEdgesForNode

// GetInputEdges returns all edges targeting a node.
var GetInputEdges = pw.GetInputEdges

// GetOutputEdges returns all edges originating from a node.
var GetOutputEdges = pw.GetOutputEdges

// IsValidDAG checks if the workflow graph is a valid DAG.
var IsValidDAG = pw.IsValidDAG

// GetNodeDepth returns the depth of a node in the DAG.
var GetNodeDepth = pw.GetNodeDepth

// GetExecutionLevels returns nodes grouped by execution level.
var GetExecutionLevels = pw.GetExecutionLevels

// Parse parses a workflow from JSON bytes.
var Parse = pw.Parse

// ParseFile parses a workflow from a JSON file.
var ParseFile = pw.ParseFile

// MustParse parses a workflow from JSON bytes, panicking on error.
var MustParse = pw.MustParse

// MustParseFile parses a workflow from a JSON file, panicking on error.
var MustParseFile = pw.MustParseFile

// Save writes a workflow to a JSON file using atomic write.
var Save = pw.Save

// SaveDirect writes a workflow to a JSON file directly.
var SaveDirect = pw.SaveDirect

// ToJSON marshals a workflow to indented JSON.
var ToJSON = pw.ToJSON

// Clone creates a deep copy of a workflow.
var Clone = pw.Clone

// LoadFromFile reads and parses a workflow from a JSON file on disk.
var LoadFromFile = pw.LoadFromFile

// SaveToFile writes a workflow to a JSON file on disk with atomic write.
var SaveToFile = pw.SaveToFile

// ValidateWorkflowFile reads a workflow file and validates its contents.
var ValidateWorkflowFile = pw.ValidateWorkflowFile

// NewSchemaMigrator creates a new SchemaMigrator with built-in migrations.
var NewSchemaMigrator = pw.NewSchemaMigrator

// NewWorkflowManager creates a new WorkflowManager.
var NewWorkflowManager = pw.NewWorkflowManager

// ============================================================================
// Engine — Workflow Execution Engine
// (Internal-only: not in packages/workflow)
// ============================================================================

// Engine executes workflows by resolving nodes, managing execution order,
// and delegating to registered node factories.
type Engine struct {
	registry *NodeRegistry
}

// NewEngine creates a new workflow Engine.
func NewEngine() *Engine {
	return &Engine{
		registry: NewNodeRegistry(),
	}
}

// Registry returns the node registry for this engine.
func (e *Engine) Registry() *NodeRegistry {
	return e.registry
}

// Run executes a workflow from its raw JSON definition bytes.
func (e *Engine) Run(ctx context.Context, data []byte) (*ExecutionResult, error) {
	wf, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("engine run parse error: %w", err)
	}
	return e.RunWorkflow(ctx, wf)
}

// RunWorkflow executes a parsed workflow.
func (e *Engine) RunWorkflow(ctx context.Context, wf *Workflow) (*ExecutionResult, error) {
	// Get topological order
	sorted, err := TopologicalSort(wf)
	if err != nil {
		return nil, fmt.Errorf("engine topological sort error: %w", err)
	}

	// Build node map
	nodeMap := make(map[string]Node)
	for _, n := range wf.Nodes {
		nodeMap[n.ID] = n
	}

	// Execute nodes in order
	results := make(map[string]map[string]interface{})
	startTime := time.Now()

	for _, node := range sorted {
		select {
		case <-ctx.Done():
			return &ExecutionResult{
				Status:    "cancelled",
				StartedAt: startTime,
				Duration:  time.Since(startTime),
				Error:     ctx.Err().Error(),
			}, nil
		default:
		}

		// Gather inputs from upstream nodes
		inputs := make(map[string]interface{})
		inputEdges := GetInputEdges(wf, node.ID)
		for _, edge := range inputEdges {
			if output, ok := results[edge.Source.NodeID]; ok {
				inputs[edge.Target.PortID] = output
			}
		}

		// Find and execute the node
		factory := e.registry.Get(node.Type)
		if factory == nil {
			return nil, fmt.Errorf("engine: no factory registered for node type %s", node.Type)
		}

		execNode := factory()
		output, err := execNode.Execute(ctx, inputs, node.Config)
		if err != nil {
			nodeEndTime := time.Now()
			return &ExecutionResult{
				Status:       "failed",
				StartedAt:    startTime,
				Duration:     nodeEndTime.Sub(startTime),
				NodeResults:  results,
				FailedNodeID: node.ID,
				Error:        fmt.Sprintf("node %s (%s) failed: %s", node.Name, node.ID, err),
			}, nil
		}

		results[node.ID] = output
	}

	endTime := time.Now()
	return &ExecutionResult{
		Status:      "completed",
		StartedAt:   startTime,
		Duration:    endTime.Sub(startTime),
		NodeResults: results,
	}, nil
}

// NodeRegistry manages registered node types and their factories.
type NodeRegistry struct {
	factories map[NodeType]NodeFactory
	defs      map[NodeType]NodeDefinition
}

// NewNodeRegistry creates a new NodeRegistry.
func NewNodeRegistry() *NodeRegistry {
	return &NodeRegistry{
		factories: make(map[NodeType]NodeFactory),
		defs:      make(map[NodeType]NodeDefinition),
	}
}

// Register registers a node type with its factory and definition.
func (r *NodeRegistry) Register(def NodeDefinition) {
	r.factories[def.Type] = def.Factory
	r.defs[def.Type] = def
}

// Get returns the factory for a node type.
func (r *NodeRegistry) Get(nodeType NodeType) NodeFactory {
	return r.factories[nodeType]
}

// GetDefinition returns the definition for a node type.
func (r *NodeRegistry) GetDefinition(nodeType NodeType) (NodeDefinition, bool) {
	def, ok := r.defs[nodeType]
	return def, ok
}

// List returns all registered node definitions.
func (r *NodeRegistry) List() []NodeDefinition {
	defs := make([]NodeDefinition, 0, len(r.defs))
	for _, def := range r.defs {
		defs = append(defs, def)
	}
	return defs
}

// NodeDefinition describes a registered node type.
type NodeDefinition struct {
	Type        NodeType   `json:"type"`
	Plugin      string     `json:"plugin,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Category    string     `json:"category,omitempty"`
	Inputs      []Port     `json:"inputs,omitempty"`
	Outputs     []Port     `json:"outputs,omitempty"`
	Factory     NodeFactory `json:"-"`
}

// NodeFactory creates a new executable node instance.
type NodeFactory func() ExecutableNode

// ExecutableNode is a node that can be executed.
type ExecutableNode interface {
	Execute(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error)
}

// ExecutionResult contains the result of a workflow execution.
type ExecutionResult struct {
	Status       string                           `json:"status"`
	StartedAt    time.Time                        `json:"started_at"`
	Duration     time.Duration                    `json:"duration"`
	NodeResults  map[string]map[string]interface{} `json:"node_results,omitempty"`
	FailedNodeID string                           `json:"failed_node_id,omitempty"`
	Error        string                           `json:"error,omitempty"`
}

// ============================================================================
// MCP (Model Context Protocol) Types
// (Internal-only: not in packages/workflow)
// ============================================================================

// MCPTool describes a tool available on an MCP server.
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	ReturnType  string                 `json:"return_type,omitempty"`
}

// MCPResult contains the result of an MCP tool call.
type MCPResult struct {
	Success bool        `json:"success"`
	Output  interface{} `json:"output,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// MCPRuntime is the interface for MCP runtime operations.
type MCPRuntime interface {
	Connect(ctx context.Context, server string) error
	Disconnect(ctx context.Context, server string) error
	ListTools(ctx context.Context, server string) ([]MCPTool, error)
	CallTool(ctx context.Context, server string, toolName string, params map[string]interface{}) (*MCPResult, error)
	ExecuteWorkflow(ctx context.Context, server string, workflowJSON []byte) (*MCPResult, error)
}
