package workflow

import "time"

const (
	NodeStatusIdle      = "idle"
	NodeStatusPending   = "pending"
	NodeStatusRunning   = "running"
	NodeStatusSuccess   = "success"
	NodeStatusError     = "error"
	NodeStatusCancelled = "cancelled"
	NodeStatusSkipped   = "skipped"

	NodeTypeVision      = "vision"
	NodeTypeNLP         = "nlp"
	NodeTypeTimeseries  = "timeseries"
	NodeTypeLogic       = "logic"
	NodeTypeSystem      = "system"
	NodeTypeSimulation  = "simulation"
	NodeTypeMCP         = "mcp"
	NodeTypeInput       = "input"
	NodeTypeOutput      = "output"
	NodeTypeAgent       = "agent"
	NodeTypeSubworkflow = "subworkflow"

	SchemaVersion = "1.0.0"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type PortConstraints struct {
	MaxSizeMB *int     `json:"max_size_mb,omitempty"`
	Formats   []string `json:"formats,omitempty"`
}

type Port struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        string           `json:"type"`
	Required    bool             `json:"required"`
	Multiple    bool             `json:"multiple,omitempty"`
	Default     interface{}      `json:"default,omitempty"`
	Description string           `json:"description,omitempty"`
	Accepts     []string         `json:"accepts,omitempty"`
	Constraints *PortConstraints `json:"constraints,omitempty"`
}

type EdgeEndpoint struct {
	NodeID string `json:"node_id"`
	PortID string `json:"port_id"`
}

type Edge struct {
	ID        string       `json:"id"`
	Source    EdgeEndpoint `json:"source"`
	Target    EdgeEndpoint `json:"target"`
	Label     string       `json:"label,omitempty"`
	Animated  bool         `json:"animated,omitempty"`
	Condition string       `json:"condition,omitempty"`
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

type NodeRuntime struct {
	Status         string                 `json:"status"`
	Progress       float64                `json:"progress,omitempty"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	FinishedAt     *time.Time             `json:"finished_at,omitempty"`
	DurationMs     *int64                 `json:"duration_ms,omitempty"`
	InputSnapshot  map[string]interface{} `json:"input_snapshot,omitempty"`
	OutputSnapshot map[string]interface{} `json:"output_snapshot,omitempty"`
	Error          string                 `json:"error,omitempty"`
	Metrics        map[string]float64     `json:"metrics,omitempty"`
	Logs           []LogEntry             `json:"logs,omitempty"`
}

type NodeConstraints struct {
	GPURequired bool `json:"gpu_required"`
	MaxRetries  int  `json:"max_retries"`
	TimeoutMs   int  `json:"timeout_ms"`
}

type Node struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Plugin      string                 `json:"plugin"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Position    Point                  `json:"position"`
	Size        *Size                  `json:"size,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Inputs      []Port                 `json:"inputs"`
	Outputs     []Port                 `json:"outputs"`
	Runtime     *NodeRuntime           `json:"runtime,omitempty"`
	Constraints *NodeConstraints       `json:"constraints,omitempty"`
}

type Workflow struct {
	SchemaVersion string                 `json:"schema_version"`
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	ProjectID     string                 `json:"project_id"`
	Version       int                    `json:"version"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Author        string                 `json:"author,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Nodes         []Node                 `json:"nodes"`
	Edges         []Edge                 `json:"edges"`
}

type ExecutionResult struct {
	TaskID      string                `json:"task_id"`
	Status      string                `json:"status"`
	Progress    float64               `json:"progress"`
	NodeOutputs map[string]NodeResult `json:"node_outputs,omitempty"`
	Error       string                `json:"error,omitempty"`
}

type NodeResult struct {
	Status   string                 `json:"status"`
	Outputs  map[string]interface{} `json:"outputs,omitempty"`
	Duration int64                  `json:"duration_ms"`
	Error    string                 `json:"error,omitempty"`
}
