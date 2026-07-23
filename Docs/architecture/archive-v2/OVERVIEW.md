# AIStudio V2 — System Architecture Overview

## What is AIStudio?

**AIStudio** is a **Visual AI Engineering Platform** that transforms human intent into real, runnable engineering projects. Users describe what they want in natural language, AIStudio generates a complete, standards-based project that can be executed independently of AIStudio itself.

### Core Philosophy

1. **Workflow First** — The `workflow.json` file is the single source of truth. Every project is defined by a declarative workflow DAG (nodes + edges). No runtime state lives in the workflow.

2. **Compiler First** — The Compiler transforms workflows into real projects. It does not execute, modify, or manage projects — it only generates them. Generators produce standard projects for each target platform.

3. **Platform First** — Projects are real, standard engineering projects. Python projects work with `pip install`, MATLAB projects with `.m` files, ROS2 projects with standard packages. No lock-in, no proprietary runtimes.

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              User / Frontend                                │
│                    (Visual Canvas, CLI, API, Agent Chat)                    │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Agent (Workflow Builder Only)                        │
│  ┌─────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │
│  │Planner  │  │ Tools    │  │Executor  │  │ Memory   │  │LLM Provider  │  │
│  │(LLM+    │  │(create_  │  │(step-by- │  │(session, │  │(OpenAI,      │  │
│  │ Rules)  │  │ workflow,│  │ step)    │  │ history) │  │ Claude, ...) │  │
│  └─────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────────┘  │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │ produces workflow.json
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Workflow DSL (workflow.json)                         │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ schema_version | id | name | target | nodes[] | edges[] | variables │   │
│  │     DAG of typed nodes with typed ports connected by edges           │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Compiler (Multi-Stage)                              │
│  ┌─────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │
│  │ Plan    │──│ Validate │──│ Generate │──│ Verify   │  │ Event Bus    │  │
│  │(preview)│  │(workflow)│  │(project) │  │(output)  │  │(progress)    │  │
│  └─────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────────┘  │
│                                                                             │
│  Generator Registry:                                                        │
│  ┌────────┬────────┬────────┬────────┬────────┬────────┬────────┬────────┐ │
│  │Python  │ MATLAB  │ ROS2   │ STM32  │ Docker │ C++    │ Unity  │ Java   │ │
│  └────────┴────────┴────────┴────────┴────────┴────────┴────────┴────────┘ │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │ produces project
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                     Generated Project (Filesystem)                          │
│  project/                                                                   │
│  ├── workflow.json        ← Single Source of Truth                         │
│  ├── src/                 ← Source code (standard for target)              │
│  ├── config/              ← Configuration files                            │
│  ├── data/                ← Dataset references                             │
│  ├── models/              ← Model artifacts                                │
│  ├── tests/               ← Test suite                                     │
│  └── requirements.txt     ← or package.xml, CMakeLists.txt, etc.           │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Runtime + Environment                              │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────────┐  │
│  │ Bundle Manager   │  │ Executor         │  │ Environment Manager      │  │
│  │ (install, cache, │  │ (Local / Docker  │  │ (Python, CUDA, deps      │  │
│  │  share bundles)  │  │  / SSH)          │  │  detection, repair)      │  │
│  └──────────────────┘  └──────────────────┘  └──────────────────────────┘  │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Plugin System V2                                   │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │ Pure Declarative Manifests (plugin.json) — NO executable code        │   │
│  │ Nodes, Ports, Config Schema (JSON Schema), Runtime Bundle refs       │   │
│  │ Dynamic discovery from Plugins/ directory                             │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Module Responsibilities

| Module | Responsibility |
|--------|---------------|
| **Workflow DSL** | `workflow.json` schema, DAG types (Node, Edge, Port), file I/O, schema migration |
| **Agent** | Natural language → Workflow.json. Planner (LLM + rules), Tools, Executor |
| **Compiler** | Workflow → Project. Plan → Validate → Generate → Verify. Generator registry |
| **Generators (8)** | Target-specific project generation: Python, MATLAB, ROS2, STM32, Docker, C++, Unity, Java |
| **Runtime** | Project execution. Bundle management (install/cache/share). Local/Docker/SSH executors |
| **Environment Manager** | Python/CUDA/dependency detection. Progressive installation. GPU detection |
| **Plugin System V2** | Pure declarative manifests. Node type registration. Dynamic discovery |
| **Event Bus** | Decoupled pub/sub communication between all modules |
| **API Layer** | REST endpoints (Gin), WebSocket, middleware (auth, CORS, rate limit) |
| **Service Layer** | Thin facade between API handlers and core modules |
| **Project Manager** | Filesystem project lifecycle (create, open, delete). workflow.json management |
| **Log Center** | Unified log aggregation and querying |
| **Diagnostic** | AI-powered error analysis, translation, fix suggestions |

---

## Data Flow

```
User Intent (NL)
    │
    ▼
┌──────────┐    Natural Language     ┌──────────┐
│  Agent   │ ───────────────────────► │ Planner  │
│  Chat    │ ◄─────────────────────── │ (LLM)    │
└──────────┘    JSON Action Plan     └──────────┘
    │
    │ execute tools
    ▼
┌────────────────┐    produces     ┌──────────────────────┐
│ Tool Registry  │ ──────────────► │ workflow.json        │
│ (create_node,  │                 │ (DAG of nodes+edges) │
│  connect_nodes,│                 └──────────────────────┘
│  fill_config,  │                         │
│  validate)     │                         │ submitted to Compiler
└────────────────┘                         ▼
                                   ┌──────────────────────┐
                                   │ Compiler             │
                                   │ 1. Plan (preview)    │
                                   │ 2. Validate (wf)     │
                                   │ 3. Generate (project)│
                                   │ 4. Verify (output)   │
                                   └──────────────────────┘
                                           │
                                           ▼
                                   ┌──────────────────────┐
                                   │ Project Directory    │
                                   │ ├── workflow.json    │
                                   │ ├── src/             │
                                   │ ├── tests/           │
                                   │ └── requirements.txt │
                                   └──────────────────────┘
                                           │
                                           ▼
                                   ┌──────────────────────┐
                                   │ Runtime              │
                                   │ 1. Detect environment│
                                   │ 2. Install bundle    │
                                   │ 3. Execute project   │
                                   │ 4. Stream logs       │
                                   └──────────────────────┘
```

---

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Workflow as file-based SSoT** | `workflow.json` is human-editable, version-controllable, and decouples state from UI |
| **Compiler-first architecture** | Separates "what to build" (workflow) from "how to build it" (generator), enabling any target |
| **Generators produce real projects** | No proprietary formats — projects are standard and independently runnable |
| **Plugin system v2 (pure declarations)** | Manifests contain zero code — only type definitions and schemas. The Generator reads them |
| **Engine simplified to algorithm library** | No third-party connectors — pure algorithm execution via node factories |
| **Agent as Workflow Builder only** | Agent only produces `workflow.json`. It does no code generation |
| **No direct third-party connectors** | All integrations go through the Compiler + Plugin system |
| **Monorepo structure** | Single repository for all components: apps, packages, Backend, Plugins |

See [ADR-001](../adr/README.md) through ADR-008 for detailed architecture decisions.

---

## Module Dependency Graph

```
Backend/cmd/main.go
  ├── config
  ├── eventbus (foundation)
  ├── database
  ├── logcenter
  ├── project
  ├── workflow (DSL types)
  ├── compiler
  │   ├── workflow
  │   ├── eventbus
  │   └── generators/ (python, matlab, ros2, docker, ...)
  ├── runtime
  │   ├── eventbus
  │   └── bundles/
  ├── plugin (v2: interfaces, manager, registry)
  ├── diagnostic
  ├── auth
  ├── skill
  ├── service (thin facade)
  ├── api (Gin router + handlers)
  └── agent
      ├── planner, executor, tool, memory, prompt
      └── LLM provider

packages/
  ├── generators/common/  (public Generator interface)
  ├── generators/python/  (standalone generator implementation)
  ├── generators/cpp/
  ├── generators/java/
  ├── generators/matlab/
  ├── generators/ros2/
  ├── generators/docker/
  ├── generators/stm32/
  ├── generators/unity/
  ├── bundles/            (bundle.json specs: yolo, transformer, ros, stm32, matlab)
  └── plugins/schema_v2.json
```

---

## Technology Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.21+, Gin (HTTP), GORM (DB), Viper (config) |
| Frontend | Vue 3 + TypeScript + Vite |
| Desktop | Tauri (Rust-based desktop shell) |
| Database | SQLite / PostgreSQL (via GORM) |
| LLM | OpenAI-compatible API (OpenAI, Claude, local) |
| Container | Docker (optional, for isolated execution) |
| Event Bus | In-memory pub/sub (eventbus package) |
| Templates | Go `text/template` + `embed.FS` |
