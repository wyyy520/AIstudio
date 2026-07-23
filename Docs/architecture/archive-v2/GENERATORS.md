# Generator Guide

## Overview

Generators are the **target-specific project creators** in AIStudio V2. Each Generator knows how to produce a complete, standards-based project for one target platform. Generators implement the `Generator` interface defined in `packages/generators/common/generator.go`.

```
Generator Interface
├── Python Generator  ───► Python project (pip installable)
├── MATLAB Generator  ───► MATLAB project (.m files, toolboxes)
├── ROS2 Generator    ───► ROS2 workspace (ament colcon)
├── STM32 Generator   ───► STM32 embedded project (CubeIDE)
├── Docker Generator  ───► Docker image (Dockerfile)
├── C++ Generator     ───► C++ project (CMakeLists.txt)
├── Unity Generator   ───► Unity project (C# scripts)
└── Java Generator    ───► Java project (Maven/Gradle)
```

## All 8 Generators

### 1. Python Generator

| Attribute | Value |
|-----------|-------|
| Target | `python` |
| Location | `packages/generators/python/` |
| Internal | `Backend/internal/compiler/generators/python/` |

Produces a standard Python project with:
- `requirements.txt` — Dependencies from workflow metadata
- `src/` — Package with `__init__.py` and generated modules
- `train.py` / `inference.py` — Entry points based on workflow nodes
- `config.yaml` — Configuration from workflow variables
- `tests/` — Test stubs

**Example output:**
```
my-project/
├── workflow.json
├── requirements.txt      # torch, ultralytics, opencv-python
├── src/
│   ├── __init__.py
│   ├── dataset.py        # Data loading logic
│   ├── model.py          # Model definition
│   ├── train.py          # Training loop
│   └── inference.py      # Inference entry point
├── config/
│   └── config.yaml       # Hyperparameters
├── data/                 # Dataset reference
├── models/               # Model output
├── tests/
│   └── test_model.py
└── outputs/              # Training logs
```

### 2. MATLAB Generator

| Attribute | Value |
|-----------|-------|
| Target | `matlab` |
| Location | `packages/generators/matlab/` |

Produces a MATLAB project with:
- `*.m` — MATLAB scripts and functions
- `*.mlx` — Live scripts
- Startup/teardown scripts
- Model export (`.mat` files)

### 3. ROS2 Generator

| Attribute | Value |
|-----------|-------|
| Target | `ros2` |
| Location | `packages/generators/ros2/` |

Produces a ROS2 workspace with:
- `package.xml` — Package manifest
- `setup.py` — ROS2 package setup
- Nodes, topics, services from workflow
- Launch files
- `colcon build` ready

### 4. STM32 Generator

| Attribute | Value |
|-----------|-------|
| Target | `stm32` |
| Location | `packages/generators/stm32/` |

Produces an STM32 embedded project with:
- CubeIDE `.project` and `.cproject`
- HAL drivers
- `main.c` with generated initialization
- `Makefile` for command-line builds
- Linker script `.ld`

### 5. Docker Generator

| Attribute | Value |
|-----------|-------|
| Target | `docker` |
| Location | `packages/generators/docker/` |

Produces a Docker image definition with:
- `Dockerfile` — Multi-stage build
- `.dockerignore`
- `docker-compose.yml` — For multi-container setups
- Entry point scripts

### 6. C++ Generator

| Attribute | Value |
|-----------|-------|
| Target | `cpp` |
| Location | `packages/generators/cpp/` |

Produces a C++ project with:
- `CMakeLists.txt` — Build configuration
- `src/main.cpp` — Entry point
- `include/` — Header files
- `lib/` — Third-party library references

### 7. Unity Generator

| Attribute | Value |
|-----------|-------|
| Target | `unity` |
| Location | `packages/generators/unity/` |

Produces a Unity project with:
- C# scripts (`.cs`)
- Unity scene files (`.unity`)
- Package manifest
- Assembly definitions

### 8. Java Generator

| Attribute | Value |
|-----------|-------|
| Target | `java` |
| Location | `packages/generators/java/` |

Produces a Java project with:
- Maven `pom.xml` or Gradle `build.gradle`
- `src/main/java/` — Source packages
- `src/test/java/` — Test classes
- Application entry point

---

## Generator Interface (Public)

Defined in `packages/generators/common/generator.go`:

```go
type Generator interface {
    ID() Target
    Name() string
    Description() string
    Version() string

    Generate(ctx context.Context, wf *Workflow, opts CompileOptions) (*GenerateResult, error)
    RuntimeRequirement(wf *Workflow) (*RuntimeRequirement, error)
    Validate(wf *Workflow) error
    EstimateResources(wf *Workflow) (*ResourceEstimate, error)
    CompileTimeValidate(ctx context.Context) error
    Plan(ctx context.Context, wf *Workflow, opts CompileOptions) (*CompilePlan, error)
}
```

### BaseGenerator

Embed `BaseGenerator` for default implementations:

```go
type MyGenerator struct {
    common.BaseGenerator
}

func NewGenerator() *MyGenerator {
    return &MyGenerator{
        BaseGenerator: common.BaseGenerator{
            TargetID:      "my_target",
            GeneratorName: "My Generator",
            GeneratorDesc: "Generates my_target projects",
            GeneratorVer:  "1.0.0",
        },
    }
}

func (g *MyGenerator) Generate(ctx context.Context, wf *common.Workflow, opts common.CompileOptions) (*common.GenerateResult, error) {
    // ... generate project files
}
```

---

## Template System

Generators use Go's `text/template` with `embed.FS` for templates:

```go
//go:embed templates/*
var templateFS embed.FS

func (g *PythonGenerator) Generate(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*GenerateResult, error) {
    // Parse and execute templates
    tmpl := template.Must(template.ParseFS(templateFS, "templates/**/*.tmpl"))
    // Execute with workflow data
    // Write output files
}
```

Template directory structure:
```
generators/python/
├── generator.go
└── templates/
    ├── requirements.txt.tmpl
    ├── src/
    │   ├── __init__.py.tmpl
    │   ├── train.py.tmpl
    │   └── inference.py.tmpl
    ├── config/
    │   └── config.yaml.tmpl
    └── tests/
        └── test_model.py.tmpl
```

---

## How to Create a New Generator

### Step 1: Create the package

```
packages/generators/<name>/
├── generator.go       ← Generator implementation
└── templates/         ← Go embed templates
    └── ...
```

### Step 2: Implement Generator interface

```go
package mygenerator

import (
    "context"
    "github.com/aistudio/backend/packages/generators/common"
)

type MyGenerator struct {
    common.BaseGenerator
}

func NewGenerator() *MyGenerator {
    return &MyGenerator{
        BaseGenerator: common.BaseGenerator{
            TargetID:      "my_target",
            GeneratorName: "My Generator",
            GeneratorDesc: "Generates my_target projects",
            GeneratorVer:  "1.0.0",
        },
    }
}

func (g *MyGenerator) Generate(ctx context.Context, wf *common.Workflow, opts common.CompileOptions) (*common.GenerateResult, error) {
    // 1. Create project directory
    // 2. Process workflow nodes → source code
    // 3. Write files from templates
    // 4. Return GenerateResult with project root and entry points
}
```

### Step 3: Register in main.go

```go
func registerGenerators(compilerEngine compiler.Compiler) {
    compilerEngine.RegisterGenerator(compilerPython.NewGenerator())
    compilerEngine.RegisterGenerator(mygenerator.NewGenerator())
}
```

---

## Generated Project Structure (Generic)

Every generated project follows this layout:

```
<project-name>/
├── workflow.json           ← SSoT (copied into project)
├── src/                    ← Source code (target-specific)
├── config/                 ← Configuration files
├── data/                   ← Dataset references / directories
├── models/                 ← Model output directories
├── outputs/                ← Execution outputs, logs
├── tests/                  ← Test suite
├── .gitignore              ← Standard ignores
└── README.md               ← Generated documentation
```

---

## Generator Adapter Pattern

The internal compiler also defines a `Generator` interface (in `Backend/internal/compiler/generators.go`) that mirrors the public interface. An **adapter** bridges between the two:

```go
// python_adapter.go
type PythonAdapter struct {
    inner *python.Generator  // From packages/generators/python
}

func (a *PythonAdapter) Generate(ctx context.Context, wf *workflow.Workflow, opts compiler.CompileOptions) (*compiler.GenerateResult, error) {
    result, err := a.inner.Generate(ctx, wf, toCommonOpts(opts))
    return fromCommonResult(result), err
}
```

---

## Estimating Resources

Each Generator provides resource estimates for the Plan step:

```go
func (g *MyGenerator) EstimateResources(wf *common.Workflow) (*common.ResourceEstimate, error) {
    return &common.ResourceEstimate{
        EstimatedFiles:  len(wf.Nodes) + 8,
        EstimatedSizeKB: len(wf.Nodes) * 10 + 50,
        RequiresGPU:     false,
        MinMemoryMB:     512,
        MinDiskMB:       100,
    }, nil
}
```

---

## Runtime Requirements

Each Generator declares what runtime environment the generated project needs:

```go
func (g *PythonGenerator) RuntimeRequirement(wf *common.Workflow) (*common.RuntimeRequirement, error) {
    return &common.RuntimeRequirement{
        Name:     "python-base",
        Version:  "1.0.0",
        Python:   ">=3.9",
        Packages: []string{"torch>=2.0", "ultralytics"},
        Commands: []string{"python3"},
        GPU:      true,
        MinMemoryMB: 4096,
        MinDiskMB:   2048,
    }, nil
}
```
