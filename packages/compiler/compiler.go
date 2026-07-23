// Package compiler transforms Workflow DSL into real engineering projects.
//
// The Compiler is the heart of AIStudio. It reads a declarative Workflow,
// selects the appropriate Generator, and produces a complete, runnable project
// on the filesystem.
//
// Architecture (EngStudio.md §3.5, §16.4):
//
//   Workflow → Compiler → Generator Registry → Generator → Project
//
// Pipeline stages:
//
//	1. Graph Optimization (dead nodes, cycles, fusion)
//	2. Planning (generator selection, estimation)
//	3. Host validation (compile-time tool checks)
//	4. Execution Plan (DAG → linear execution order)
//	5. EWIR Split (UI ↔ Engineering separation)
//	6. Domain Dispatch (multi-domain routing)
//	7. Plugin Manifest (dependency declaration)
//	8. Generate (template-driven code generation)
//	9. Verify (output validation)
//	10. Complete (result assembly)
//
// The Compiler does NOT:
//   - Execute projects (that's Runtime's job)
//   - Modify projects (that's the user's job)
//   - Install dependencies (that's Environment's job)
//   - Generate code from AI (that's Agent's job)
package compiler

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/aistudio/packages/event"
	"github.com/aistudio/packages/workflow"
)

const (
	BufferSize   = 4096
	numWorkers   = 4
	genCacheSize = 16
)

// ============================================================================
// Compiler Interface
// ============================================================================

// Compiler compiles a Workflow into a project.
// It is the single entry point for all project generation.
// Compiler is the ONLY module that reads workflow.json.
type Compiler interface {
	// Compile compiles a workflow into a project directory.
	Compile(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*CompileResult, error)

	// Plan returns a compilation plan without writing any files.
	Plan(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*CompilePlan, error)

	// ListTargets returns all available compilation targets.
	ListTargets() []TargetInfo

	// ListAllTargets returns all available compilation targets (alias).
	ListAllTargets() []TargetInfo

	// GetGenerator returns the generator for a given target.
	GetGenerator(target workflow.Target) (Generator, bool)

	// RegisterGenerator registers a new generator.
	RegisterGenerator(g Generator) error

	// SetOutputBase sets the base output directory for generated projects.
	SetOutputBase(dir string)
}

// ============================================================================
// Constructor
// ============================================================================

// NewCompiler creates a new Compiler with the given event bus.
func NewCompiler(bus *event.EventBus) Compiler {
	return &compilerImpl{
		registry:          NewGeneratorRegistry(),
		bus:               bus,
		graphOptimizer:    NewGraphOptimizer(),
		ewirBuilder:       NewEWIRBuilder("1.0.0"),
		manifestGenerator: NewPluginManifestGenerator(),
		domainDispatcher:  NewDomainDispatcher(),
		templateEngine:    NewTemplateEngine(),
	}
}

// ============================================================================
// Implementation
// ============================================================================

type compilerImpl struct {
	registry          *GeneratorRegistry
	bus               *event.EventBus
	graphOptimizer    *GraphOptimizer
	ewirBuilder       *EWIRBuilder
	manifestGenerator *PluginManifestGenerator
	domainDispatcher  *DomainDispatcher
	templateEngine    *TemplateEngine
	outputBase        string
}

func (c *compilerImpl) SetOutputBase(dir string) {
	c.outputBase = dir
}

// Plan returns a compilation plan without writing files (dry-run).
func (c *compilerImpl) Plan(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*CompilePlan, error) {
	target := wf.Target
	if opts.Target != "" {
		target = opts.Target
	}

	generator, ok := c.registry.Get(target)
	if !ok {
		return nil, fmt.Errorf("no generator registered for target %q (available: %s)", target, c.availableTargets())
	}

	var warnings []string
	if err := generator.Validate(wf); err != nil {
		warnings = append(warnings, err.Error())
	}

	rr, _ := generator.RuntimeRequirement(wf)

	est, _ := generator.EstimateResources(wf)
	estFiles := 0
	estSizeKB := 0
	if est != nil {
		estFiles = est.EstimatedFiles
		estSizeKB = est.EstimatedSizeKB
	}

	projectName := opts.ProjectName
	if projectName == "" {
		projectName = sanitizeName(wf.Name)
	}

	return &CompilePlan{
		GeneratorID:     target,
		GeneratorName:   generator.Name(),
		ProjectName:     projectName,
		OutputDir:       opts.OutputDir,
		EstimatedFiles:  estFiles,
		EstimatedSizeKB: estSizeKB,
		Validated:       len(warnings) == 0,
		Warnings:        warnings,
		RuntimeReq:      rr,
	}, nil
}

// Compile compiles a workflow into a project directory using the full 10-stage pipeline.
func (c *compilerImpl) Compile(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*CompileResult, error) {
	start := time.Now()

	target := wf.Target
	if opts.Target != "" {
		target = opts.Target
	}

	// ================================================================
	// Stage 1: Publish started event
	// ================================================================
	c.publishProgress(event.TopicCompileStarted, wf.ID, target, "", 0.0, "")

	// ================================================================
	// Stage 2: Find Generator
	// ================================================================
	generator, ok := c.registry.Get(target)
	if !ok {
		err := fmt.Errorf("no generator registered for target %q (available: %s)", target, c.availableTargets())
		c.publishProgress(event.TopicCompileFailed, wf.ID, target, "", 0.0, err.Error())
		return nil, err
	}

	// ================================================================
	// Stage 3: Graph Optimization (dead nodes, cycles, fusion)
	// ================================================================
	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.05, "optimizing graph")

	optResult := c.graphOptimizer.Optimize(wf.Nodes, wf.Edges)
	if len(optResult.Warnings) > 0 {
		for _, w := range optResult.Warnings {
			log.Printf("[compiler] optimizer: %s", w)
		}
	}

	// Check for cycles
	hasCycle, cycleNodes := c.graphOptimizer.DetectCycles(optResult.Nodes, optResult.Edges)
	if hasCycle {
		err := fmt.Errorf("workflow contains cycle involving nodes: %v", cycleNodes)
		c.publishProgress(event.TopicCompileFailed, wf.ID, target, "", 0.0, err.Error())
		return nil, err
	}

	// Update workflow with optimized nodes/edges
	wf.Nodes = optResult.Nodes
	wf.Edges = optResult.Edges

	if optResult.OriginalCount != optResult.OptimizedCount {
		log.Printf("[compiler] graph optimized: %d → %d nodes", optResult.OriginalCount, optResult.OptimizedCount)
	}

	// ================================================================
	// Stage 4: Plan (generator selection, resource estimation)
	// ================================================================
	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.1, "planning")

	plan, planErr := c.Plan(ctx, wf, opts)
	if planErr != nil {
		c.publishProgress(event.TopicCompileFailed, wf.ID, target, "", 0.0, planErr.Error())
		return nil, fmt.Errorf("compilation planning failed: %w", planErr)
	}

	if len(plan.Warnings) > 0 {
		for _, w := range plan.Warnings {
			log.Printf("[compiler] warning: %s", w)
		}
	}

	// ================================================================
	// Stage 5: Compile-time host validation
	// ================================================================
	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.2, "validating host environment")

	if err := generator.CompileTimeValidate(ctx); err != nil {
		c.publishProgress(event.TopicCompileFailed, wf.ID, target, "", 0.0, err.Error())
		return nil, fmt.Errorf("host validation failed: %w", err)
	}

	// ================================================================
	// Stage 6: Build Execution Plan (validate workflow, topological sort)
	// ================================================================
	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.3, "building execution plan")

	execPlan, planErr := BuildExecutionPlan(wf, opts)
	if planErr != nil {
		c.publishProgress(event.TopicCompileFailed, wf.ID, target, "", 0.0, planErr.Error())
		return nil, fmt.Errorf("execution plan build failed: %w", planErr)
	}

	// Attach runtime requirement from generator
	if rr, rrErr := generator.RuntimeRequirement(wf); rrErr == nil {
		execPlan.RuntimeReq = rr
	}

	log.Printf("[compiler] execution plan: workflow=%q target=%s planVersion=%s nodes=%d",
		wf.Name, target, execPlan.PlanVersion, len(execPlan.NodePlans))

	// ================================================================
	// Stage 7: Build EWIR (UI ↔ Engineering separation)
	// ================================================================
	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.5, "building EWIR")

	ewirOutputDir := opts.OutputDir
	if ewirOutputDir == "" {
		if c.outputBase != "" {
			ewirOutputDir = filepath.Join(c.outputBase, ".engstudio")
		} else {
			ewirOutputDir = filepath.Join("projects", wf.ID, ".engstudio")
		}
	}

	splitResult, err := c.ewirBuilder.Split(wf, ewirOutputDir)
	if err != nil {
		log.Printf("[compiler] warning: failed to build EWIR: %v", err)
	} else {
		log.Printf("[compiler] EWIR built: ui.json=%s, workflow.ir.json=%s", splitResult.UIPath, splitResult.EWIRPath)
	}

	// ================================================================
	// Stage 8: Domain Dispatch (multi-domain routing)
	// ================================================================
	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.6, "dispatching domains")

	domains := c.domainDispatcher.GetDomains(execPlan)
	if len(domains) > 1 {
		domainPlans, dispatchErr := c.domainDispatcher.Dispatch(execPlan)
		if dispatchErr != nil {
			log.Printf("[compiler] warning: domain dispatch failed: %v", dispatchErr)
		} else {
			log.Printf("[compiler] dispatched to %d domains: %v", len(domainPlans), domains)
			execPlan.Domains = domains
		}
	} else {
		log.Printf("[compiler] single domain: %s", target)
	}

	// ================================================================
	// Stage 9: Generate Plugin Manifest
	// ================================================================
	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.7, "generating plugin manifest")

	manifestPath, err := c.manifestGenerator.GenerateAndWrite(wf, ewirOutputDir)
	if err != nil {
		log.Printf("[compiler] warning: failed to generate plugin manifest: %v", err)
	} else {
		log.Printf("[compiler] plugin manifest generated: %s", manifestPath)
	}

	// ================================================================
	// Stage 10: Generate (template-driven code generation)
	// ================================================================
	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.8, "generating project")

	genResult, genErr := generator.Generate(ctx, execPlan, opts)
	if genErr != nil {
		c.publishProgress(event.TopicCompileFailed, wf.ID, target, "", 0.0, genErr.Error())
		return nil, fmt.Errorf("generation failed: %w", genErr)
	}

	// ================================================================
	// Stage 11: Verify output
	// ================================================================
	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.95, "verifying output")

	var runtimeReq *RuntimeRequirement
	if rr, err := generator.RuntimeRequirement(wf); err == nil {
		runtimeReq = rr
	}

	// ================================================================
	// Stage 12: Complete — assemble result
	// ================================================================
	result := &CompileResult{
		Target:      target,
		ProjectRoot: genResult.ProjectRoot,
		EntryPoints: genResult.EntryPoints,
		Files:       genResult.Files,
		RuntimeReq:  runtimeReq,
		Duration:    time.Since(start),
		WorkflowID:  wf.ID,
		GeneratorID: string(generator.ID()),
	}

	c.publishProgress(event.TopicCompileCompleted, wf.ID, target, result.ProjectRoot, 1.0, result.Duration.String())

	log.Printf("[compiler] compiled workflow %s → %s project (duration=%s, files=%d)",
		wf.Name, target, result.Duration, len(result.Files))

	return result, nil
}

// ============================================================================
// Event Publishing
// ============================================================================

func (c *compilerImpl) publishProgress(topic event.Topic, workflowID string, target workflow.Target, outputDir string, progress float64, message string) {
	if c.bus == nil {
		return
	}
	data := event.CompileEventData{
		WorkflowID: workflowID,
		Target:     string(target),
		OutputDir:  outputDir,
		Progress:   progress,
	}
	if topic == event.TopicCompileFailed {
		data.Error = message
	} else if topic == event.TopicCompileCompleted {
		data.Duration = message
	}
	c.bus.Publish(topic, data)
}

// ============================================================================
// Registry Methods
// ============================================================================

func (c *compilerImpl) ListAllTargets() []TargetInfo {
	return c.ListTargets()
}

func (c *compilerImpl) ListTargets() []TargetInfo {
	generators := c.registry.List()
	infos := make([]TargetInfo, 0, len(generators))
	for _, g := range generators {
		infos = append(infos, TargetInfo{
			Target:      g.ID(),
			Name:        g.Name(),
			Description: g.Description(),
			Version:     g.Version(),
		})
	}
	return infos
}

func (c *compilerImpl) GetGenerator(target workflow.Target) (Generator, bool) {
	return c.registry.Get(target)
}

func (c *compilerImpl) RegisterGenerator(g Generator) error {
	return c.registry.Register(g)
}

// ============================================================================
// Helpers
// ============================================================================

func sanitizeName(name string) string {
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		ch := name[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' {
			result = append(result, ch)
		} else if ch == ' ' {
			result = append(result, '_')
		}
	}
	if len(result) == 0 {
		return "project"
	}
	return string(result)
}

func (c *compilerImpl) availableTargets() string {
	infos := c.ListTargets()
	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = string(info.Target)
	}
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, ", ")
}
