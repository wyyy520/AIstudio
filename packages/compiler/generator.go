package compiler

import (
	"context"

	"github.com/aistudio/packages/workflow"
)

// Generator generates a complete project from a workflow.
// Each Generator is responsible for one target platform.
type Generator interface {
	// ID returns the unique generator identifier (e.g., "python", "matlab").
	ID() workflow.Target

	// Name returns the human-readable generator name.
	Name() string

	// Description returns a description of what this generator produces.
	Description() string

	// Version returns the generator version.
	Version() string

	// Generate generates a complete project from a workflow.
	Generate(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*GenerateResult, error)

	// RuntimeRequirement returns the runtime bundle required for this project.
	RuntimeRequirement(wf *workflow.Workflow) (*RuntimeRequirement, error)

	// Validate checks if the workflow can be compiled to this target.
	Validate(wf *workflow.Workflow) error

	// EstimateResources estimates the resources needed for generation.
	EstimateResources(wf *workflow.Workflow) (*ResourceEstimate, error)

	// CompileTimeValidate checks if the host system has the required tools.
	CompileTimeValidate(ctx context.Context) error

	// Plan returns a compilation plan without writing files.
	Plan(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*CompilePlan, error)
}

// BaseGenerator provides common generator functionality.
// Embed this in your generator to get default implementations.
type BaseGenerator struct {
	TargetID      workflow.Target
	GeneratorName string
	GeneratorDesc string
	GeneratorVer  string
}

func (b *BaseGenerator) ID() workflow.Target  { return b.TargetID }
func (b *BaseGenerator) Name() string          { return b.GeneratorName }
func (b *BaseGenerator) Description() string   { return b.GeneratorDesc }
func (b *BaseGenerator) Version() string       { return b.GeneratorVer }

// DefaultValidate performs basic validation common to all generators.
func (b *BaseGenerator) DefaultValidate(wf *workflow.Workflow) error {
	return nil
}

// DefaultEstimateResources provides a conservative resource estimate.
func (b *BaseGenerator) DefaultEstimateResources(wf *workflow.Workflow) (*ResourceEstimate, error) {
	return &ResourceEstimate{
		EstimatedFiles:  len(wf.Nodes) + 8,
		EstimatedSizeKB: len(wf.Nodes)*10 + 50,
		MinMemoryMB:     512,
		MinDiskMB:       100,
	}, nil
}

// DefaultCompileTimeValidate is a no-op validation.
func (b *BaseGenerator) DefaultCompileTimeValidate(ctx context.Context) error {
	return nil
}

// DefaultPlan returns a basic plan from available metadata.
func (b *BaseGenerator) DefaultPlan(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*CompilePlan, error) {
	est, _ := b.DefaultEstimateResources(wf)
	return &CompilePlan{
		GeneratorID:     b.TargetID,
		GeneratorName:   b.GeneratorName,
		ProjectName:     opts.ProjectName,
		OutputDir:       opts.OutputDir,
		EstimatedFiles:  est.EstimatedFiles,
		EstimatedSizeKB: est.EstimatedSizeKB,
		Validated:       true,
	}, nil
}