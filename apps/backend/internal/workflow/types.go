// Package workflow defines the AIStudio Workflow DSL.
//
// Workflow is the Single Source of Truth for the entire platform.
// It contains only declaration data — no runtime state, no execution status.
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
	"os"
	"time"
)

// ============================================================================
// Workflow — The Single Source of Truth
// ============================================================================

// Workflow represents a complete AI engineering workflow.
// It is the single source of truth for the entire platform.
// Workflow contains only declaration data — no runtime state.
type Workflow struct {
	SchemaVersion string            `json:"schema_version" yaml:"schema_version"`
	ID            string            `json:"id" yaml:"id"`
	Name          string            `json:"name" yaml:"name"`
	Description   string            `json:"description,omitempty" yaml:"description,omitempty"`
	Version       int               `json:"version" yaml:"version"`
	Author        string            `json:"author,omitempty" yaml:"author,omitempty"`
	Tags          []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Metadata      map[string]any    `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Variables     map[string]any    `json:"variables,omitempty" yaml:"variables,omitempty"`
	Target        Target            `json:"target" yaml:"target"`
	Nodes         []Node            `json:"nodes" yaml:"nodes"`
	Edges         []Edge            `json:"edges" yaml:"edges"`
	CreatedAt     time.Time         `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt     time.Time         `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

// WorkflowFile wraps a Workflow with its on-disk file path.
// It represents the workflow.json file on the filesystem.
type WorkflowFile struct {
	Workflow *Workflow `json:"workflow"`
	FilePath string    `json:"file_path"`
}

// ============================================================================
// Workflow File I/O
// ============================================================================

// LoadFromFile reads and parses a workflow from a JSON file on disk.
func LoadFromFile(path string) (*Workflow, error) {
	return ParseFile(path)
}

// SaveToFile writes a workflow to a JSON file on disk with atomic write.
// Writes to a temp file first, then renames to prevent corruption.
func SaveToFile(wf *Workflow, path string) error {
	return Save(wf, path)
}

// ============================================================================
// Schema Migration
// ============================================================================

// SchemaMigrator handles workflow schema version migration.
type SchemaMigrator struct {
	migrations map[string]func(wf *Workflow) error
}

// NewSchemaMigrator creates a new SchemaMigrator with built-in migrations.
func NewSchemaMigrator() *SchemaMigrator {
	m := &SchemaMigrator{
		migrations: make(map[string]func(wf *Workflow) error),
	}
	m.registerDefaults()
	return m
}

// Register adds a migration from one version to another.
func (m *SchemaMigrator) Register(fromVersion string, fn func(wf *Workflow) error) {
	m.migrations[fromVersion] = fn
}

// Migrate migrates a workflow from its current version to the target version.
// Returns true if migration was performed, false if already at target.
func (m *SchemaMigrator) Migrate(wf *Workflow, targetVersion string) (bool, error) {
	if wf.SchemaVersion == targetVersion {
		return false, nil
	}

	current := wf.SchemaVersion
	maxIter := 20
	for i := 0; i < maxIter; i++ {
		if current == targetVersion {
			return true, nil
		}
		fn, ok := m.migrations[current]
		if !ok {
			return false, &MigrationError{
				FromVersion: current,
				ToVersion:   targetVersion,
				Message:     "no migration path available",
			}
		}
		if err := fn(wf); err != nil {
			return false, &MigrationError{
				FromVersion: current,
				ToVersion:   targetVersion,
				Message:     err.Error(),
			}
		}
		current = wf.SchemaVersion
	}
	return false, &MigrationError{
		FromVersion: current,
		ToVersion:   targetVersion,
		Message:     "migration exceeded max iterations (possible cycle)",
	}
}

// registerDefaults registers built-in schema migrations.
func (m *SchemaMigrator) registerDefaults() {
	m.Register("1.0.0", func(wf *Workflow) error {
		wf.SchemaVersion = "2.0.0"
		return nil
	})
}

// MigrationError represents a workflow schema migration error.
type MigrationError struct {
	FromVersion string `json:"from_version"`
	ToVersion   string `json:"to_version"`
	Message     string `json:"message"`
}

func (e *MigrationError) Error() string {
	return "workflow migration " + e.FromVersion + " -> " + e.ToVersion + ": " + e.Message
}

// ============================================================================
// Workflow Manager — File-based Workflow CRUD
// ============================================================================

// WorkflowManager handles file-based workflow CRUD operations.
// It is the bridge between the workflow DSL and the filesystem.
type WorkflowManager struct {
	migrator *SchemaMigrator
}

// NewWorkflowManager creates a new WorkflowManager.
func NewWorkflowManager() *WorkflowManager {
	return &WorkflowManager{
		migrator: NewSchemaMigrator(),
	}
}

// Read reads and parses a workflow from a JSON file.
// Automatically migrates the workflow to the latest schema version.
func (m *WorkflowManager) Read(path string) (*Workflow, error) {
	wf, err := LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	migrated, err := m.migrator.Migrate(wf, CurrentSchemaVersion)
	if err != nil {
		return nil, err
	}
	if migrated {
		if err := SaveToFile(wf, path); err != nil {
			return nil, err
		}
	}

	return wf, nil
}

// Write saves a workflow to a JSON file atomically.
func (m *WorkflowManager) Write(wf *Workflow, path string) error {
	return SaveToFile(wf, path)
}

// Exists checks if a workflow file exists at the given path.
func (m *WorkflowManager) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Delete removes a workflow file from disk.
func (m *WorkflowManager) Delete(path string) error {
	return os.Remove(path)
}

// CreateDefault creates a new default workflow with the given parameters.
func (m *WorkflowManager) CreateDefault(projectID, name, target string) *Workflow {
	now := time.Now()
	return &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            projectID,
		Name:          name,
		Version:       1,
		Target:        Target(target),
		Nodes:         make([]Node, 0),
		Edges:         make([]Edge, 0),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// ============================================================================
// Engine — Workflow Execution Engine
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
			result := &ExecutionResult{
				Status:        "failed",
				StartedAt:     startTime,
				Duration:      nodeEndTime.Sub(startTime),
				NodeResults:   results,
				FailedNodeID:  node.ID,
				Error:         fmt.Sprintf("node %s (%s) failed: %s", node.Name, node.ID, err),
			}
			return result, nil
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
	Status       string                       `json:"status"`
	StartedAt    time.Time                    `json:"started_at"`
	Duration     time.Duration                `json:"duration"`
	NodeResults  map[string]map[string]interface{} `json:"node_results,omitempty"`
	FailedNodeID string                       `json:"failed_node_id,omitempty"`
	Error        string                       `json:"error,omitempty"`
}

// ============================================================================
// Node — A single step in the workflow
// ============================================================================

// Node represents a single step in the workflow DAG.
// It is a pure declaration — no runtime state.
type Node struct {
	ID          string         `json:"id" yaml:"id"`
	Type        NodeType       `json:"type" yaml:"type"`
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	Position    Point          `json:"position" yaml:"position"`
	Size        *Size          `json:"size,omitempty" yaml:"size,omitempty"`
	Config      map[string]any `json:"config,omitempty" yaml:"config,omitempty"`
	Inputs      []Port         `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs     []Port         `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Constraints *Constraints   `json:"constraints,omitempty" yaml:"constraints,omitempty"`
}

// NodeType identifies the type of workflow node.
type NodeType string

// Built-in DSL native node types (not plugin-backed).
const (
	// Logic control structures — built into the Workflow DSL, not plugins
	NodeTypeCondition NodeType = "control.condition"
	NodeTypeLoop      NodeType = "control.loop"
	NodeTypeSwitch    NodeType = "control.switch"
	NodeTypeRetry     NodeType = "control.retry"
)

// Algorithm/processing node types (may be provided by plugins).
const (
	NodeTypeDataLoader     NodeType = "data_loader"
	NodeTypeDataPreprocess NodeType = "data_preprocessor"
	NodeTypeDataAugment    NodeType = "data_augmentation"
	NodeTypeModelTrainer   NodeType = "model_trainer"
	NodeTypeModelEvaluator NodeType = "model_evaluator"
	NodeTypeModelExporter   NodeType = "model_exporter"
	NodeTypeModelInference  NodeType = "model_inference"
	NodeTypeDataSplit       NodeType = "data_split"
	NodeTypeFeatureExtract  NodeType = "feature_extractor"
	NodeTypeHyperparamTune  NodeType = "hyperparameter_tuning"
	NodeTypeVisualization   NodeType = "visualization"
	NodeTypeMetricCompute   NodeType = "metric_computation"
	NodeTypeCustom          NodeType = "custom"
)

// ============================================================================
// Edge — Connection between two nodes
// ============================================================================

// Edge represents a connection between two nodes in the workflow DAG.
// It defines the data flow from a source node's output port to a target node's input port.
type Edge struct {
	ID     string       `json:"id" yaml:"id"`
	Source EdgeEndpoint `json:"source" yaml:"source"`
	Target EdgeEndpoint `json:"target" yaml:"target"`
}

// EdgeEndpoint identifies a specific port on a node.
type EdgeEndpoint struct {
	NodeID string `json:"node_id" yaml:"node_id"`
	PortID string `json:"port_id" yaml:"port_id"`
}

// ============================================================================
// Port — Input/Output port of a node
// ============================================================================

// Port defines a typed input or output of a node.
type Port struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Type        DataType `json:"type" yaml:"type"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool     `json:"required,omitempty" yaml:"required,omitempty"`
}

// DataType represents the type of data flowing through a port.
type DataType string

const (
	DataTypeImage    DataType = "image"
	DataTypeTensor   DataType = "tensor"
	DataTypeDataset  DataType = "dataset"
	DataTypeModel    DataType = "model"
	DataTypeText     DataType = "text"
	DataTypeNumber   DataType = "number"
	DataTypeBoolean  DataType = "boolean"
	DataTypeJSON     DataType = "json"
	DataTypeFile     DataType = "file"
	DataTypeStream   DataType = "stream"
	DataTypeAny      DataType = "any"
	DataTypeResult   DataType = "result"
)

// ============================================================================
// Target — Target platform for compilation
// ============================================================================

// Target identifies the target platform for compilation.
type Target string

const (
	TargetPython Target = "python"
	TargetMATLAB Target = "matlab"
	TargetROS2   Target = "ros2"
	TargetSTM32  Target = "stm32"
	TargetDocker Target = "docker"
	TargetCPP    Target = "cpp"
	TargetUnity  Target = "unity"
	TargetJava   Target = "java"
)

// ============================================================================
// Utility Types
// ============================================================================

// Point represents a 2D position on the canvas.
type Point struct {
	X float64 `json:"x" yaml:"x"`
	Y float64 `json:"y" yaml:"y"`
}

// Size represents a 2D size.
type Size struct {
	Width  float64 `json:"width" yaml:"width"`
	Height float64 `json:"height" yaml:"height"`
}

// Constraints defines constraints for a node.
type Constraints struct {
	MinInputs      int      `json:"min_inputs,omitempty" yaml:"min_inputs,omitempty"`
	MaxInputs      int      `json:"max_inputs,omitempty" yaml:"max_inputs,omitempty"`
	MinOutputs     int      `json:"min_outputs,omitempty" yaml:"min_outputs,omitempty"`
	MaxOutputs     int      `json:"max_outputs,omitempty" yaml:"max_outputs,omitempty"`
	RequiredConfig []string `json:"required_config,omitempty" yaml:"required_config,omitempty"`
	AllowedTypes   []string `json:"allowed_types,omitempty" yaml:"allowed_types,omitempty"`
}

// ============================================================================
// Schema Version
// ============================================================================

const CurrentSchemaVersion = "2.0.0"

// ValidTargets returns all valid target platforms.
func ValidTargets() []Target {
	return []Target{
		TargetPython,
		TargetMATLAB,
		TargetROS2,
		TargetSTM32,
		TargetDocker,
		TargetCPP,
		TargetUnity,
		TargetJava,
	}
}

// ValidNodeTypes returns all valid node types including DSL-native control nodes.
func ValidNodeTypes() []NodeType {
	return []NodeType{
		// DSL native control nodes
		NodeTypeCondition,
		NodeTypeLoop,
		NodeTypeSwitch,
		NodeTypeRetry,
		// Algorithm/processing nodes
		NodeTypeDataLoader,
		NodeTypeDataPreprocess,
		NodeTypeDataAugment,
		NodeTypeModelTrainer,
		NodeTypeModelEvaluator,
		NodeTypeModelExporter,
		NodeTypeModelInference,
		NodeTypeDataSplit,
		NodeTypeFeatureExtract,
		NodeTypeHyperparamTune,
		NodeTypeVisualization,
		NodeTypeMetricCompute,
		NodeTypeCustom,
	}
}

// ============================================================================
// DSL Native Control Node Configuration Types
// ============================================================================

// ConditionConfig defines the configuration for a control.condition node.
type ConditionConfig struct {
	Expression string `json:"expression" yaml:"expression"`
	TrueBranch string  `json:"true_branch,omitempty" yaml:"true_branch,omitempty"`
	FalseBranch string `json:"false_branch,omitempty" yaml:"false_branch,omitempty"`
}

// LoopConfig defines the configuration for a control.loop node.
type LoopConfig struct {
	Iterations int    `json:"iterations" yaml:"iterations"`
	IteratorVar string `json:"iterator_var,omitempty" yaml:"iterator_var,omitempty"`
	BreakExpr  string `json:"break_expression,omitempty" yaml:"break_expression,omitempty"`
}

// SwitchConfig defines the configuration for a control.switch node.
type SwitchConfig struct {
	Expression string              `json:"expression" yaml:"expression"`
	Cases      []SwitchCase        `json:"cases" yaml:"cases"`
	DefaultCase string             `json:"default_case,omitempty" yaml:"default_case,omitempty"`
}

// SwitchCase defines a single case in a switch node.
type SwitchCase struct {
	Value   any    `json:"value" yaml:"value"`
	BranchID string `json:"branch_id" yaml:"branch_id"`
}

// RetryConfig defines the configuration for a control.retry node.
type RetryConfig struct {
	MaxRetries  int `json:"max_retries" yaml:"max_retries"`
	BackoffMS   int `json:"backoff_ms,omitempty" yaml:"backoff_ms,omitempty"`
	RetryOnAny  bool `json:"retry_on_any,omitempty" yaml:"retry_on_any,omitempty"`
}

// ValidDataTypes returns all valid data types.
func ValidDataTypes() []DataType {
	return []DataType{
		DataTypeImage,
		DataTypeTensor,
		DataTypeDataset,
		DataTypeModel,
		DataTypeText,
		DataTypeNumber,
		DataTypeBoolean,
		DataTypeJSON,
		DataTypeFile,
		DataTypeStream,
		DataTypeAny,
	}
}

// ============================================================================
// MCP (Model Context Protocol) Types
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