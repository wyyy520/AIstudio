package compiler

import (
	"time"

	"github.com/aistudio/backend/internal/workflow"
)

// ============================================================================
// ExecutionPlan — Intermediate representation between Compiler and Generator
// ============================================================================

// ExecutionPlan is the intermediate representation between Compiler and Generator.
// It represents the resolved, validated, and ordered plan for generating a project
// from a workflow DAG.
type ExecutionPlan struct {
	PlanVersion    string                `json:"plan_version"`
	WorkflowID     string                `json:"workflow_id"`
	WorkflowName   string                `json:"workflow_name"`
	CreatedAt      time.Time             `json:"created_at"`
	GeneratorID    workflow.Target       `json:"generator_id"`
	Domains        []workflow.Target     `json:"domains,omitempty"`
	ExecutionOrder []string              `json:"execution_order"`
	NodePlans      map[string]NodeExecutionPlan `json:"node_plans"`
	OutputDir      string                `json:"output_dir,omitempty"`
	ProjectName    string                `json:"project_name,omitempty"`
	RuntimeReq     *RuntimeRequirement   `json:"runtime_requirement,omitempty"`
	Workflow       *workflow.Workflow    `json:"-"` // Full workflow data for generators
}

// NodeExecutionPlan describes how a single workflow node should be generated.
type NodeExecutionPlan struct {
	NodeID       string         `json:"node_id"`
	NodeType     string         `json:"node_type"`
	NodeName     string         `json:"node_name"`
	Order        int            `json:"order"`
	Dependencies []string       `json:"dependencies"`
	Config       map[string]any `json:"config,omitempty"`
	TemplatePath string         `json:"template_path,omitempty"`
	Domain       string         `json:"domain,omitempty"`
}

// CompileError represents a structured compilation error with a machine-readable code.
type CompileError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *CompileError) Error() string {
	return "[" + e.Code + "] " + e.Message
}
