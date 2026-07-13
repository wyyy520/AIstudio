package compiler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aistudio/packages/event"
	"github.com/aistudio/packages/workflow"
)

const (
	BufferSize    = 4096
	numWorkers    = 4
	genCacheSize  = 16
)

var (
	compileResultPool = sync.Pool{
		New: func() any {
			return &CompileResult{
				EntryPoints: make([]string, 0, 4),
			}
		},
	}
	compilePlanPool = sync.Pool{
		New: func() any {
			return &CompilePlan{
				Warnings: make([]string, 0),
			}
		},
	}
)

// Compiler compiles a Workflow into a project.
// It is the single entry point for all project generation.
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
}

// NewCompiler creates a new Compiler with the given event bus.
func NewCompiler(bus *event.EventBus) Compiler {
	return &compilerImpl{
		registry: NewGeneratorRegistry(),
		bus:      bus,
	}
}

type compilerImpl struct {
	registry *GeneratorRegistry
	bus      *event.EventBus
}

func (c *compilerImpl) Plan(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*CompilePlan, error) {
	target := wf.Target
	if opts.Target != "" {
		target = opts.Target
	}

	generator, ok := c.registry.Get(target)
	if !ok {
		return nil, fmt.Errorf("no generator registered for target %q", target)
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

	plan := compilePlanPool.Get().(*CompilePlan)
	plan.GeneratorID = target
	plan.GeneratorName = generator.Name()
	plan.ProjectName = projectName
	plan.OutputDir = opts.OutputDir
	plan.EstimatedFiles = estFiles
	plan.EstimatedSizeKB = estSizeKB
	plan.Validated = len(warnings) == 0
	plan.Warnings = warnings
	plan.RuntimeReq = rr
	return plan, nil
}

func (c *compilerImpl) Compile(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*CompileResult, error) {
	start := time.Now()

	target := wf.Target
	if opts.Target != "" {
		target = opts.Target
	}

	c.publishProgress(event.TopicCompileStarted, wf.ID, target, "", 0.0, "")

	generator, ok := c.registry.Get(target)
	if !ok {
		err := fmt.Errorf("no generator registered for target %q", target)
		c.publishProgress(event.TopicCompileFailed, wf.ID, target, "", 0.0, err.Error())
		return nil, err
	}

	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.1, "planning")

	plan, planErr := c.Plan(ctx, wf, opts)
	if planErr != nil {
		c.publishProgress(event.TopicCompileFailed, wf.ID, target, "", 0.0, planErr.Error())
		return nil, fmt.Errorf("compilation planning failed: %w", planErr)
	}

	if len(plan.Warnings) > 0 {
		for _, w := range plan.Warnings {
			_ = w
		}
	}

	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.2, "validating host environment")

	if err := generator.CompileTimeValidate(ctx); err != nil {
		c.publishProgress(event.TopicCompileFailed, wf.ID, target, "", 0.0, err.Error())
		return nil, fmt.Errorf("host validation failed: %w", err)
	}

	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.3, "validating workflow")

	if err := generator.Validate(wf); err != nil {
		c.publishProgress(event.TopicCompileFailed, wf.ID, target, "", 0.0, err.Error())
		return nil, fmt.Errorf("workflow validation failed for target %s: %w", target, err)
	}

	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.5, "generating project")

	genResult, err := generator.Generate(ctx, wf, opts)
	if err != nil {
		c.publishProgress(event.TopicCompileFailed, wf.ID, target, "", 0.0, err.Error())
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	c.publishProgress(event.TopicCompileProgress, wf.ID, target, "", 0.9, "verifying output")

	var runtimeReq *RuntimeRequirement
	if rr, err := generator.RuntimeRequirement(wf); err == nil {
		runtimeReq = rr
	}

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

	return result, nil
}

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

func availableTargets(infos []TargetInfo) string {
	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = string(info.Target)
	}
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, ", ")
}