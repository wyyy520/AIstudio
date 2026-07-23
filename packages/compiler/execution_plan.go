package compiler

import (
	"time"

	"github.com/aistudio/packages/workflow"
)

// ExecutionPlan is the intermediate representation between Compiler and Generator.
// The Compiler produces this from a Workflow; Generators consume it.
// This is the sole output of the Compiler — generators must NOT read workflow.json directly.
// Corresponds to silu.md section 3.10.
type ExecutionPlan struct {
	// Plan metadata
	PlanVersion  string    `json:"plan_version"`
	WorkflowID   string    `json:"workflow_id"`
	WorkflowName string    `json:"workflow_name"`
	CreatedAt    time.Time `json:"created_at"`

	// Target/domain info
	GeneratorID      workflow.Target   `json:"generator_id"`
	GeneratorVersion string            `json:"generator_version"`
	SourceTarget     workflow.Target   `json:"source_target"`
	Domains          []workflow.Target `json:"domains,omitempty"`

	// Execution order (result of topological sort) — node IDs in sequence
	ExecutionOrder []string `json:"execution_order"`

	// Per-node execution info, keyed by node ID
	NodePlans map[string]NodeExecutionPlan `json:"node_plans"`

	// Template info
	TemplateBase string `json:"template_base,omitempty"`
	OutputDir    string `json:"output_dir,omitempty"`
	ProjectName  string `json:"project_name,omitempty"`

	// Runtime requirements
	RuntimeReq *RuntimeRequirement `json:"runtime_requirement,omitempty"`

	// Plugin info
	Plugins []PluginExecution `json:"plugins,omitempty"`
}

// NodeExecutionPlan describes how a single node should be executed/generated.
type NodeExecutionPlan struct {
	NodeID       string            `json:"node_id"`
	NodeType     string            `json:"node_type"`
	NodeName     string            `json:"node_name"`
	Order        int               `json:"order"`
	Dependencies []string          `json:"dependencies"`
	Config       map[string]any    `json:"config,omitempty"`
	TemplatePath string            `json:"template_path,omitempty"`
	InputFiles   []string          `json:"input_files,omitempty"`
	OutputFiles  []string          `json:"output_files,omitempty"`
	Domain       string            `json:"domain,omitempty"`
	Generator    string            `json:"generator,omitempty"`
}

// PluginExecution describes a plugin used during execution.
type PluginExecution struct {
	PluginID      string   `json:"plugin_id"`
	PluginVersion string   `json:"plugin_version"`
	NodeIDs       []string `json:"node_ids"`
	TemplateDirs  []string `json:"template_dirs,omitempty"`
}

// BuildExecutionPlan builds an ExecutionPlan from a Workflow.
// It performs validation, topological sorting, and resolves templates.
func BuildExecutionPlan(wf *workflow.Workflow, opts CompileOptions) (*ExecutionPlan, error) {
	// 1. Validate workflow
	validationResult := workflow.ValidateWorkflow(wf)
	if !validationResult.Valid {
		return nil, &CompileError{
			Code:    "WORKFLOW_VALIDATION_FAILED",
			Message: "workflow validation failed",
			Details: validationResult.Errors,
		}
	}

	// 2. Topological sort
	sortedNodes, err := workflow.TopologicalSort(wf)
	if err != nil {
		return nil, &CompileError{
			Code:    "TOPOLOGICAL_SORT_FAILED",
			Message: err.Error(),
		}
	}

	// 3. Build execution steps
	steps := make([]string, 0, len(sortedNodes))
	nodePlans := make(map[string]NodeExecutionPlan, len(sortedNodes))
	domains := make(map[workflow.Target]bool)

	for i, node := range sortedNodes {
		steps = append(steps, node.ID)

		// Gather dependencies from input edges
		var deps []string
		inputEdges := workflow.GetInputEdges(wf, node.ID)
		for _, edge := range inputEdges {
			deps = append(deps, edge.Source.NodeID)
		}

		nodeDomain := node.Domain
		if nodeDomain == "" {
			nodeDomain = string(wf.Target)
		}

		nodePlans[node.ID] = NodeExecutionPlan{
			NodeID:       node.ID,
			NodeType:     string(node.Type),
			NodeName:     node.Name,
			Order:        i,
			Dependencies: deps,
			Config:       node.Config,
			Domain:       nodeDomain,
		}
		domains[workflow.Target(nodeDomain)] = true
	}

	// 4. Build domain list
	domainList := make([]workflow.Target, 0, len(domains))
	for d := range domains {
		domainList = append(domainList, d)
	}

	// 5. Resolve project name
	projectName := opts.ProjectName
	if projectName == "" {
		projectName = sanitizeName(wf.Name)
	}

	plan := &ExecutionPlan{
		PlanVersion:    "1.0.0",
		WorkflowID:     wf.ID,
		WorkflowName:   wf.Name,
		CreatedAt:      time.Now(),
		GeneratorID:    wf.Target,
		SourceTarget:   wf.Target,
		Domains:        domainList,
		ExecutionOrder: steps,
		NodePlans:      nodePlans,
		OutputDir:      opts.OutputDir,
		ProjectName:    projectName,
	}

	return plan, nil
}

// CompileError represents a compilation error.
type CompileError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *CompileError) Error() string {
	return "[" + e.Code + "] " + e.Message
}
