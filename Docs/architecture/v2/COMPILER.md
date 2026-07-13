# Compiler Architecture

## Overview

The Compiler is the **heart of AIStudio V2**. It transforms a declarative `Workflow` (DSL) into a real, runnable engineering project on the filesystem. The Compiler does **not** execute projects, modify projects, install dependencies, or generate code from AI.

```
                  ┌──────────────────────────────────────┐
                  │            Compiler                   │
                  │                                       │
  workflow.json   │  ┌──────┐  ┌────────┐  ┌──────────┐  │   Generated
  ───────────────┼─►│ Plan │─►│Validate│─►│ Generate  │──┼──► Project
                  │  └──────┘  └────────┘  └──────────┘  │   Directory
                  │       │         │            │        │
                  │       ▼         ▼            ▼        │
                  │  ┌─────────────────────────────────┐  │
                  │  │       Generator Registry        │  │
                  │  │  Python │ MATLAB │ ROS2 │ ...   │  │
                  │  └─────────────────────────────────┘  │
                  │                                       │
                  │  Events: ───► EventBus (progress)     │
                  └──────────────────────────────────────┘
```

## Compilation Pipeline

The compilation pipeline consists of 4 stages:

### Stage 1: Plan (Preview)

Before any files are written, the Compiler generates a `CompilePlan` that describes what would be generated:

```
Plan(wf, opts) → CompilePlan{
    GeneratorID:     "python",
    GeneratorName:   "Python Generator",
    ProjectName:     "my-project",
    OutputDir:       "./output/my-project",
    EstimatedFiles:  12,
    EstimatedSizeKB: 170,
    Validated:       true,
    Warnings:        [],
    RuntimeReq:      {Name: "yolo", Python: ">=3.9", GPU: true},
}
```

Used for **dry-run mode** (`opts.DryRun = true`) and **preview in UI**.

### Stage 2: Validate (Workflow + Host)

Two levels of validation:

1. **Workflow Validation** (`generator.Validate(wf)`) — Checks if the workflow is semantically valid for the target (e.g., required nodes exist, port types match)
2. **Host Validation** (`generator.CompileTimeValidate(ctx)`) — Checks if the host system has required tools (e.g., `python3`, `matlab`, `g++`)

### Stage 3: Generate (Project)

The selected Generator creates the complete project:

- Creates directory structure
- Generates source files from templates
- Writes configuration files
- Returns `GenerateResult` with project root, entry points, and file list

### Stage 4: Verify (Output)

After generation, the Compiler collects:
- Runtime requirements from the generator
- File listing
- Duration statistics
- Any warnings

## Event-Based Progress Reporting

The Compiler publishes progress events via the EventBus throughout the pipeline:

| Topic | Progress | Description |
|-------|----------|-------------|
| `compile.started` | 0.0 | Compilation started |
| `compile.progress` | 0.1 | Planning |
| `compile.progress` | 0.2 | Validating host environment |
| `compile.progress` | 0.3 | Validating workflow |
| `compile.progress` | 0.5 | Generating project |
| `compile.progress` | 0.9 | Verifying output |
| `compile.completed` | 1.0 | Compilation completed |

On failure, `compile.failed` is published with the error message.

### Event Data

```go
type CompileEventData struct {
    WorkflowID string  `json:"workflowId"`
    Target     string  `json:"target"`
    OutputDir  string  `json:"outputDir,omitempty"`
    Progress   float64 `json:"progress,omitempty"`
    Error      string  `json:"error,omitempty"`
    Duration   string  `json:"duration,omitempty"`
}
```

## Compiler Interface

```go
type Compiler interface {
    // Compile compiles a workflow into a project directory.
    Compile(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*CompileResult, error)

    // Plan returns a compilation plan without writing any files.
    Plan(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*CompilePlan, error)

    // ListTargets returns all available compilation targets.
    ListTargets() []TargetInfo

    // GetGenerator returns the generator for a given target.
    GetGenerator(target workflow.Target) (Generator, bool)

    // RegisterGenerator registers a new generator.
    RegisterGenerator(g Generator) error
}
```

## CompileOptions

```go
type CompileOptions struct {
    OutputDir   string            // Output directory
    Target      workflow.Target   // Override target
    Variables   map[string]string // Template variable overrides
    Force       bool              // Overwrite existing output
    DryRun      bool              // Validate without writing files
    ProjectName string            // Override project name
}
```

## CompileResult

```go
type CompileResult struct {
    Target      workflow.Target     `json:"target"`
    ProjectRoot string              `json:"projectRoot"`
    EntryPoints []string            `json:"entryPoints"`
    Files       []GeneratedFile     `json:"files,omitempty"`
    RuntimeReq  *RuntimeRequirement `json:"runtimeRequirement,omitempty"`
    Duration    time.Duration       `json:"duration"`
    WorkflowID  string              `json:"workflowId"`
    GeneratorID string              `json:"generatorId"`
    Warnings    []string            `json:"warnings,omitempty"`
}
```

## Generator Interface

```go
type Generator interface {
    ID() workflow.Target
    Name() string
    Description() string
    Version() string

    Generate(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*GenerateResult, error)
    RuntimeRequirement(wf *workflow.Workflow) (*RuntimeRequirement, error)
    Validate(wf *workflow.Workflow) error
    EstimateResources(wf *workflow.Workflow) (*ResourceEstimate, error)
    CompileTimeValidate(ctx context.Context) error
}
```

## How to Create a New Generator

1. Create a new directory under `packages/generators/<name>/`
2. Implement the `Generator` interface (or embed `BaseGenerator` for defaults)
3. Register the generator in `Backend/cmd/main.go`:

```go
func registerGenerators(compilerEngine compiler.Compiler) {
    compilerEngine.RegisterGenerator(compilerPython.NewGenerator())
    // Add new generators here:
    // compilerEngine.RegisterGenerator(mygenerator.NewGenerator())
}
```

See [GENERATORS.md](GENERATORS.md) for detailed guidance.

## Generator Registry

The `GeneratorRegistry` manages all registered generators:

```go
type GeneratorRegistry struct {
    generators map[workflow.Target]Generator
}

// Methods: Register, MustRegister, Get, List, Unregister, Count, HasTarget
```

## Dry-Run Mode

When `opts.DryRun = true`, the Compiler calls `Plan()` only — no files are written. The UI uses this for:
- Previewing what will be generated
- Showing estimated file count and size
- Validating the workflow before committing

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/compiler/targets` | List all available targets |
| POST | `/api/v1/compiler/compile` | Compile a workflow (full) |
| POST | `/api/v1/compiler/validate` | Validate a workflow |
