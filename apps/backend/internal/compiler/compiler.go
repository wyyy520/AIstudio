// Package compiler transforms Workflow DSL into real engineering projects.
//
// This is the integration layer that wraps packages/compiler for the backend.
// It handles file I/O (reading workflow.json) and bridges the internal
// generator registry with the packages/compiler pipeline.
//
// Architecture:
//
//	Workflow file → internal/compiler (IO) → packages/compiler (pipeline) → Generator → Project
//
// The internal compiler delegates ALL pipeline logic (graph optimization,
// EWIR, domain dispatch, plugin manifest) to packages/compiler.
package compiler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aistudio/backend/internal/eventbus"
	"github.com/aistudio/backend/internal/workflow"
	pkgcompiler "github.com/aistudio/packages/compiler"
	pkgworkflow "github.com/aistudio/packages/workflow"
)

// ============================================================================
// Compiler Interface (backward compatible)
// ============================================================================

// Compiler compiles a Workflow into a project.
type Compiler interface {
	Compile(ctx context.Context, workflowPath string, opts CompileOptions) (*CompileResult, error)
	Plan(ctx context.Context, workflowPath string, opts CompileOptions) (*CompilePlan, error)
	ListTargets() []TargetInfo
	ListAllTargets() []TargetInfo
	GetGenerator(target workflow.Target) (Generator, bool)
	RegisterGenerator(g Generator) error
}

// ============================================================================
// Types (re-exported from packages/compiler with backward compatibility)
// ============================================================================

type CompileOptions = pkgcompiler.CompileOptions
type CompileResult = pkgcompiler.CompileResult
type CompilePlan = pkgcompiler.CompilePlan
type GeneratedFile = pkgcompiler.GeneratedFile
type TargetInfo = pkgcompiler.TargetInfo
type RuntimeRequirement = pkgcompiler.RuntimeRequirement
type ResourceEstimate = pkgcompiler.ResourceEstimate
type GenerateResult = pkgcompiler.GenerateResult

// ============================================================================
// Generator Interface (internal, takes ExecutionPlan)
// ============================================================================

// Generator generates a complete project from an ExecutionPlan.
// This is the internal generator interface used by adapters.
type Generator interface {
	ID() workflow.Target
	Name() string
	Description() string
	Version() string
	Generate(ctx context.Context, plan *ExecutionPlan, opts CompileOptions) (*GenerateResult, error)
	RuntimeRequirement(wf *workflow.Workflow) (*RuntimeRequirement, error)
	Validate(wf *workflow.Workflow) error
	EstimateResources(wf *workflow.Workflow) (*ResourceEstimate, error)
	CompileTimeValidate(ctx context.Context) error
}

// ============================================================================
// Constructor
// ============================================================================

// NewCompiler creates a new Compiler that delegates to packages/compiler.
func NewCompiler(bus *eventbus.EventBus) Compiler {
	// internal/eventbus is a thin re-export of packages/event.
	// We pass nil to packages/compiler since we handle all events at this layer.
	pkgComp := pkgcompiler.NewCompiler(nil)

	return &compilerImpl{
		registry: NewRegistry(),
		bus:      bus,
		engine:   pkgComp,
	}
}

// ============================================================================
// Implementation
// ============================================================================

type compilerImpl struct {
	registry *GeneratorRegistry
	bus      *eventbus.EventBus
	engine   pkgcompiler.Compiler
}

// readWorkflow reads and parses a workflow.json file.
func (c *compilerImpl) readWorkflow(workflowPath string) (*workflow.Workflow, error) {
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, fmt.Errorf("read workflow.json %s: %w", workflowPath, err)
	}
	var wf workflow.Workflow
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("unmarshal workflow.json %s: %w", workflowPath, err)
	}
	return &wf, nil
}

// toPkgWorkflow converts internal workflow to packages workflow.
// Since internal/workflow uses type aliases (type Workflow = pw.Workflow),
// this is a zero-cost identity conversion.
func toPkgWorkflow(wf *workflow.Workflow) *pkgworkflow.Workflow {
	if wf == nil {
		return nil
	}
	return (*pkgworkflow.Workflow)(wf)
}

// Plan returns a compilation plan without writing files.
func (c *compilerImpl) Plan(ctx context.Context, workflowPath string, opts CompileOptions) (*CompilePlan, error) {
	start := time.Now()

	wf, err := c.readWorkflow(workflowPath)
	if err != nil {
		return nil, err
	}

	target := wf.Target
	if opts.Target != "" {
		target = opts.Target
	}

	// Use internal registry for generator lookup
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

	log.Printf("[compiler] plan: %s → %s (estimated %d files, %d KB, duration=%v)",
		wf.Name, target, estFiles, estSizeKB, time.Since(start))

	return &CompilePlan{
		GeneratorID:     pkgworkflow.Target(target),
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

// Compile compiles a workflow from its file path into a project directory.
// Uses the internal generator registry for adapter-based generation.
func (c *compilerImpl) Compile(ctx context.Context, workflowPath string, opts CompileOptions) (*CompileResult, error) {
	start := time.Now()

	// 1. Read workflow file
	wf, err := c.readWorkflow(workflowPath)
	if err != nil {
		return nil, err
	}

	target := wf.Target
	if opts.Target != "" {
		target = opts.Target
	}

	pkgTarget := pkgworkflow.Target(target)

	// 2. Find generator from internal registry
	generator, ok := c.registry.Get(target)
	if !ok {
		err := fmt.Errorf("no generator registered for target %q (available: %s)", target, c.availableTargets())
		c.publishProgress(eventbus.TopicCompileFailed, wf.ID, target, "", 0.0, err.Error())
		return nil, err
	}

	// ================================================================
	// Stage 1: Started
	// ================================================================
	c.publishProgress(eventbus.TopicCompileStarted, wf.ID, target, "", 0.0, "")

	// Stage 2: Graph Optimization
	c.publishProgress(eventbus.TopicCompileProgress, wf.ID, target, "", 0.05, "optimizing graph")

	pkgOptimizer := pkgcompiler.NewGraphOptimizer()
	// Type conversion: internal/workflow types are aliases to packages/workflow
	// so this works transparently
	optResult := pkgOptimizer.Optimize(wf.Nodes, wf.Edges)
	if len(optResult.Warnings) > 0 {
		for _, w := range optResult.Warnings {
			log.Printf("[compiler] optimizer: %s", w)
		}
	}

	// Check for cycles
	hasCycle, cycleNodes := pkgOptimizer.DetectCycles(optResult.Nodes, optResult.Edges)
	if hasCycle {
		err := fmt.Errorf("workflow contains cycle involving nodes: %v", cycleNodes)
		c.publishProgress(eventbus.TopicCompileFailed, wf.ID, target, "", 0.0, err.Error())
		return nil, err
	}

	// Update workflow with optimized nodes/edges
	wf.Nodes = optResult.Nodes
	wf.Edges = optResult.Edges

	if optResult.OriginalCount != optResult.OptimizedCount {
		log.Printf("[compiler] graph optimized: %d → %d nodes", optResult.OriginalCount, optResult.OptimizedCount)
	}

	// Stage 3: Validate workflow
	c.publishProgress(eventbus.TopicCompileProgress, wf.ID, target, "", 0.1, "validating")
	if err := generator.Validate(wf); err != nil {
		c.publishProgress(eventbus.TopicCompileFailed, wf.ID, target, "", 0.0, err.Error())
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Stage 4: Host validation
	c.publishProgress(eventbus.TopicCompileProgress, wf.ID, target, "", 0.2, "validating host environment")
	if err := generator.CompileTimeValidate(ctx); err != nil {
		c.publishProgress(eventbus.TopicCompileFailed, wf.ID, target, "", 0.0, err.Error())
		return nil, fmt.Errorf("host validation failed: %w", err)
	}

	// Stage 4: Build ExecutionPlan (delegate to packages/compiler)
	c.publishProgress(eventbus.TopicCompileProgress, wf.ID, target, "", 0.3, "building execution plan")
	execPlan, planErr := BuildExecutionPlan(wf, opts)
	if planErr != nil {
		c.publishProgress(eventbus.TopicCompileFailed, wf.ID, target, "", 0.0, planErr.Error())
		return nil, fmt.Errorf("execution plan build failed: %w", planErr)
	}

	// Attach runtime requirement
	if rr, rrErr := generator.RuntimeRequirement(wf); rrErr == nil {
		execPlan.RuntimeReq = rr
	}

	// Stage 5: EWIR
	c.publishProgress(eventbus.TopicCompileProgress, wf.ID, target, "", 0.5, "building EWIR")
	ewirOutputDir := opts.OutputDir
	if ewirOutputDir == "" {
		ewirOutputDir = filepath.Join(filepath.Dir(workflowPath), ".engstudio")
	}
	if err := c.buildEWIR(wf, ewirOutputDir); err != nil {
		log.Printf("[compiler] warning: EWIR build failed: %v", err)
	}

	// Stage 6: Domain dispatch
	c.publishProgress(eventbus.TopicCompileProgress, wf.ID, target, "", 0.6, "dispatching domains")
	domains := getDomains(execPlan)
	if len(domains) > 1 {
		log.Printf("[compiler] multi-domain workflow: %d domains: %v", len(domains), domains)
	}

	// Stage 7: Plugin manifest
	c.publishProgress(eventbus.TopicCompileProgress, wf.ID, target, "", 0.7, "generating plugin manifest")
	if err := c.buildPluginManifest(wf, ewirOutputDir); err != nil {
		log.Printf("[compiler] warning: plugin manifest failed: %v", err)
	}

	// ================================================================
	// Stage 8: Generate via internal generator adapter
	// ================================================================
	c.publishProgress(eventbus.TopicCompileProgress, wf.ID, target, "", 0.8, "generating project")

	genResult, genErr := generator.Generate(ctx, execPlan, opts)
	if genErr != nil {
		c.publishProgress(eventbus.TopicCompileFailed, wf.ID, target, "", 0.0, genErr.Error())
		return nil, fmt.Errorf("generation failed: %w", genErr)
	}

	// Stage 9: Verify
	c.publishProgress(eventbus.TopicCompileProgress, wf.ID, target, "", 0.95, "verifying output")

	var runtimeReq *RuntimeRequirement
	if rr, err := generator.RuntimeRequirement(wf); err == nil {
		runtimeReq = rr
	}

	// Stage 10: Complete
	result := &CompileResult{
		Target:      pkgTarget,
		ProjectRoot: genResult.ProjectRoot,
		EntryPoints: genResult.EntryPoints,
		Files:       genResult.Files,
		RuntimeReq:  runtimeReq,
		Duration:    time.Since(start),
		WorkflowID:  wf.ID,
		GeneratorID: string(generator.ID()),
	}

	c.publishProgress(eventbus.TopicCompileCompleted, wf.ID, target, result.ProjectRoot, 1.0, result.Duration.String())

	log.Printf("[compiler] compiled workflow %s → %s project (duration=%s, files=%d)",
		wf.Name, target, result.Duration, len(result.Files))

	return result, nil
}

// buildEWIR delegates EWIR generation to packages/compiler.
func (c *compilerImpl) buildEWIR(wf *workflow.Workflow, outputDir string) error {
	pkgWf := toPkgWorkflow(wf)
	builder := pkgcompiler.NewEWIRBuilder("1.0.0")
	_, err := builder.Split(pkgWf, outputDir)
	return err
}

// buildPluginManifest delegates plugin manifest generation to packages/compiler.
func (c *compilerImpl) buildPluginManifest(wf *workflow.Workflow, outputDir string) error {
	pkgWf := toPkgWorkflow(wf)
	gen := pkgcompiler.NewPluginManifestGenerator()
	_, err := gen.GenerateAndWrite(pkgWf, outputDir)
	return err
}

// getDomains extracts distinct domains from an ExecutionPlan.
func getDomains(plan *ExecutionPlan) []workflow.Target {
	seen := make(map[workflow.Target]bool)
	var domains []workflow.Target
	for _, np := range plan.NodePlans {
		domain := workflow.Target(np.Domain)
		if domain == "" {
			continue
		}
		if !seen[domain] {
			seen[domain] = true
			domains = append(domains, domain)
		}
	}
	return domains
}

// ============================================================================
// Event Publishing
// ============================================================================

func (c *compilerImpl) publishProgress(topic eventbus.Topic, workflowID string, target workflow.Target, outputDir string, progress float64, message string) {
	if c.bus == nil {
		return
	}
	data := eventbus.CompileEventData{
		WorkflowID: workflowID,
		Target:     string(target),
		OutputDir:  outputDir,
		Progress:   progress,
	}
	if topic == eventbus.TopicCompileFailed {
		data.Error = message
	} else if topic == eventbus.TopicCompileCompleted {
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
			Target:      pkgworkflow.Target(g.ID()),
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

// ============================================================================
// Helpers
// ============================================================================

func sanitizeName(name string) string {
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			result = append(result, c)
		} else if c == ' ' {
			result = append(result, '_')
		}
	}
	if len(result) == 0 {
		return "project"
	}
	return string(result)
}

// ============================================================================
// Compile-time EventBus type check
// ============================================================================

// EventBus is the same as packages/event.EventBus (via type alias)
type EventBus = eventbus.EventBus
