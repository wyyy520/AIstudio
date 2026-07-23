package compiler

import (
	"fmt"
	"log"
	"time"

	"github.com/aistudio/backend/internal/workflow"
)

// ============================================================================
// BuildExecutionPlan
//
// BuildExecutionPlan transforms a validated workflow and compile options into an
// intermediate ExecutionPlan. The plan contains the resolved execution order,
// per-node plans with dependencies, and metadata needed by generators.
//
// This is the bridge between the Workflow DSL and the Generator layer.
// ============================================================================

// BuildExecutionPlan validates the workflow, topologically sorts its nodes,
// and produces an ExecutionPlan that generators can consume directly.
func BuildExecutionPlan(wf *workflow.Workflow, opts CompileOptions) (*ExecutionPlan, error) {
	// 1. Validate workflow using the internal workflow validator
	validationResult := workflow.Validate(wf)
	if !validationResult.Valid {
		errMsg := fmt.Sprintf("workflow validation failed with %d error(s)", len(validationResult.Errors))
		for _, ve := range validationResult.Errors {
			errMsg += "\n  - " + ve.Error()
		}
		return nil, &CompileError{
			Code:    "VALIDATION_FAILED",
			Message: errMsg,
		}
	}

	// Log warnings
	for _, w := range validationResult.Warnings {
		log.Printf("[compiler] workflow validation warning: %s", w)
	}

	// 2. Topological sort to determine execution order
	sortedNodes, err := workflow.TopologicalSort(wf)
	if err != nil {
		return nil, &CompileError{
			Code:    "TOPOLOGICAL_SORT_FAILED",
			Message: fmt.Sprintf("topological sort failed: %s", err.Error()),
		}
	}

	// 3. Determine target
	target := wf.Target
	if opts.Target != "" {
		target = opts.Target
	}

	// 4. Build per-node plans with dependency information
	nodePlans := make(map[string]NodeExecutionPlan, len(sortedNodes))
	executionOrder := make([]string, 0, len(sortedNodes))

	for order, node := range sortedNodes {
		// Gather dependencies from input edges
		inputEdges := workflow.GetInputEdges(wf, node.ID)
		deps := make([]string, 0, len(inputEdges))
		for _, edge := range inputEdges {
			deps = append(deps, edge.Source.NodeID)
		}

		nodePlans[node.ID] = NodeExecutionPlan{
			NodeID:       node.ID,
			NodeType:     string(node.Type),
			NodeName:     node.Name,
			Order:        order,
			Dependencies: deps,
			Config:       node.Config,
			Domain:       node.Domain,
		}
		executionOrder = append(executionOrder, node.ID)
	}

	// 5. Determine project name
	projectName := opts.ProjectName
	if projectName == "" {
		projectName = sanitizeName(wf.Name)
	}

	// 6. Build and return the ExecutionPlan
	plan := &ExecutionPlan{
		PlanVersion:    "1.0.0",
		WorkflowID:     wf.ID,
		WorkflowName:   wf.Name,
		CreatedAt:      time.Now(),
		GeneratorID:    target,
		ExecutionOrder: executionOrder,
		NodePlans:      nodePlans,
		OutputDir:      opts.OutputDir,
		ProjectName:    projectName,
		Workflow:       wf,
	}

	log.Printf("[compiler] execution plan built: workflow=%q target=%s nodes=%d planVersion=%s",
		wf.Name, target, len(sortedNodes), plan.PlanVersion)

	return plan, nil
}
