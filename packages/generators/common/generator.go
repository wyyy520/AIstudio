// Package common defines the public Generator interface and types
// that all external generator implementations must follow.
//
// External generator developers implement this interface to create
// new compilation targets for AIStudio. The interface mirrors the
// internal compiler.Generator but uses standalone types so that
// external packages do not need to import AIStudio internals.
//
// Usage:
//
//	type MyGenerator struct {
//	    common.BaseGenerator
//	}
//
//	func (g *MyGenerator) Generate(ctx context.Context, wf *common.Workflow, opts common.CompileOptions) (*common.GenerateResult, error) {
//	    // ... generate project files
//	}
package common

import "context"

// ============================================================================
// Public Types
// ============================================================================

// Target identifies the target platform for compilation.
type Target string

// Workflow represents a workflow for generation.
type Workflow struct {
	SchemaVersion string            `json:"schema_version"`
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Version       int               `json:"version"`
	Author        string            `json:"author,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	Metadata      map[string]any    `json:"metadata,omitempty"`
	Variables     map[string]any    `json:"variables,omitempty"`
	Target        Target            `json:"target"`
	Nodes         []Node            `json:"nodes"`
	Edges         []Edge            `json:"edges"`
}

// Node represents a single step in the workflow DAG.
type Node struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Config      map[string]any    `json:"config,omitempty"`
	Inputs      []Port            `json:"inputs,omitempty"`
	Outputs     []Port            `json:"outputs,omitempty"`
}

// Port defines a typed input or output of a node.
type Port struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required,omitempty"`
}

// Edge represents a connection between two nodes.
type Edge struct {
	ID     string       `json:"id"`
	Source EdgeEndpoint `json:"source"`
	Target EdgeEndpoint `json:"target"`
}

// EdgeEndpoint identifies a specific port on a node.
type EdgeEndpoint struct {
	NodeID string `json:"node_id"`
	PortID string `json:"port_id"`
}

// CompileOptions controls the compilation behavior.
type CompileOptions struct {
	OutputDir   string
	Target      Target
	Variables   map[string]string
	Force       bool
	DryRun      bool
	ProjectName string
}

// GenerateResult contains the generation output.
type GenerateResult struct {
	Target      Target
	ProjectRoot string
	EntryPoints []string
	Files       []GeneratedFile
	ProjectName string
}

// GeneratedFile represents a single generated file.
type GeneratedFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Mode    uint32 `json:"mode"`
}

// ResourceEstimate describes estimated resources for generation.
type ResourceEstimate struct {
	EstimatedFiles  int   `json:"estimatedFiles"`
	EstimatedSizeKB int   `json:"estimatedSizeKB"`
	RequiresGPU     bool  `json:"requiresGpu"`
	MinMemoryMB     int   `json:"minMemoryMb"`
	MinDiskMB       int   `json:"minDiskMb"`
}

// RuntimeRequirement declares what runtime environment is needed.
type RuntimeRequirement struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Python      string   `json:"python"`
	Packages    []string `json:"packages"`
	Commands    []string `json:"commands"`
	GPU         bool     `json:"gpu"`
	MinMemoryMB int      `json:"minMemoryMb"`
	MinDiskMB   int      `json:"minDiskMb"`
}

// CompilePlan describes what will be generated before doing it.
type CompilePlan struct {
	GeneratorID      Target             `json:"generatorId"`
	GeneratorName    string             `json:"generatorName"`
	ProjectName      string             `json:"projectName"`
	OutputDir        string             `json:"outputDir"`
	EstimatedFiles   int                `json:"estimatedFiles"`
	EstimatedSizeKB  int                `json:"estimatedSizeKB"`
	Validated        bool               `json:"validated"`
	Warnings         []string           `json:"warnings,omitempty"`
	RuntimeReq       *RuntimeRequirement `json:"runtimeRequirement,omitempty"`
}

// ============================================================================
// Generator Interface
//
// This is the public API contract for all external generator implementations.
// ============================================================================

// Generator generates a complete project from a workflow.
// Implement this interface to add a new compilation target.
type Generator interface {
	// ID returns the unique generator identifier (e.g., "python", "matlab").
	ID() Target

	// Name returns the human-readable generator name.
	Name() string

	// Description returns a description of what this generator produces.
	Description() string

	// Version returns the generator version.
	Version() string

	// Generate generates a complete project from a workflow.
	Generate(ctx context.Context, wf *Workflow, opts CompileOptions) (*GenerateResult, error)

	// RuntimeRequirement returns the runtime bundle required for this project.
	RuntimeRequirement(wf *Workflow) (*RuntimeRequirement, error)

	// Validate checks if the workflow can be compiled to this target.
	Validate(wf *Workflow) error

	// EstimateResources estimates the resources needed for generation.
	EstimateResources(wf *Workflow) (*ResourceEstimate, error)

	// CompileTimeValidate checks if the host system has the required tools.
	CompileTimeValidate(ctx context.Context) error

	// Plan returns a compilation plan without writing files.
	Plan(ctx context.Context, wf *Workflow, opts CompileOptions) (*CompilePlan, error)
}

// ============================================================================
// BaseGenerator
// ============================================================================

// BaseGenerator provides default implementations for common Generator methods.
type BaseGenerator struct {
	TargetID      Target
	GeneratorName string
	GeneratorDesc string
	GeneratorVer  string
}

func (b *BaseGenerator) ID() Target         { return b.TargetID }
func (b *BaseGenerator) Name() string        { return b.GeneratorName }
func (b *BaseGenerator) Description() string  { return b.GeneratorDesc }
func (b *BaseGenerator) Version() string      { return b.GeneratorVer }

// DefaultValidate performs basic validation common to all generators.
func (b *BaseGenerator) DefaultValidate(wf *Workflow) error {
	return nil
}

// DefaultCompileTimeValidate checks for basic host tools.
func (b *BaseGenerator) DefaultCompileTimeValidate(ctx context.Context) error {
	return nil
}

// DefaultEstimateResources provides a conservative resource estimate.
func (b *BaseGenerator) DefaultEstimateResources(wf *Workflow) (*ResourceEstimate, error) {
	return &ResourceEstimate{
		EstimatedFiles:  len(wf.Nodes) + 8,
		EstimatedSizeKB: len(wf.Nodes) * 10 + 50,
		MinMemoryMB:     512,
		MinDiskMB:       100,
	}, nil
}

// DefaultPlan returns a basic plan from available metadata.
func (b *BaseGenerator) DefaultPlan(ctx context.Context, wf *Workflow, opts CompileOptions) (*CompilePlan, error) {
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
