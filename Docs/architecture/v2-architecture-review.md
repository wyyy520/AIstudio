# AIStudio V2 Architecture Review

> **Author:** Chief Software Architect
> **Date:** 2026-07-12
> **Status:** Approved
> **Version:** 2.0.0

---

## Table of Contents

1. [Current Architecture Analysis](#1-current-architecture-analysis)
2. [Identified Problems](#2-identified-problems)
3. [Refactoring Goals](#3-refactoring-goals)
4. [New Architecture Design](#4-new-architecture-design)
5. [Module Division](#5-module-division)
6. [Monorepo Directory Design](#6-monorepo-directory-design)
7. [Compiler Architecture](#7-compiler-architecture)
8. [Workflow DSL](#8-workflow-dsl)
9. [Runtime Lifecycle](#9-runtime-lifecycle)
10. [Generator Interface](#10-generator-interface)
11. [Environment Manager](#11-environment-manager)
12. [Logger](#12-logger)
13. [Diagnostic](#13-diagnostic)
14. [Plugin](#14-plugin)
15. [Cloud](#15-cloud)
16. [Security](#16-security)
17. [Project Manager](#17-project-manager)
18. [API Gateway](#18-api-gateway)
19. [Event Bus](#19-event-bus)
20. [Sprint Roadmap](#20-sprint-roadmap)

---

## 1. Current Architecture Analysis

### 1.1 V1 System Overview

The current AIStudio V1 codebase is organized as follows:

```
AIStudio/
├── Backend/           # Go backend (Gin framework)
│   ├── cmd/main.go    # Entry point
│   └── internal/
│       ├── api/       # HTTP handlers + middleware
│       ├── workflow/  # Workflow engine (DAG + executor)
│       ├── task/      # Task scheduler
│       ├── agent/     # AI Agent (LLM + planner)
│       ├── engine/    # Python bridge (gRPC/subprocess)
│       ├── plugin/    # Plugin system
│       ├── environment/ # Environment detection
│       ├── mcp/       # MCP protocol
│       ├── service/   # Business logic layer
│       ├── auth/      # Authentication
│       ├── config/    # Configuration
│       ├── database/  # Database + models
│       └── launcher/  # Module lifecycle
├── Frontend/          # Vue3 + TypeScript + Tauri
│   ├── src/
│   │   ├── pages/     # Page components (unused by router)
│   │   ├── views/     # View components (used by router)
│   │   ├── stores/    # New store system
│   │   ├── store/     # Old store system
│   │   ├── components/ # Shared components
│   │   └── api/       # API client layer
├── Engine/            # Python AI Engine
│   ├── runtime/       # Python runtime
│   ├── trainer/       # Training modules
│   ├── inference/     # Inference modules
│   ├── model/         # Model management
│   ├── vision/        # Vision-specific (YOLO)
│   └── sdk/           # Python SDK
├── Config/            # Configuration files
├── Docs/              # Documentation
├── Plugins/           # Plugin directory
├── Runtime/           # Runtime cache
├── Storage/           # Persistent storage
└── Scripts/           # Build scripts
```

### 1.2 V1 Data Flow

```
User Input → Frontend → HTTP API → Service Layer
    → Workflow Engine (parse + execute DAG)
        → Task Scheduler (queue + worker pool)
            → Plugin Executor → Python Engine (subprocess)
                → Result → Frontend
```

### 1.3 V1 Component Dependencies

```
api/ → service/ → workflow/ → task/ → plugin/ → engine/
  │                    │          │
  │                    └──────────┴──→ environment/
  │
  ├──→ agent/ → mcp/
  ├──→ auth/
  └──→ config/ (global)
```

---

## 2. Identified Problems

### 2.1 Critical Issues

| # | Problem | Severity | Location | Impact |
|---|---------|----------|----------|--------|
| P1 | **No Compiler layer** | 🔴 Critical | Missing | Workflow directly drives execution, no project generation |
| P2 | **Workflow mixes declaration + runtime state** | 🔴 Critical | `workflow/types.go` | `Node.Runtime` in `workflow.json` violates separation of concerns |
| P3 | **Agent generates execution logic, not workflow** | 🔴 Critical | `agent/workflow_generator.go` | Agent generates YOLO training params directly, should generate declarative workflow |
| P4 | **Engine contains domain logic** | 🔴 Critical | `Engine/` | Python engine has YOLO training/inference hardcoded |
| P5 | **No Skill concept** | 🟡 High | Missing | No workflow template system |
| P6 | **Runtime is not unified** | 🟡 High | `engine/runner.go` | Python-specific runner, not a generic runtime |
| P7 | **No Runtime Bundle** | 🟡 High | Missing | No versioned, cached runtime environment |
| P8 | **Project is database-only** | 🟡 High | `service/project_service.go` | No real filesystem project structure |
| P9 | **Frontend dual routing** | 🟡 Medium | `pages/` vs `views/` | Two parallel UI systems |
| P10 | **Config system dual track** | 🟡 Medium | `Config/` vs `Backend/config/` | Two config systems |
| P11 | **No Event Bus** | 🟡 Medium | Missing | Module communication is hardcoded |
| P12 | **No AI Diagnostic** | 🟡 Medium | Missing | Error analysis is basic |
| P13 | **Plugin security missing** | 🟡 Medium | `plugin/` | No permission/signature/sandbox |
| P14 | **Logger scattered** | 🟢 Low | Multiple locations | Logging is fragmented across modules |
| P15 | **No Reverse Parser (correctly absent)** | 🟢 Low | N/A | Correct decision to not implement |

### 2.2 Root Cause Analysis

The core issue is that V1 was built as a **Workflow Execution Engine** rather than a **Workflow Compilation Platform**. The architecture conflates:

- **What to do** (Workflow declaration) with **How to do it** (execution logic)
- **Project generation** with **Task execution**
- **Environment management** with **Plugin execution**

### 2.3 Architectural Debt

1. **God Module**: `workflow/engine.go` handles parsing, validation, execution, and result collection
2. **Circular Dependency Risk**: `plugin` → `engine` → `task` → `workflow` → `plugin`
3. **Hardcoded Pathways**: `main.go` wires everything manually with `launcher`
4. **Missing Interfaces**: No clear `Compiler`, `Generator`, `Runtime` interfaces

---

## 3. Refactoring Goals

### 3.1 Core Principles

1. **Workflow First** — Workflow is the single source of truth
2. **Compiler First** — All project generation goes through Compiler
3. **Engineering First** — Generate real, runnable projects
4. **Platform First** — Design for extensibility, not feature completeness

### 3.2 Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Workflow is pure DSL | No runtime state in workflow.json |
| Compiler generates projects | Workflow → Compiler → Generator → Project |
| Runtime is unified | Runtime doesn't know about workflow, only executes commands |
| Agent outputs only workflow.json | Agent never generates code |
| Skill is workflow template | Skill never generates code |
| Project is filesystem directory | Real project, not database record |
| Generator creates standard projects | Python, MATLAB, ROS2, etc. — each is a standard project |
| Environment is Runtime Bundle | Versioned, cached, shared across projects |

### 3.3 Success Criteria

- [ ] Workflow.json contains only declaration data
- [ ] Compiler can generate a runnable Python project from workflow
- [ ] Agent generates only workflow.json, never code
- [ ] Runtime executes any project type via standard commands
- [ ] All projects are real filesystem directories
- [ ] Plugin system supports permission declaration
- [ ] Diagnostic AI analyzes logs and maps to workflow nodes
- [ ] Full test coverage for core modules

---

## 4. New Architecture Design

### 4.1 V2 System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         UI Layer (Vue3 + Tauri)                       │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │ Workflow  │ │ Dashboard │ │  Log     │ │ Plugin   │ │ Settings │  │
│  │  Editor   │ │          │ │  Center  │ │  Store   │ │          │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
└───────────────────────────┬─────────────────────────────────────────┘
                            │ HTTP REST / WebSocket / Tauri IPC
┌───────────────────────────┴─────────────────────────────────────────┐
│                      API Gateway (Gin)                               │
│  Auth | Rate Limit | CORS | Logging | Recovery | WebSocket           │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────────────┐
│                        Service Layer                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │ Project  │ │ Workflow │ │  Agent   │ │  Plugin  │ │  Log     │  │
│  │ Service  │ │ Service  │ │  Service │ │  Service │ │  Service │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────────────┐
│                    Core Engine Layer                                  │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                     Workflow DSL                               │   │
│  │  Pure declaration types | Validator | Parser | DAG            │   │
│  └──────────────────────────────┬───────────────────────────────┘   │
│                                 │                                    │
│  ┌──────────────────────────────┴───────────────────────────────┐   │
│  │                     Compiler                                   │   │
│  │  Compiler Interface | Generator Registry | Pipeline           │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │   │
│  │  │ Python   │ │ MATLAB   │ │  ROS2    │ │  Docker  │        │   │
│  │  │Generator │ │Generator │ │ Generator│ │ Generator│        │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │   │
│  └──────────────────────────────┬───────────────────────────────┘   │
│                                 │                                    │
│  ┌──────────────────────────────┴───────────────────────────────┐   │
│  │                     Runtime                                   │   │
│  │  Environment | Bundle | Lifecycle | Execution | Logging      │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌──────────────────────────┐ ┌──────────────────────────┐         │
│  │      Skill Manager        │ │      Task Scheduler       │         │
│  │  Workflow Templates       │ │  Queue | Workers | State │         │
│  └──────────────────────────┘ └──────────────────────────┘         │
│                                                                      │
│  ┌──────────────────────────┐ ┌──────────────────────────┐         │
│  │      Diagnostic           │ │      Event Bus           │         │
│  │  AI Error Analysis        │ │  Pub/Sub | Events        │         │
│  └──────────────────────────┘ └──────────────────────────┘         │
│                                                                      │
│  ┌──────────────────────────┐ ┌──────────────────────────┐         │
│  │      Plugin Manager       │ │      MCP Bridge          │         │
│  │  Load | Security | Exec   │ │  Model Context Protocol  │         │
│  └──────────────────────────┘ └──────────────────────────┘         │
│                                                                      │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────────────┐
│                      Data Layer                                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │ SQLite / │ │    FS    │ │  Cache   │ │  Log     │ │  Config  │  │
│  │ Postgres │ │  Storage │ │  Layer   │ │  Store   │ │  Store   │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.2 V2 Data Flow

```
User Request
    │
    ▼
Agent ──► Workflow DSL (workflow.json)
    │
    ▼
Compiler ──► Generator Registry
    │           │
    │           ├──► Python Generator ──► Python Project (filesystem)
    │           ├──► MATLAB Generator ──► MATLAB Project (filesystem)
    │           ├──► ROS2 Generator  ──► ROS2 Package (filesystem)
    │           └──► ... Generator   ──► ... Project (filesystem)
    │
    ▼
Runtime ──► Environment Check
    │           │
    │           ├──► Install Runtime Bundle
    │           ├──► Execute Project (standard command)
    │           └──► Stream Logs
    │
    ▼
Diagnostic ──► AI Analysis
    │           │
    │           ├──► Translate Logs
    │           ├──► Locate Workflow Node
    │           └──► Suggest Fixes
    │
    ▼
User Review
```

### 4.3 V2 Component Dependency

```
api/ → service/ → agent/ → workflow/ → compiler/ → generators/
  │       │          │         │            │
  │       │          │         │            └──→ runtime/
  │       │          │         │
  │       │          │         └──→ skill/
  │       │          │
  │       │          └──→ mcp/
  │       │
  │       ├──→ project/ → fs/
  │       ├──→ plugin/
  │       ├──→ logcenter/
  │       ├──→ diagnostic/
  │       ├──→ eventbus/ (global)
  │       └──→ config/ (global)
```

**Key Rules:**
- No circular dependencies
- `workflow/` knows nothing about `compiler/` or `runtime/`
- `compiler/` knows nothing about `runtime/`
- `runtime/` knows nothing about `workflow/`
- `agent/` only outputs workflow.json, never calls `compiler/` or `runtime/`
- `generators/` only depend on `workflow/` types

---

## 5. Module Division

### 5.1 Module Inventory

| Module | Responsibility | Dependencies | Status |
|--------|---------------|--------------|--------|
| `workflow/` | DSL types, parser, validator, DAG | None | 🔄 Refactor |
| `compiler/` | Compile workflow → generator pipeline | `workflow/` | 🆕 Create |
| `generators/` | Project generators (Python, MATLAB, etc.) | `workflow/` | 🆕 Create |
| `runtime/` | Unified executor, bundle management | None | 🔄 Refactor |
| `skill/` | Workflow templates | `workflow/` | 🆕 Create |
| `agent/` | NL → Workflow generation | `workflow/`, `skill/` | 🔄 Refactor |
| `logcenter/` | Log collection, storage, streaming | `eventbus/` | 🔄 Refactor |
| `diagnostic/` | AI error analysis, translation | `logcenter/`, `workflow/` | 🆕 Create |
| `plugin/` | Plugin lifecycle, security, execution | `eventbus/` | 🔄 Enhance |
| `task/` | Task queue, scheduling, state machine | `eventbus/` | ✅ Keep |
| `project/` | Project CRUD, filesystem management | `fs/` | 🔄 Refactor |
| `eventbus/` | Pub/sub event system | None | 🆕 Create |
| `mcp/` | MCP protocol bridge | `eventbus/` | ✅ Keep |
| `auth/` | Authentication, authorization | `database/` | ✅ Keep |
| `config/` | Configuration management | None | 🔄 Refactor |
| `api/` | HTTP handlers, middleware | `service/` | 🔄 Refactor |
| `service/` | Business logic layer | All modules | 🔄 Refactor |

### 5.2 Module Dependency Graph

```
eventbus/ (no dependencies)
    ↑
config/ (no dependencies)
    ↑
workflow/ (no dependencies)
    ↑
compiler/ ───→ generators/
    ↑
runtime/ ───→ environment/
    ↑
skill/ ───→ workflow/
    ↑
agent/ ───→ workflow/, skill/, mcp/
    ↑
logcenter/ ───→ eventbus/
    ↑
diagnostic/ ───→ logcenter/, workflow/
    ↑
plugin/ ───→ eventbus/, config/
    ↑
task/ ───→ eventbus/, plugin/
    ↑
project/ ───→ fs/
    ↑
service/ ───→ all modules (facade)
    ↑
api/ ───→ service/
    ↑
auth/ ───→ database/
```

---

## 6. Monorepo Directory Design

### 6.1 Target Directory Structure

```
AIStudio/
├── .github/                     # GitHub Actions, CI/CD
├── Backend/                     # Go backend (module: github.com/aistudio/backend)
│   ├── cmd/
│   │   └── aistudio/
│   │       └── main.go          # Clean entry point
│   ├── internal/
│   │   ├── api/                 # HTTP API layer
│   │   │   ├── handlers/        # Route handlers
│   │   │   ├── middleware/      # Auth, cors, logger, rate-limit
│   │   │   └── router.go       # Route registration
│   │   ├── service/            # Business logic facade
│   │   │   ├── project_service.go
│   │   │   ├── workflow_service.go
│   │   │   ├── agent_service.go
│   │   │   ├── compiler_service.go
│   │   │   ├── runtime_service.go
│   │   │   ├── log_service.go
│   │   │   ├── diagnostic_service.go
│   │   │   ├── plugin_service.go
│   │   │   ├── skill_service.go
│   │   │   ├── task_service.go
│   │   │   └── service.go       # Service container
│   │   ├── workflow/            # 🆕 Pure DSL (no runtime state)
│   │   │   ├── types.go         # Workflow, Node, Edge, Port declarations
│   │   │   ├── parser.go        # JSON parse/validate
│   │   │   ├── validator.go     # DAG validation
│   │   │   ├── dag.go           # Topological sort
│   │   │   └── interfaces.go    # Public interfaces
│   │   ├── compiler/            # 🆕 Compiler module
│   │   │   ├── compiler.go      # Compiler interface + implementation
│   │   │   ├── registry.go      # Generator registry
│   │   │   ├── pipeline.go      # Compilation pipeline
│   │   │   └── types.go         # Compiler types
│   │   ├── generators/          # 🆕 Project generators
│   │   │   ├── generator.go     # Generator interface
│   │   │   ├── python/          # Python project generator
│   │   │   │   ├── generator.go
│   │   │   │   └── templates/   # Go templates
│   │   │   ├── matlab/          # MATLAB project generator
│   │   │   │   └── generator.go
│   │   │   ├── ros2/            # ROS2 package generator
│   │   │   │   └── generator.go
│   │   │   ├── docker/          # Docker project generator
│   │   │   │   └── generator.go
│   │   │   └── stm32/           # STM32 CubeMX generator
│   │   │       └── generator.go
│   │   ├── runtime/             # 🔄 Unified runtime
│   │   │   ├── runtime.go       # Runtime interface
│   │   │   ├── executor.go      # Command executor
│   │   │   ├── lifecycle.go     # Process lifecycle
│   │   │   ├── bundles/         # Runtime bundles
│   │   │   │   ├── bundle.go    # Bundle interface
│   │   │   │   ├── python.go    # Python bundle
│   │   │   │   ├── matlab.go    # MATLAB bundle
│   │   │   │   └── ros2.go      # ROS2 bundle
│   │   │   └── types.go
│   │   ├── skill/               # 🆕 Skill templates
│   │   │   ├── skill.go         # Skill interface
│   │   │   ├── registry.go      # Skill registry
│   │   │   ├── loader.go        # Load from JSON/plugin
│   │   │   └── templates/       # Built-in templates
│   │   │       ├── yolo_detection.json
│   │   │       ├── transformer_classification.json
│   │   │       └── traffic_simulation.json
│   │   ├── logcenter/           # 🔄 Log center
│   │   │   ├── collector.go     # Log collection
│   │   │   ├── store.go         # Log storage
│   │   │   ├── stream.go        # Real-time streaming
│   │   │   └── types.go
│   │   ├── diagnostic/          # 🆕 AI Diagnostic
│   │   │   ├── diagnostic.go    # Diagnostic interface
│   │   │   ├── analyzer.go      # Log analysis
│   │   │   ├── translator.go    # Log translation
│   │   │   ├── linker.go        # Workflow node linker
│   │   │   └── types.go
│   │   ├── plugin/              # 🔄 Enhanced plugin
│   │   │   ├── manager.go       # Lifecycle management
│   │   │   ├── loader.go        # Plugin loading
│   │   │   ├── registry.go      # Plugin registry
│   │   │   ├── security.go      # 🆕 Permission + signature
│   │   │   ├── sandbox.go       # 🆕 Sandbox execution
│   │   │   ├── executor.go      # Plugin execution
│   │   │   └── types.go
│   │   ├── agent/               # 🔄 Refactored agent
│   │   │   ├── agent.go         # Agent core
│   │   │   ├── planner.go       # NL → workflow plan
│   │   │   ├── executor.go      # Execute plan → workflow.json
│   │   │   ├── memory.go        # Conversation memory
│   │   │   ├── llm_provider.go  # LLM integration
│   │   │   └── types.go
│   │   ├── task/                # ✅ Task scheduler (keep)
│   │   │   ├── manager.go
│   │   │   ├── queue.go
│   │   │   ├── worker.go
│   │   │   ├── scheduler.go
│   │   │   ├── state.go
│   │   │   └── types.go
│   │   ├── eventbus/            # 🆕 Event bus
│   │   │   ├── eventbus.go      # Pub/sub implementation
│   │   │   ├── topics.go        # Event topics
│   │   │   └── types.go
│   │   ├── project/             # 🔄 Project manager
│   │   │   ├── manager.go       # Project CRUD + filesystem
│   │   │   ├── template.go      # Project template
│   │   │   └── types.go
│   │   ├── environment/         # 🔄 Environment manager
│   │   │   ├── manager.go
│   │   │   ├── detector.go
│   │   │   ├── installer.go
│   │   │   └── types.go
│   │   ├── mcp/                 # ✅ MCP bridge (keep)
│   │   │   ├── manager.go
│   │   │   ├── client.go
│   │   │   ├── server.go
│   │   │   └── types.go
│   │   ├── auth/                # ✅ Auth (keep)
│   │   ├── config/              # 🔄 Unified config
│   │   │   ├── config.go
│   │   │   └── defaults.go
│   │   └── database/            # ✅ Database (keep)
│   │       ├── database.go
│   │       ├── migrate.go
│   │       └── models/
│   ├── pkg/                     # Shared packages
│   │   ├── plugin/              # Plugin SDK interfaces
│   │   └── runtime/             # Runtime SDK interfaces
│   ├── go.mod
│   └── go.sum
├── Frontend/                    # Vue3 + TypeScript + Tauri
│   ├── src/
│   │   ├── api/                 # API client
│   │   ├── components/          # Shared components
│   │   ├── composables/         # Vue composables
│   │   ├── pages/               # 🆕 Unified page components (migrate from views/)
│   │   │   ├── Dashboard/
│   │   │   ├── Workflow/
│   │   │   ├── AIChat/
│   │   │   ├── Project/
│   │   │   ├── Logs/
│   │   │   ├── PluginStore/
│   │   │   └── Settings/
│   │   ├── stores/              # 🆕 Unified store (remove store/)
│   │   ├── router/
│   │   ├── types/
│   │   └── App.vue
│   ├── package.json
│   └── vite.config.ts
├── Engine/                      # Python AI Engine (simplified)
│   ├── runtime/                 # Python runtime entry
│   ├── sdk/                     # Python SDK
│   └── runner.py                # Task runner (generate → execute)
├── Plugins/                     # Plugin directory
│   ├── built-in/                # Built-in plugins
│   └── community/               # Community plugins
├── Config/                      # 🆕 Unified config (merged)
│   ├── app.yaml
│   └── backend.yaml
├── Docs/                        # Refined documentation
│   ├── architecture/
│   ├── api/
│   ├── compiler/
│   ├── runtime/
│   ├── plugin-sdk/
│   ├── ADR/                     # Architecture Decision Records
│   └── RFC/                     # RFCs
├── Scripts/                     # Build scripts
├── Tests/                       # 🆕 Integration tests
│   ├── backend/
│   ├── frontend/
│   └── e2e/
├── Examples/                    # 🆕 Example projects
│   ├── yolo-detection/
│   ├── traffic-simulation/
│   └── nlp-classification/
└── Storage/                     # Runtime storage
    ├── projects/
    ├── datasets/
    ├── models/
    └── bundles/                 # Runtime bundles cache
```

---

## 7. Compiler Architecture

### 7.1 Design Philosophy

The Compiler is the **heart of AIStudio V2**. It transforms a declarative Workflow into a real, runnable engineering project. This is inspired by how traditional compilers work:

```
Source Code (DSL) → Compiler Frontend → IR → Compiler Backend → Target Code
Workflow JSON    → Workflow Parser   → IR → Generator        → Project Files
```

### 7.2 Compiler Interface

```go
// Compiler is the core compilation engine.
// It reads a Workflow, selects the appropriate Generator,
// and produces a complete project directory.
type Compiler interface {
    // Compile compiles a workflow into a project directory.
    // Returns the project root path and any errors.
    Compile(ctx context.Context, wf *workflow.Workflow, opts CompileOptions) (*CompileResult, error)
    
    // Generators returns the list of registered generators.
    Generators() []GeneratorInfo
}

// CompileOptions controls compilation behavior.
type CompileOptions struct {
    OutputDir  string            // Output directory for generated project
    Target     Target            // Target platform (python, matlab, ros2, etc.)
    Variables  map[string]string // Template variables
    Force      bool              // Overwrite existing output
    DryRun     bool              // Validate without writing files
}

// CompileResult contains the compilation output.
type CompileResult struct {
    Target      Target            // Target platform
    ProjectRoot string            // Absolute path to generated project
    EntryPoints []string          // Entry point files
    Files       []GeneratedFile   // All generated files
    RuntimeReq  *RuntimeRequirement // Required runtime bundle
    Duration    time.Duration
    WorkflowID  string
}

// GeneratedFile represents a single generated file.
type GeneratedFile struct {
    Path    string // Relative path within project
    Content string // File content
    Mode    os.FileMode
}
```

### 7.3 Generator Interface

```go
// Generator generates a complete project from a workflow.
// Each Generator is responsible for one target platform.
type Generator interface {
    // ID returns the unique generator identifier.
    ID() Target
    
    // Name returns the human-readable generator name.
    Name() string
    
    // Generate generates a complete project from a workflow.
    Generate(ctx context.Context, wf *workflow.Workflow, outputDir string) (*GenerateResult, error)
    
    // RuntimeRequirement returns the runtime bundle required for this project.
    RuntimeRequirement(wf *workflow.Workflow) (*RuntimeRequirement, error)
    
    // Validate checks if the workflow can be compiled to this target.
    Validate(wf *workflow.Workflow) error
}

// Target identifies the target platform.
type Target string

const (
    TargetPython Target = "python"
    TargetMATLAB Target = "matlab"
    TargetROS2   Target = "ros2"
    TargetSTM32  Target = "stm32"
    TargetDocker Target = "docker"
    TargetCPP    Target = "cpp"
    TargetUnity  Target = "unity"
)

// GenerateResult contains the generation output.
type GenerateResult struct {
    Target      Target
    Files       []GeneratedFile
    EntryPoints []string
    ProjectName string
}
```

### 7.4 Compilation Pipeline

```
Workflow JSON
    │
    ▼
┌─────────────────────────────────────┐
│  Phase 1: Parse & Validate          │
│  • Parse workflow.json              │
│  • Validate schema                  │
│  • Validate DAG (no cycles)         │
│  • Validate node types exist        │
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐
│  Phase 2: Target Selection          │
│  • Detect target from workflow      │
│  • Select matching Generator        │
│  • Validate generator compatibility │
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐
│  Phase 3: IR Generation             │
│  • Convert workflow to IR           │
│  • Resolve variables                │
│  • Optimize DAG                     │
│  • Generate dependency graph        │
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐
│  Phase 4: Project Generation        │
│  • Create directory structure       │
│  • Generate source files            │
│  • Generate config files            │
│  • Generate entry point             │
│  • Generate requirements/deps       │
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐
│  Phase 5: Validation                │
│  • Validate project structure       │
│  • Validate all files generated     │
│  • Run syntax check (if available)  │
│  • Return CompileResult             │
└─────────────────────────────────────┘
```

---

## 8. Workflow DSL

### 8.1 Design Principles

1. **Pure Declaration** — Workflow contains only "what", not "how"
2. **No Runtime State** — No status, progress, logs, errors in workflow.json
3. **Self-Describing** — Workflow contains all information needed for compilation
4. **Versioned Schema** — Schema version for backward compatibility
5. **Extensible** — Custom metadata, variables, tags

### 8.2 Workflow JSON Schema (V2)

```json
{
  "schema_version": "2.0.0",
  "id": "wf-1234-5678",
  "name": "YOLO Target Detection Pipeline",
  "description": "Complete pipeline for vehicle detection training",
  "version": 1,
  "author": "user@example.com",
  "tags": ["vision", "yolo", "detection", "training"],
  "metadata": {
    "created_with": "skill:yolo-detection",
    "skill_version": "1.0.0"
  },
  "variables": {
    "dataset_path": "/storage/datasets/vehicle",
    "epochs": 100,
    "batch_size": 16
  },
  "target": "python",
  "nodes": [
    {
      "id": "node-1",
      "type": "data_loader",
      "name": "Dataset Loader",
      "description": "Load vehicle detection dataset",
      "position": { "x": 100, "y": 200 },
      "config": {
        "source": "${dataset_path}",
        "format": "yolo",
        "split_ratio": 0.8
      },
      "inputs": [],
      "outputs": [
        { "id": "out_dataset", "name": "dataset", "type": "dataset" }
      ]
    },
    {
      "id": "node-2",
      "type": "model_trainer",
      "name": "YOLOv8 Trainer",
      "description": "Train YOLOv8 model",
      "position": { "x": 400, "y": 200 },
      "config": {
        "model": "yolov8n.pt",
        "epochs": "${epochs}",
        "batch_size": "${batch_size}",
        "device": "cuda"
      },
      "inputs": [
        { "id": "in_dataset", "name": "dataset", "type": "dataset", "required": true }
      ],
      "outputs": [
        { "id": "out_model", "name": "trained_model", "type": "model" }
      ]
    },
    {
      "id": "node-3",
      "type": "model_exporter",
      "name": "ONNX Exporter",
      "description": "Export to ONNX format",
      "position": { "x": 700, "y": 200 },
      "config": {
        "format": "onnx",
        "optimize": true
      },
      "inputs": [
        { "id": "in_model", "name": "model", "type": "model", "required": true }
      ],
      "outputs": [
        { "id": "out_export", "name": "exported_model", "type": "file" }
      ]
    }
  ],
  "edges": [
    {
      "id": "edge-1",
      "source": { "node_id": "node-1", "port_id": "out_dataset" },
      "target": { "node_id": "node-2", "port_id": "in_dataset" }
    },
    {
      "id": "edge-2",
      "source": { "node_id": "node-2", "port_id": "out_model" },
      "target": { "node_id": "node-3", "port_id": "in_model" }
    }
  ]
}
```

### 8.3 Key Differences from V1

| Aspect | V1 | V2 |
|--------|----|----|
| `Node.Runtime` | Included (status, progress, logs) | **Removed** — runtime state is separate |
| `Node.Parameters` | `map[string]interface{}` | Renamed to `Config` |
| `Workflow.Target` | Not present | **Added** — declares target platform |
| `Workflow.Variables` | Optional | **Standardized** — template variables |
| `ExecutionResult` | In workflow engine | **Moved** to runtime module |
| `NodeConstraints` | In node | **Moved** to generator config |

### 8.4 Workflow Package Structure (V2)

```go
package workflow

// Workflow is the single source of truth.
// It contains only declaration data, no runtime state.
type Workflow struct {
    SchemaVersion string            `json:"schema_version"`
    ID            string            `json:"id"`
    Name          string            `json:"name"`
    Description   string            `json:"description,omitempty"`
    Version       int               `json:"version"`
    Author        string            `json:"author,omitempty"`
    Tags          []string          `json:"tags,omitempty"`
    Metadata      map[string]any    `json:"metadata,omitempty"`
    Variables     map[string]string `json:"variables,omitempty"`
    Target        Target            `json:"target"`
    Nodes         []Node            `json:"nodes"`
    Edges         []Edge            `json:"edges"`
}

// Node is a pure declaration node.
type Node struct {
    ID          string            `json:"id"`
    Type        string            `json:"type"`
    Name        string            `json:"name"`
    Description string            `json:"description,omitempty"`
    Position    Point             `json:"position"`
    Config      map[string]any    `json:"config,omitempty"`
    Inputs      []Port            `json:"inputs"`
    Outputs     []Port            `json:"outputs"`
}

// Edge connects two nodes.
type Edge struct {
    ID     string       `json:"id"`
    Source EdgeEndpoint `json:"source"`
    Target EdgeEndpoint `json:"target"`
}

// Target is the target platform for compilation.
type Target string

const (
    TargetPython Target = "python"
    TargetMATLAB Target = "matlab"
    TargetROS2   Target = "ros2"
    TargetSTM32  Target = "stm32"
    TargetDocker Target = "docker"
    TargetCPP    Target = "cpp"
)
```

---

## 9. Runtime Lifecycle

### 9.1 Runtime States

```
┌──────────┐
│  Idle     │
└────┬─────┘
     │
     ▼
┌──────────┐     ┌──────────┐
│ Detecting │────►│  Ready   │
└──────────┘     └────┬─────┘
                      │
                      ▼
              ┌──────────────┐
              │ Installing    │◄──── Retry
              │ Bundle        │
              └──────┬───────┘
                     │
                     ▼
              ┌──────────────┐
              │   Prepared   │
              └──────┬───────┘
                     │
                     ▼
              ┌──────────────┐
              │  Executing   │
              │  (Running)   │
              └──────┬───────┘
                     │
              ┌──────┴──────┐
              │             │
              ▼             ▼
        ┌──────────┐  ┌──────────┐
        │ Completed │  │  Failed  │
        └──────────┘  └──────────┘
```

### 9.2 Runtime Interface

```go
// Runtime is the unified execution engine.
// It does not know about Workflow — it only executes projects.
type Runtime interface {
    // Detect checks if the environment meets the runtime requirements.
    Detect(ctx context.Context, req *RuntimeRequirement) (*EnvironmentReport, error)
    
    // Prepare installs the runtime bundle and prepares the environment.
    Prepare(ctx context.Context, req *RuntimeRequirement) error
    
    // Execute runs the project with the given configuration.
    Execute(ctx context.Context, project *Project, config *RunConfig) (*RunResult, error)
    
    // Stop terminates a running execution.
    Stop(ctx context.Context, runID string) error
    
    // Status returns the current status of a running execution.
    Status(ctx context.Context, runID string) (*RunStatus, error)
}

// RuntimeRequirement declares what runtime environment is needed.
type RuntimeRequirement struct {
    Name     string   // Bundle name, e.g. "yolo", "transformer"
    Version  string   // Bundle version
    Python   string   // Python version requirement
    Packages []string // pip packages
    Commands []string // Required commands (e.g., ["python", "matlab"])
    GPU      bool     // GPU required
    MemoryMB int      // Minimum memory
    DiskMB   int      // Minimum disk space
}

// RunConfig controls execution behavior.
type RunConfig struct {
    ProjectDir string            // Project directory
    EntryPoint string            // Entry point file/command
    Args       []string          // Command arguments
    Env        map[string]string // Environment variables
    Timeout    time.Duration     // Execution timeout
    LogCallback func(*LogEntry)  // Real-time log callback
}
```

### 9.3 Runtime Bundle System

```go
// BundleManager manages runtime bundles.
type BundleManager interface {
    // List returns all installed bundles.
    List() []*Bundle
    
    // Install installs a runtime bundle.
    Install(ctx context.Context, req *RuntimeRequirement) (*Bundle, error)
    
    // Uninstall removes a runtime bundle.
    Uninstall(name string) error
    
    // Get returns a bundle by name.
    Get(name string) (*Bundle, bool)
    
    // CachePath returns the cache directory for bundles.
    CachePath() string
}

// Bundle is a versioned runtime environment.
type Bundle struct {
    Name        string
    Version     string
    PythonPath  string
    Packages    []string
    EnvVars     map[string]string
    Path        string // Installation path in cache
    InstalledAt time.Time
}
```

---

## 10. Generator Interface

### 10.1 Generator Registration

```go
// GeneratorRegistry manages all registered generators.
type GeneratorRegistry struct {
    generators map[Target]Generator
}

func (r *GeneratorRegistry) Register(g Generator) {
    r.generators[g.ID()] = g
}

func (r *GeneratorRegistry) Get(target Target) (Generator, bool) {
    g, ok := r.generators[target]
    return g, ok
}

func (r *GeneratorRegistry) List() []GeneratorInfo {
    // Returns metadata for all registered generators
}
```

### 10.2 Python Generator Example

```go
// PythonGenerator generates Python projects from workflows.
type PythonGenerator struct{}

func (g *PythonGenerator) ID() Target { return TargetPython }

func (g *PythonGenerator) Generate(ctx context.Context, wf *workflow.Workflow, outputDir string) (*GenerateResult, error) {
    // 1. Create project directory structure
    //    project/
    //    ├── main.py
    //    ├── requirements.txt
    //    ├── config.yaml
    //    ├── data/
    //    ├── models/
    //    ├── src/
    //    │   ├── __init__.py
    //    │   ├── data_loader.py
    //    │   ├── model.py
    //    │   └── train.py
    //    └── outputs/
    
    // 2. Generate requirements.txt from workflow nodes
    // 3. Generate config.yaml from workflow variables
    // 4. Generate main.py as entry point (DAG execution)
    // 5. Generate individual module files
    // 6. Generate README.md
    // 7. Return result
}
```

### 10.3 Generator Generated Project Structure

```
project-name/
├── .aistudio/
│   └── workflow.json         # Original workflow (for reference)
├── src/
│   ├── __init__.py
│   ├── data_loader.py
│   ├── model.py
│   ├── train.py
│   └── utils.py
├── config/
│   └── config.yaml
├── data/                     # Data directory
├── models/                   # Model output directory
├── outputs/                  # Output directory
├── requirements.txt
├── setup.py
├── main.py                   # Entry point
├── README.md                 # Generated documentation
└── .gitignore
```

---

## 11. Environment Manager

### 11.1 Architecture

The Environment Manager is refactored from V1's `environment/` package to support **Runtime Bundles**:

```
Environment Manager
    │
    ├── Detector — Detect system capabilities
    │   ├── Python detection
    │   ├── CUDA/GPU detection
    │   ├── System package detection
    │   └── Command availability
    │
    ├── Installer — Install runtime dependencies
    │   ├── Pip installer
    │   ├── Conda installer
    │   ├── System package installer
    │   └── Bundle installer
    │
    ├── Bundle Manager — Manage runtime bundles
    │   ├── Version management
    │   ├── Cache management
    │   ├── Dependency resolution
    │   └── Cross-project sharing
    │
    └── Repair — Auto-repair common issues
        ├── Python path repair
        ├── Dependency conflict resolution
        └── Environment variable repair
```

### 11.2 Bundle Cache

```
Storage/bundles/
├── python-3.10/
│   ├── bin/
│   ├── lib/
│   └── manifest.json
├── yolo-v8/
│   ├── lib/
│   ├── packages/
│   └── manifest.json
├── transformer-4.30/
│   ├── lib/
│   └── manifest.json
└── ros2-humble/
    ├── setup.bash
    └── manifest.json
```

---

## 12. Logger

### 12.1 Log Center Architecture

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Collector    │───►│  Store       │───►│  Stream      │
│  (gather logs)│    │  (persist)   │    │  (real-time) │
└──────┬───────┘    └──────┬───────┘    └──────┬───────┘
       │                   │                   │
       │            ┌──────┴───────┐           │
       │            │  Database    │           │
       │            │  + File      │           │
       │            └──────────────┘           │
       │                                       │
       │            ┌──────────────────────────┘
       │            │
       ▼            ▼
┌──────────────────────────────────────┐
│         Diagnostic Engine             │
│  AI Analysis | Translation | Linking │
└──────────────────────────────────────┘
```

### 12.2 Log Levels

```go
type LogLevel string

const (
    LogLevelDebug   LogLevel = "DEBUG"
    LogLevelInfo    LogLevel = "INFO"
    LogLevelWarning LogLevel = "WARN"
    LogLevelError   LogLevel = "ERROR"
    LogLevelFatal   LogLevel = "FATAL"
)
```

### 12.3 Log Entry Structure

```go
type LogEntry struct {
    ID          string    `json:"id"`
    Timestamp   time.Time `json:"timestamp"`
    Level       LogLevel  `json:"level"`
    Source      string    `json:"source"`      // Module name
    Message     string    `json:"message"`
    Detail      string    `json:"detail,omitempty"`
    TaskID      string    `json:"taskId,omitempty"`
    WorkflowID  string    `json:"workflowId,omitempty"`
    NodeID      string    `json:"nodeId,omitempty"`  // Workflow node
    RunID       string    `json:"runId,omitempty"`
    Raw         string    `json:"raw,omitempty"`     // Original raw log
    Metadata    map[string]any `json:"metadata,omitempty"`
}
```

---

## 13. Diagnostic

### 13.1 AI Diagnostic Engine

The Diagnostic module is a **new module** that provides AI-powered error analysis:

```
Diagnostic Engine
    │
    ├── Log Analyzer — AI analysis of execution logs
    │   ├── Error pattern detection
    │   ├── Root cause analysis
    │   ├── Severity classification
    │   └── Historical comparison
    │
    ├── Translator — Log translation
    │   ├── Technical → Human readable
    │   ├── English → Chinese (and other languages)
    │   └── Stack trace simplification
    │
    ├── Workflow Linker — Map errors to workflow nodes
    │   ├── Error → Node mapping
    │   ├── Dependency chain analysis
    │   └── Configuration validation
    │
    └── Fix Suggester — Generate fix recommendations
        ├── Parameter adjustment
        ├── Dependency installation
        ├── Environment repair
        └── Workflow modification
```

### 13.2 Diagnostic Interface

```go
// Diagnostic provides AI-powered error analysis.
type Diagnostic interface {
    // Analyze analyzes a log entry and returns diagnostic results.
    Analyze(ctx context.Context, entry *logcenter.LogEntry, wf *workflow.Workflow) (*DiagnosticResult, error)
    
    // AnalyzeTask analyzes all logs for a task.
    AnalyzeTask(ctx context.Context, taskID string) (*TaskDiagnostic, error)
    
    // Translate translates a technical error message to human-readable form.
    Translate(ctx context.Context, message string, lang string) (string, error)
    
    // SuggestFix generates fix suggestions for an error.
    SuggestFix(ctx context.Context, result *DiagnosticResult) ([]*FixSuggestion, error)
}

type DiagnosticResult struct {
    ErrorID     string              `json:"errorId"`
    Severity    Severity            `json:"severity"`
    Summary     string              `json:"summary"`         // Human-readable summary
    Detail      string              `json:"detail"`          // Technical detail
    WorkflowID  string              `json:"workflowId,omitempty"`
    NodeID      string              `json:"nodeId,omitempty"` // Related workflow node
    NodeName    string              `json:"nodeName,omitempty"`
    RootCause   string              `json:"rootCause"`
    Suggestions []*FixSuggestion    `json:"suggestions,omitempty"`
    RawLog      string              `json:"rawLog"`          // Original log
}

type FixSuggestion struct {
    ID          string          `json:"id"`
    Type        FixType         `json:"type"`    // parameter, dependency, environment, workflow
    Title       string          `json:"title"`
    Description string          `json:"description"`
    AutoFix     bool            `json:"autoFix"` // Can be applied automatically
    Action      *FixAction      `json:"action,omitempty"`
}

type FixAction struct {
    Type    string      `json:"type"`    // update_config, install_package, modify_workflow
    Target  string      `json:"target"`
    Value   interface{} `json:"value"`
}
```

---

## 14. Plugin

### 14.1 V2 Plugin Enhancements

| Feature | V1 | V2 |
|---------|----|----|
| Plugin types | Vision, NLP, Logic, System, MCP | + Generator, Skill, Runtime, UI, Log, Diagnostic |
| Permissions | None | **Declared permissions** |
| Signing | None | **Digital signature verification** |
| Sandbox | None | **Sandbox execution** |
| Lifecycle | Basic | **Full lifecycle management** |
| Dependency | Basic | **Versioned dependency resolution** |

### 14.2 Plugin Manifest (V2)

```json
{
  "name": "yolo-detector",
  "version": "2.0.0",
  "type": "generator",
  "description": "YOLO target detection generator",
  "author": "AIStudio",
  "license": "MIT",
  "permissions": [
    "filesystem:read:storage/datasets",
    "filesystem:write:storage/models",
    "network:localhost",
    "gpu:compute"
  ],
  "signature": "base64-encoded-signature",
  "extends": {
    "generators": ["python"],
    "skills": ["yolo-detection"],
    "nodes": ["data_loader", "model_trainer", "model_exporter"]
  },
  "dependencies": {
    "python": ">=3.9,<3.12",
    "pip": {
      "ultralytics": ">=8.0.0",
      "torch": ">=2.0.0"
    }
  },
  "runtime": {
    "bundle": "yolo-v8",
    "gpu": true,
    "min_memory_mb": 4096
  }
}
```

### 14.3 Plugin Security

```go
// Permission represents a declared plugin permission.
type Permission string

const (
    PermissionFileRead     Permission = "filesystem:read"
    PermissionFileWrite    Permission = "filesystem:write"
    PermissionNetwork      Permission = "network"
    PermissionGPU          Permission = "gpu:compute"
    PermissionProcess      Permission = "process:exec"
    PermissionEnvironment  Permission = "environment:modify"
)

// SecurityManager handles plugin security.
type SecurityManager interface {
    // VerifySignature verifies the plugin's digital signature.
    VerifySignature(manifest *PluginManifest) error
    
    // CheckPermissions validates that the plugin's permissions are allowed.
    CheckPermissions(manifest *PluginManifest) error
    
    // CreateSandbox creates a sandboxed execution environment.
    CreateSandbox(plugin *Plugin) (*Sandbox, error)
    
    // RevokePermissions revokes a plugin's permissions.
    RevokePermissions(pluginID string) error
}
```

---

## 15. Cloud

### 15.1 Cloud Architecture (Future)

```
┌─────────────────────────────────────────────────────────────┐
│                   AIStudio Cloud                             │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Cloud API Gateway                        │   │
│  └──────────────────────┬───────────────────────────────┘   │
│                         │                                    │
│  ┌──────────────────────┴───────────────────────────────┐   │
│  │              Cloud Services                           │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐│   │
│  │  │ Project  │ │ Workflow │ │  Agent   │ │  Runtime ││   │
│  │  │ Sync     │ │  Share   │ │  Cloud   │ │  Remote  ││   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘│   │
│  └──────────────────────┬───────────────────────────────┘   │
│                         │                                    │
│  ┌──────────────────────┴───────────────────────────────┐   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐│   │
│  │  │  User    │ │  Plugin  │ │  Skill   │ │  Model   ││   │
│  │  │  Mgmt   │ │  Registry│ │  Store   │ │  Hub     ││   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘│   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Distributed Runtime                       │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐│   │
│  │  │  Local   │ │  Remote  │ │  Cluster │ │  Edge    ││   │
│  │  │  Runtime │ │  Runtime │ │  Runtime │ │  Runtime ││   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘│   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 15.2 Cloud Features (Phase 2+)

- **Project Sync** — Sync projects between local and cloud
- **Workflow Sharing** — Share workflows via cloud registry
- **Plugin Registry** — Cloud plugin marketplace
- **Skill Store** — Cloud skill template marketplace
- **Model Hub** — Model versioning and sharing
- **Remote Runtime** — Execute on cloud GPU instances
- **Team Collaboration** — Multi-user project collaboration

---

## 16. Security

### 16.1 Security Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Security Layer                             │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Authentication                                       │   │
│  │  JWT | API Keys | OAuth2 | Session Management        │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Authorization                                        │   │
│  │  RBAC | Permission Check | Resource Ownership        │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Plugin Security                                      │   │
│  │  Permission Declaration | Signature Verification     │   │
│  │  Sandbox Execution | Capability Limiting              │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Runtime Security                                     │   │
│  │  Process Isolation | Resource Limits | Timeout       │   │
│  │  Network Control | Filesystem Access Control          │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Data Security                                        │   │
│  │  Encryption at Rest | Encryption in Transit          │   │
│  │  Secrets Management | Key Rotation                   │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Audit                                                │   │
│  │  Security Logging | Audit Trail | Compliance         │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 16.2 Security Checklist

- [ ] JWT with refresh token rotation
- [ ] API key hashing (not plaintext)
- [ ] Rate limiting per user/IP
- [ ] SQL injection prevention (GORM parameterized queries)
- [ ] XSS prevention (no raw HTML rendering)
- [ ] CSRF protection
- [ ] Plugin permission whitelist
- [ ] Plugin digital signature verification
- [ ] Sandboxed process execution
- [ ] Secrets manager (not in config files)
- [ ] Audit logging for all security events
- [ ] Regular dependency scanning (go.mod, package.json)

---

## 17. Project Manager

### 17.1 Project as Filesystem

```go
// ProjectManager manages project lifecycle on the filesystem.
type ProjectManager interface {
    // Create creates a new project directory.
    Create(ctx context.Context, name string, opts CreateOptions) (*Project, error)
    
    // Open opens an existing project.
    Open(path string) (*Project, error)
    
    // Delete removes a project.
    Delete(id string) error
    
    // List returns all projects.
    List() ([]*Project, error)
    
    // Export exports a project to a distributable format.
    Export(id string, format ExportFormat) (string, error)
    
    // Import imports a project from a distributable format.
    Import(path string) (*Project, error)
}

// Project represents a project on the filesystem.
type Project struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    RootPath    string    `json:"rootPath"`     // Absolute path
    WorkflowID  string    `json:"workflowId"`   // Linked workflow
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

### 17.2 Project Directory Structure

```
Storage/projects/
└── my-project/
    ├── .aistudio/
    │   ├── workflow.json        # Workflow definition
    │   ├── project.json         # Project metadata
    │   └── config.yaml          # Project configuration
    ├── src/                     # Source code (generated)
    ├── data/                    # Data files
    ├── models/                  # Model outputs
    ├── outputs/                 # Execution outputs
    ├── requirements.txt         # Dependencies
    ├── main.py                  # Entry point
    └── README.md                # Project documentation
```

---

## 18. API Gateway

### 18.1 API Versioning

```
/api/v1/health
/api/v1/auth/login
/api/v1/auth/register
/api/v1/projects
/api/v1/workflows
/api/v1/compiler/compile
/api/v1/runtime/execute
/api/v1/runtime/bundles
/api/v1/skills
/api/v1/plugins
/api/v1/agent/chat
/api/v1/logs
/api/v1/diagnostic/analyze
/api/v1/mcp/tools
/api/v1/environment/status
```

### 18.2 API Grouping

```
/api/v1/
├── health
├── auth/           # Login, register, refresh, logout
├── users/          # User management
├── projects/       # Project CRUD
├── workflows/      # Workflow CRUD + compile
├── compiler/       # Compile workflow → project
│   ├── compile
│   ├── targets     # List available targets
│   └── validate    # Validate workflow for target
├── runtime/        # Execute + bundle management
│   ├── execute
│   ├── stop
│   ├── status
│   └── bundles
├── skills/         # Skill templates
│   ├── list
│   ├── apply
│   └── create
├── tasks/          # Task management
├── plugins/        # Plugin management
├── agent/          # AI Agent
│   ├── chat
│   ├── plan
│   └── generate-workflow
├── logs/           # Log center
│   ├── query
│   └── stream (WebSocket)
├── diagnostic/     # AI diagnostic
│   ├── analyze
│   ├── translate
│   └── suggest-fix
├── environment/    # Environment management
├── mcp/            # MCP bridge
└── settings/       # User settings
```

---

## 19. Event Bus

### 19.1 Event Bus Architecture

```go
// EventBus provides publish/subscribe messaging between modules.
// This decouples modules and enables async communication.
type EventBus interface {
    // Publish publishes an event to a topic.
    Publish(topic Topic, event Event)
    
    // Subscribe subscribes to a topic.
    Subscribe(topic Topic, handler EventHandler) Subscription
    
    // Unsubscribe removes a subscription.
    Unsubscribe(sub Subscription)
    
    // Close closes the event bus.
    Close()
}

type Topic string
type EventHandler func(Event)

// Event is the base event type.
type Event struct {
    ID        string
    Topic     Topic
    Timestamp time.Time
    Data      interface{}
    Source    string // Source module name
}
```

### 19.2 Event Topics

```go
const (
    // Workflow events
    TopicWorkflowCreated   Topic = "workflow.created"
    TopicWorkflowUpdated   Topic = "workflow.updated"
    TopicWorkflowDeleted   Topic = "workflow.deleted"
    TopicWorkflowCompiled  Topic = "workflow.compiled"
    
    // Runtime events
    TopicRuntimeStarted    Topic = "runtime.started"
    TopicRuntimeCompleted  Topic = "runtime.completed"
    TopicRuntimeFailed     Topic = "runtime.failed"
    TopicRuntimeLog        Topic = "runtime.log"
    
    // Task events
    TopicTaskCreated       Topic = "task.created"
    TopicTaskStarted       Topic = "task.started"
    TopicTaskCompleted     Topic = "task.completed"
    TopicTaskFailed        Topic = "task.failed"
    TopicTaskProgress      Topic = "task.progress"
    
    // Plugin events
    TopicPluginInstalled   Topic = "plugin.installed"
    TopicPluginUninstalled Topic = "plugin.uninstalled"
    TopicPluginUpdated     Topic = "plugin.updated"
    
    // Project events
    TopicProjectCreated    Topic = "project.created"
    TopicProjectDeleted    Topic = "project.deleted"
    
    // Log events
    TopicLogEntry          Topic = "log.entry"
    TopicLogError          Topic = "log.error"
    
    // Diagnostic events
    TopicDiagnosticReady   Topic = "diagnostic.ready"
    
    // System events
    TopicSystemShutdown    Topic = "system.shutdown"
    TopicSystemConfigReload Topic = "system.config.reload"
)
```

---

## 20. Sprint Roadmap

### Phase 1: Foundation (Sprint 1-2)

| Sprint | Module | Tasks | Deliverables |
|--------|--------|-------|-------------|
| S1 | **Architecture Setup** | Restructure directory, create new modules, update go.mod | New monorepo structure |
| S1 | **Event Bus** | Implement pub/sub event system | `eventbus/` package |
| S1 | **Workflow DSL V2** | Refactor types.go, remove runtime state, add Target | Pure declaration workflow |
| S2 | **Compiler Interface** | Create Compiler interface + registry | `compiler/` package |
| S2 | **Python Generator** | Implement Python project generator | `generators/python/` |

### Phase 2: Core (Sprint 3-4)

| Sprint | Module | Tasks | Deliverables |
|--------|--------|-------|-------------|
| S3 | **Runtime** | Refactor engine/ → runtime/, implement unified executor | `runtime/` package |
| S3 | **Runtime Bundle** | Implement bundle manager, Python bundle | Bundle system |
| S4 | **Skill Manager** | Create skill interface, registry, built-in templates | `skill/` package |
| S4 | **Agent Refactor** | Agent outputs only workflow.json, no code generation | Refactored Agent |

### Phase 3: Intelligence (Sprint 5-6)

| Sprint | Module | Tasks | Deliverables |
|--------|--------|-------|-------------|
| S5 | **Log Center** | Unified log collection, storage, streaming | `logcenter/` package |
| S5 | **Diagnostic** | AI error analysis, translation, workflow linking | `diagnostic/` package |
| S6 | **Plugin Security** | Permission declaration, signature, sandbox | Enhanced plugin system |
| S6 | **Project Manager** | Filesystem-based