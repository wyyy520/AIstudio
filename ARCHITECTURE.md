# AIStudio Architecture (Unified)

## Architecture Pipeline

```
Project
  │
  ▼
Workflow (Single Source of Truth)
  │
  ▼
Compiler
  │
  ▼
Template Engine
  │
  ▼
Generator (8 languages: Python, MATLAB, C++, Java, ROS2, Unity, Docker, STM32)
  │
  ▼
Runtime (Local, Docker, SSH)
  │
  ├──► Terminal
  │
  ├──► Log
  │
  ▼
Diagnose
  │
  ▼
Skill (AI Optional)
```

## Layer 0: Project
- Directory that contains the workflow.json + generated code
- Each project is a file-system directory

## Layer 1: Workflow — Single Source of Truth
- workflow.json is the ONLY source of all configuration
- No other state is stored for workflow configuration
- Workflow DSL defines: nodes, edges, parameters, targets, metadata

## Layer 2: Compiler
- Input: workflow.json
- Output: CompilePlan (ordered list of compilation steps)
- Validates workflow schema
- Resolves dependencies
- Produces compilation plan

## Layer 3: Template Engine
- Input: CompilePlan
- Output: Rendered templates with workflow data
- Supported templates: .tmpl files per language
- Template directories:
  - templates/python/
  - templates/matlab/
  - templates/stm32/
  - templates/cpp/
  - templates/java/
  - templates/ros2/
  - templates/unity/
  - templates/docker/

## Layer 4: Generator
- Input: Rendered templates
- Output: Complete project files
- Each language has a self-contained generator package
- Generators NEVER concatenate strings directly - always use templates

## Layer 5: Runtime
- Input: Generated project
- Execution modes: Local, Docker, SSH
- Output: RunResult (status, logs, artifacts)

## Layer 6: Log + Terminal
- Runtime → Log (structured storage)
- Runtime → Terminal (real-time streaming)
- Log Center provides unified query interface

## Layer 7: Diagnose
- Analyzes runtime logs
- Detects errors and suggests fixes
- Works with or without LLM

## Layer 8: Skill (AI Optional)
- LLM-based: Planner, Explain, Optimize
- System works fully WITHOUT LLM
- LLM is purely a skill-layer enhancement

## Project Structure

```
aistudio/
├── apps/
│   ├── backend/       # Go API server (orchestrates the pipeline)
│   └── desktop/       # Vue 3 + Tauri desktop app
├── packages/          # Shared Go libraries (canonical implementations)
│   ├── workflow/      # Workflow DSL, schema, validation
│   ├── compiler/      # Compiler framework
│   ├── generators/    # All language generators (delegates to templates)
│   ├── runtime/       # Runtime execution
│   ├── project/       # Project management
│   ├── plugin/        # Plugin system
│   ├── environment/   # Environment detection
│   ├── diagnostic/    # Error analysis
│   ├── agent/         # AI agent framework
│   ├── event/         # Event bus
│   ├── common/        # Shared utilities
│   ├── logger/        # Logging framework
│   ├── security/      # Auth, JWT, encryption
│   ├── storage/       # Database layer
│   ├── skill/         # Skill management
│   └── sdk/           # Public SDK
├── templates/         # Centralized template directory
│   ├── python/
│   ├── matlab/
│   ├── stm32/
│   ├── cpp/
│   ├── java/
│   ├── ros2/
│   ├── unity/
│   └── docker/
├── Engine/            # Python AI engine (for inference/training)
├── Plugins/           # Plugin definitions (JSON manifests)
└── Launcher/          # Service process manager
```

## Naming Conventions

### Go
- Manager: `WorkflowManager`, `ProjectManager`, `RuntimeManager`, `PluginManager`, `EnvironmentManager`
- NO: `ManagerService`, `ServiceManager`, `NewManager`, `ManagerV2`
- Package names: lowercase, single word
- NO: `eventbus` (use `event`), `logcenter` (use `logger`)

### TypeScript/Vue
- Components: PascalCase, directory-based (`ComponentName/ComponentName.vue`)
- Stores: camelCase files (`workflow.ts`, `project.ts`)
- API: camelCase files (`workflow.ts`, `project.ts`)
- Types: PascalCase per page (`pages/Workflow/types/workflow.ts`)

## Rule: No Duplicate Implementations
- Each domain has exactly ONE implementation in `packages/`
- `apps/backend/internal/` imports from `packages/`
- Migrations from `internal/` to `packages/` must be type-safe
- Generator adapters are the one bridge that converts between type systems

## Rule: Workflow JSON is the Single Source of Truth
- All workflow state is in `workflow.json`
- No Vue store holds workflow state independently
- No Go struct holds workflow state independently
- The file system is the source of truth

## Rule: All Generators Use Templates
- No string concatenation in any generator
- All code generation uses `.tmpl` files
- Template engine processes: Template Data + Workflow JSON → Rendered Output
