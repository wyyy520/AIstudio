package workflow

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

func init() {
	initCompat()
}

var (
	onceValidNodeTypes sync.Once
	cachedValidNodeTypes []NodeType

	onceValidTargets sync.Once
	cachedValidTargets []Target

	onceValidDataTypes sync.Once
	cachedValidDataTypes []DataType

	validateNodePool = sync.Pool{
		New: func() any {
			return &ValidationError{}
		},
	}
)

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
type WorkflowFile struct {
	Workflow *Workflow `json:"workflow"`
	FilePath string    `json:"file_path"`
}

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

// Control flow node types — built into the Workflow DSL.
const (
	NodeTypeCondition NodeType = "control.condition"
	NodeTypeLoop      NodeType = "control.loop"
	NodeTypeSwitch    NodeType = "control.switch"
	NodeTypeRetry     NodeType = "control.retry"
)

// Algorithm/processing node types.
const (
	NodeTypeDataLoader     NodeType = "data_loader"
	NodeTypeDataPreprocess NodeType = "data_preprocessor"
	NodeTypeDataAugment    NodeType = "data_augmentation"
	NodeTypeModelTrainer   NodeType = "model_trainer"
	NodeTypeModelEvaluator NodeType = "model_evaluator"
	NodeTypeModelExporter  NodeType = "model_exporter"
	NodeTypeModelInference NodeType = "model_inference"
	NodeTypeDataSplit      NodeType = "data_split"
	NodeTypeFeatureExtract NodeType = "feature_extractor"
	NodeTypeHyperparamTune NodeType = "hyperparameter_tuning"
	NodeTypeVisualization  NodeType = "visualization"
	NodeTypeMetricCompute  NodeType = "metric_computation"
	NodeTypeCustom         NodeType = "custom"
)

// Edge represents a connection between two nodes in the workflow DAG.
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
)

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

// ConditionConfig defines the configuration for a control.condition node.
type ConditionConfig struct {
	Expression   string `json:"expression" yaml:"expression"`
	TrueBranch   string `json:"true_branch,omitempty" yaml:"true_branch,omitempty"`
	FalseBranch  string `json:"false_branch,omitempty" yaml:"false_branch,omitempty"`
}

// LoopConfig defines the configuration for a control.loop node.
type LoopConfig struct {
	Iterations  int    `json:"iterations" yaml:"iterations"`
	IteratorVar string `json:"iterator_var,omitempty" yaml:"iterator_var,omitempty"`
	BreakExpr   string `json:"break_expression,omitempty" yaml:"break_expression,omitempty"`
}

// SwitchConfig defines the configuration for a control.switch node.
type SwitchConfig struct {
	Expression  string       `json:"expression" yaml:"expression"`
	Cases       []SwitchCase `json:"cases" yaml:"cases"`
	DefaultCase string       `json:"default_case,omitempty" yaml:"default_case,omitempty"`
}

// SwitchCase defines a single case in a switch node.
type SwitchCase struct {
	Value    any    `json:"value" yaml:"value"`
	BranchID string `json:"branch_id" yaml:"branch_id"`
}

// RetryConfig defines the configuration for a control.retry node.
type RetryConfig struct {
	MaxRetries int  `json:"max_retries" yaml:"max_retries"`
	BackoffMS  int  `json:"backoff_ms,omitempty" yaml:"backoff_ms,omitempty"`
	RetryOnAny bool `json:"retry_on_any,omitempty" yaml:"retry_on_any,omitempty"`
}

// ============================================================================
// Validation
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
	Valid    bool               `json:"valid"`
	Errors   []*ValidationError `json:"errors,omitempty"`
	Warnings []string           `json:"warnings,omitempty"`
}

// ValidateWorkflow performs comprehensive validation of a workflow.
func ValidateWorkflow(wf *Workflow) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]*ValidationError, 0),
		Warnings: make([]string, 0),
	}

	if err := validateDAG(wf); err != nil {
		result.Errors = append(result.Errors, err)
	}

	validateEdges(wf, result)
	validateDataTypes(wf, result)
	validateRequiredInputs(wf, result)
	validateNodeTypes(wf, result)
	validateVariables(wf, result)

	for _, node := range wf.Nodes {
		if err := ValidateNode(node); err != nil {
			result.Errors = append(result.Errors, &ValidationError{
				Code:    "INVALID_NODE",
				Message: err.Error(),
				NodeID:  node.ID,
			})
		}
	}

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}

// ValidateNode validates an individual workflow node.
func ValidateNode(node Node) error {
	if node.ID == "" {
		return fmt.Errorf("node ID is required")
	}
	if node.Type == "" {
		return fmt.Errorf("node type is required for node %q", node.ID)
	}
	if node.Name == "" {
		return fmt.Errorf("node name is required for node %q", node.ID)
	}

	validTypes := ValidNodeTypesMap()
	if !validTypes[node.Type] {
		return fmt.Errorf("node %q (%s): invalid node type %q", node.ID, node.Name, node.Type)
	}

	switch node.Type {
	case NodeTypeCondition:
		if node.Config == nil {
			return fmt.Errorf("node %q (%s): control.condition requires 'expression' in config", node.ID, node.Name)
		}
		if _, ok := node.Config["expression"]; !ok {
			return fmt.Errorf("node %q (%s): control.condition requires 'expression' in config", node.ID, node.Name)
		}
	case NodeTypeLoop:
		if node.Config == nil {
			return fmt.Errorf("node %q (%s): control.loop requires 'iterations' in config", node.ID, node.Name)
		}
		if _, ok := node.Config["iterations"]; !ok {
			return fmt.Errorf("node %q (%s): control.loop requires 'iterations' in config", node.ID, node.Name)
		}
	case NodeTypeSwitch:
		if node.Config == nil {
			return fmt.Errorf("node %q (%s): control.switch requires 'cases' in config", node.ID, node.Name)
		}
		if _, ok := node.Config["cases"]; !ok {
			return fmt.Errorf("node %q (%s): control.switch requires 'cases' in config", node.ID, node.Name)
		}
	case NodeTypeRetry:
		if node.Config == nil {
			return fmt.Errorf("node %q (%s): control.retry requires 'max_retries' in config", node.ID, node.Name)
		}
		if _, ok := node.Config["max_retries"]; !ok {
			return fmt.Errorf("node %q (%s): control.retry requires 'max_retries' in config", node.ID, node.Name)
		}
	case NodeTypeDataLoader:
		if len(node.Outputs) == 0 {
			return fmt.Errorf("node %q (%s): data_loader must have at least one output port", node.ID, node.Name)
		}
	case NodeTypeModelTrainer:
		if len(node.Inputs) == 0 {
			return fmt.Errorf("node %q (%s): model_trainer must have at least one input port", node.ID, node.Name)
		}
		if len(node.Outputs) == 0 {
			return fmt.Errorf("node %q (%s): model_trainer must have at least one output port", node.ID, node.Name)
		}
	case NodeTypeModelInference:
		if len(node.Inputs) == 0 {
			return fmt.Errorf("node %q (%s): model_inference must have at least one input port", node.ID, node.Name)
		}
	}

	return nil
}

// validateSchema validates the workflow schema and required fields.
func validateSchema(wf *Workflow) error {
	if wf.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}
	if wf.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	if wf.SchemaVersion == "" {
		wf.SchemaVersion = CurrentSchemaVersion
	}
	if wf.Target == "" {
		return fmt.Errorf("workflow target is required")
	}

	validTargets := ValidTargetsMap()
	if !validTargets[wf.Target] {
		return fmt.Errorf("invalid target: %s", wf.Target)
	}

	if len(wf.Nodes) == 0 {
		return nil
	}

	nodeIDs := make(map[string]bool)
	for _, node := range wf.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node ID is required")
		}
		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true
	}

	for _, edge := range wf.Edges {
		if edge.ID == "" {
			return fmt.Errorf("edge ID is required")
		}
		if !nodeIDs[edge.Source.NodeID] {
			return fmt.Errorf("edge %s: source node %s not found", edge.ID, edge.Source.NodeID)
		}
		if !nodeIDs[edge.Target.NodeID] {
			return fmt.Errorf("edge %s: target node %s not found", edge.ID, edge.Target.NodeID)
		}
	}

	return nil
}

// validateDAG checks for cycles in the workflow graph.
func validateDAG(wf *Workflow) *ValidationError {
	graph := make(map[string][]string, len(wf.Nodes))
	for _, edge := range wf.Edges {
		graph[edge.Source.NodeID] = append(graph[edge.Source.NodeID], edge.Target.NodeID)
	}

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

	visited := make(map[string]bool, len(wf.Nodes))
	recStack := make(map[string]bool, len(wf.Nodes))

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
	nodePorts := make(map[string]map[string]Port, len(wf.Nodes))
	for _, node := range wf.Nodes {
		totalPorts := len(node.Inputs) + len(node.Outputs)
		ports := make(map[string]Port, totalPorts)
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
	connectedInputs := make(map[string]bool)
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

// validateEdges checks edges for duplicates and structural issues.
func validateEdges(wf *Workflow, result *ValidationResult) {
	edgeSet := make(map[string]bool)
	for _, edge := range wf.Edges {
		key := edge.Source.NodeID + ":" + edge.Source.PortID + "->" + edge.Target.NodeID + ":" + edge.Target.PortID
		if edgeSet[key] {
			result.Errors = append(result.Errors, &ValidationError{
				Code:    "DUPLICATE_EDGE",
				Message: fmt.Sprintf("duplicate edge: %s", key),
				EdgeID:  edge.ID,
			})
		}
		edgeSet[key] = true

		if edge.Source.NodeID == edge.Target.NodeID {
			result.Errors = append(result.Errors, &ValidationError{
				Code:    "SELF_LOOP",
				Message: "self-loop detected",
				EdgeID:  edge.ID,
				NodeID:  edge.Source.NodeID,
			})
		}
	}
}

// validateNodeTypes checks that node types are valid.
func validateNodeTypes(wf *Workflow, result *ValidationResult) {
	validTypes := ValidNodeTypesMap()

	for _, node := range wf.Nodes {
		if !validTypes[node.Type] {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("node %s (%s): unknown node type '%s'", node.Name, node.ID, node.Type))
		}
	}
}

// validateVariables checks that variable references are valid.
func validateVariables(wf *Workflow, result *ValidationResult) {
	for _, node := range wf.Nodes {
		for key, val := range node.Config {
			if str, ok := val.(string); ok {
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

// typeCompatibilitySet defines which data types are compatible, using sets for O(1) lookup.
var typeCompatibilitySet = map[DataType]map[DataType]bool{}

func initCompat() {
	typeCompatibilitySet = map[DataType]map[DataType]bool{
		DataTypeImage:   {DataTypeImage: true, DataTypeTensor: true, DataTypeAny: true},
		DataTypeTensor:  {DataTypeTensor: true, DataTypeModel: true, DataTypeAny: true},
		DataTypeDataset: {DataTypeDataset: true, DataTypeAny: true},
		DataTypeModel:   {DataTypeModel: true, DataTypeAny: true},
		DataTypeText:    {DataTypeText: true, DataTypeJSON: true, DataTypeAny: true},
		DataTypeNumber:  {DataTypeNumber: true, DataTypeJSON: true, DataTypeAny: true},
		DataTypeBoolean: {DataTypeBoolean: true, DataTypeAny: true},
		DataTypeJSON:    {DataTypeJSON: true, DataTypeText: true, DataTypeAny: true},
		DataTypeFile:    {DataTypeFile: true, DataTypeAny: true},
		DataTypeStream:  {DataTypeStream: true, DataTypeAny: true},
		DataTypeAny:     {DataTypeAny: true},
	}
}

// isTypeCompatible checks if sourceType can be connected to targetType.
func isTypeCompatible(sourceType, targetType DataType) bool {
	compatible, ok := typeCompatibilitySet[sourceType]
	if !ok {
		return sourceType == targetType
	}
	return compatible[targetType] || sourceType == targetType
}

// ============================================================================
// Valid Helpers
// ============================================================================

// ValidTargets returns all valid target platforms. Result is cached.
func ValidTargets() []Target {
	onceValidTargets.Do(func() {
		cachedValidTargets = []Target{
			TargetPython,
			TargetMATLAB,
			TargetROS2,
			TargetSTM32,
			TargetDocker,
			TargetCPP,
			TargetUnity,
			TargetJava,
		}
	})
	return cachedValidTargets
}

// ValidTargetsMap returns a map of valid target platforms for O(1) lookup.
func ValidTargetsMap() map[Target]bool {
	onceValidTargets.Do(func() {
		cachedValidTargets = []Target{
			TargetPython,
			TargetMATLAB,
			TargetROS2,
			TargetSTM32,
			TargetDocker,
			TargetCPP,
			TargetUnity,
			TargetJava,
		}
	})
	m := make(map[Target]bool, len(cachedValidTargets))
	for _, t := range cachedValidTargets {
		m[t] = true
	}
	return m
}

// ValidNodeTypes returns all valid node types including DSL-native control nodes. Result is cached.
func ValidNodeTypes() []NodeType {
	onceValidNodeTypes.Do(func() {
		cachedValidNodeTypes = []NodeType{
			NodeTypeCondition,
			NodeTypeLoop,
			NodeTypeSwitch,
			NodeTypeRetry,
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
	})
	return cachedValidNodeTypes
}

// ValidNodeTypesMap returns a map of valid node types for O(1) lookup.
func ValidNodeTypesMap() map[NodeType]bool {
	onceValidNodeTypes.Do(func() {
		cachedValidNodeTypes = []NodeType{
			NodeTypeCondition,
			NodeTypeLoop,
			NodeTypeSwitch,
			NodeTypeRetry,
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
	})
	m := make(map[NodeType]bool, len(cachedValidNodeTypes))
	for _, t := range cachedValidNodeTypes {
		m[t] = true
	}
	return m
}

// ValidDataTypes returns all valid data types. Result is cached.
func ValidDataTypes() []DataType {
	onceValidDataTypes.Do(func() {
		cachedValidDataTypes = []DataType{
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
	})
	return cachedValidDataTypes
}

// ============================================================================
// DAG (Directed Acyclic Graph) Operations
// ============================================================================

// TopologicalSort performs a topological sort of the workflow nodes.
func TopologicalSort(wf *Workflow) ([]Node, error) {
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	for _, node := range wf.Nodes {
		inDegree[node.ID] = 0
	}

	for _, edge := range wf.Edges {
		adjList[edge.Source.NodeID] = append(adjList[edge.Source.NodeID], edge.Target.NodeID)
		inDegree[edge.Target.NodeID]++
	}

	var queue []string
	for _, node := range wf.Nodes {
		if inDegree[node.ID] == 0 {
			queue = append(queue, node.ID)
		}
	}

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

// makeNodeMap creates a map from node ID to Node.
func makeNodeMap(wf *Workflow) map[string]Node {
	nodeMap := make(map[string]Node, len(wf.Nodes))
	for _, node := range wf.Nodes {
		nodeMap[node.ID] = node
	}
	return nodeMap
}
