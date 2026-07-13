# Project Structure — Monorepo

## Overview

AIStudio V2 is organized as a **monorepo** with three main app entry points (Backend, Frontend, Launcher), shared packages, and a runtime plugins directory.

```
AIStudio/
├── Backend/                 ← Go backend server (main entry point)
├── Frontend/                ← Vue 3 + TypeScript web UI
├── Launcher/                ← Desktop launcher
├── packages/                ← Shared libraries
│   ├── generators/          ← Generator implementations (8 targets)
│   ├── bundles/             ← Runtime bundle specs
│   └── plugins/             ← Plugin schema definitions
├── Plugins/                 ← Dynamic plugin discovery directory
├── examples/                ← Example workflow.json files
├── Config/                  ← Configuration templates
├── Engine/                  ← Legacy (simplified to algorithm library)
├── Runtime/                 ← Legacy runtime scripts
├── Storage/                 ← Default storage directory
├── Scripts/                 ← Build and utility scripts
├── Tests/                   ← Integration tests
└── Docs/                    ← Documentation
```

## Root Directory

| Path | Description |
|------|-------------|
| `Backend/` | Go backend server — the core of AIStudio V2 |
| `Frontend/` | Vue 3 + TypeScript + Vite web frontend with Tauri desktop shell |
| `Launcher/` | Desktop launcher application |
| `packages/` | Shared Go packages (generators, bundles, plugin schemas) |
| `Plugins/` | Plugin manifest directory (auto-discovered) |
| `examples/` | Workflow JSON examples |
| `Config/` | Configuration files and templates |
| `Engine/` | Legacy engine directory (now simplified to algorithm library) |
| `Runtime/` | Legacy runtime scripts |
| `Storage/` | Default project and data storage |
| `Scripts/` | Build, start, and utility scripts |
| `Tests/` | Integration and end-to-end tests |
| `Docs/` | Documentation (API, architecture, ADRs) |

---

## Backend Structure

```
Backend/
├── cmd/
│   └── main.go              ← Application entry point
├── internal/
│   ├── agent/               ← Agent system (Workflow Builder)
│   │   ├── agent.go         ← Agent struct and Process flow
│   │   ├── planner.go       ← LLM + Rule-based planning
│   │   ├── executor.go      ← Step-by-step execution
│   │   ├── tool.go          ← Tool interface and implementations
│   │   ├── memory.go        ← Conversation memory
│   │   ├── context.go       ← Session context manager
│   │   ├── prompt.go        ← Prompt templates and builder
│   │   └── agent_test.go
│   ├── api/                 ← HTTP API layer (Gin)
│   │   ├── router.go        ← Route setup (legacy)
│   │   ├── handlers/        ← Request handlers
│   │   │   ├── agent.go
│   │   │   ├── auth.go
│   │   │   ├── apikey.go
│   │   │   ├── environment.go
│   │   │   ├── error_analysis.go
│   │   │   ├── health.go
│   │   │   ├── log.go
│   │   │   ├── mcp.go
│   │   │   ├── plugin.go
│   │   │   ├── profile.go
│   │   │   ├── project.go
│   │   │   ├── quota.go
│   │   │   ├── settings.go
│   │   │   ├── task.go
│   │   │   ├── user.go
│   │   │   ├── websocket.go
│   │   │   └── workflow.go
│   │   └── middleware/      ← HTTP middleware
│   ├── auth/                ← Authentication and authorization
│   ├── common/              ← Shared utilities
│   ├── compiler/            ← Multi-stage compiler
│   │   ├── compiler.go      ← Compiler interface and implementation
│   │   ├── generators.go    ← Generator interface (internal)
│   │   ├── registry.go      ← Generator registry
│   │   ├── generators/      ← Internal generator adapters
│   │   │   ├── python/      ← Python generator (internal)
│   │   │   ├── matlab/
│   │   │   ├── docker/
│   │   │   └── ros2/
│   ├── config/              ← Configuration loading (Viper)
│   ├── database/            ← Database initialization (GORM)
│   ├── diagnostic/          ← Error analysis and fix suggestions
│   ├── engine/              ← Engine (simplified — algorithm library)
│   ├── environment/         ← Environment manager
│   │   ├── manager.go       ← Manager with detection/install/repair
│   │   ├── checker.go
│   │   ├── cuda.go          ← CUDA/GPU detection
│   │   ├── dependency.go    ← Dependency checking
│   │   ├── detector.go      ← Python detection
│   │   ├── installer.go     ← Package installation
│   │   ├── python.go        ← Python info types
│   │   └── repair.go        ← Auto-repair
│   ├── eventbus/            ← Pub/sub event system
│   │   ├── eventbus.go      ← EventBus implementation
│   │   └── topics.go        ← All event topic definitions
│   ├── launcher/            ← Launcher service
│   ├── logcenter/           ← Log aggregation and querying
│   ├── mcp/                 ← MCP (Model Context Protocol) server
│   ├── plugin/              ← Plugin system V2
│   │   ├── interfaces.go    ← Manifest, Registry, Discovery interfaces
│   │   ├── models.go        ← Plugin, ManifestV2, PluginNode types
│   │   ├── manager.go       ← Plugin Manager implementation
│   │   ├── registry.go       ← Plugin Registry implementation
│   │   └── repository.go    ← Plugin repository (DB)
│   ├── project/             ← Filesystem project management
│   │   └── manager.go       ← Create, Open, List, Delete projects
│   ├── runtime/             ← Runtime execution engine
│   │   ├── runtime.go       ← Runtime interface + types
│   │   ├── executor.go      ← Local/Docker/SSH executors
│   │   ├── detector.go      ← Environment detection
│   │   ├── bundle.go        ← Bundle manager
│   │   └── bundles/         ← Bundle specs directory
│   ├── service/             ← Business logic facade
│   │   ├── service.go       ← Container + Services
│   │   ├── agent_service.go
│   │   ├── bundle_service.go
│   │   ├── environment_service.go
│   │   ├── log_service.go
│   │   ├── mcp_service.go
│   │   ├── plugin_service.go
│   │   ├── project_service.go
│   │   ├── runtime_service.go
│   │   ├── task_service.go
│   │   ├── user_service.go
│   │   └── workflow_service.go
│   ├── skill/               ← Workflow templates/skills
│   └── task/                ← Task management
├── pkg/                     ← Public/shared packages
│   ├── plugin/              ← Public plugin utilities
│   └── runtime/             ← Public runtime utilities
├── config/                  ← YAML configuration files
├── docs/                    ← Backend-specific docs
├── logs/                    ← Log output directory
├── go.mod
├── go.sum
└── README.md
```

---

## Frontend Structure

```
Frontend/
├── src/                     ← Vue 3 application source
├── src-tauri/               ← Tauri desktop shell (Rust)
├── tests/                   ← Frontend tests
├── public/                  ← Static assets
├── index.html
├── vite.config.ts
├── tsconfig.json
├── package.json
└── start-dev.bat
```

---

## Packages Structure

```
packages/
├── generators/              ← Public Generator API + implementations
│   ├── common/              ← Shared Generator interface + types
│   │   └── generator.go     ← Generator, Workflow, Node, Edge types
│   ├── python/              ← Python generator
│   │   ├── generator.go
│   │   └── templates/       ← Go embed templates
│   ├── cpp/                 ← C++ generator
│   ├── java/                ← Java generator
│   ├── matlab/              ← MATLAB generator
│   ├── ros2/                ← ROS2 generator
│   ├── docker/              ← Docker generator
│   ├── stm32/               ← STM32 generator
│   └── unity/               ← Unity generator
├── bundles/                 ← Runtime bundle specs
│   ├── yolo/
│   │   └── bundle.json
│   ├── transformer/
│   │   └── bundle.json
│   ├── ros/
│   │   └── bundle.json
│   ├── stm32/
│   │   └── bundle.json
│   └── matlab/
│       └── bundle.json
└── plugins/
    └── schema_v2.json       ← Plugin manifest JSON Schema
```

---

## Plugins Directory

```
Plugins/                     ← Dynamic plugin discovery
├── Vision/                  ← Vision plugins
│   └── plugin.json
├── NLP/                     ← NLP plugins
│   └── plugin.json
├── TimeSeries/              ← Time series plugins
│   └── plugin.json
├── Simulation/              ← Simulation plugins
│   └── plugin.json
├── MCP/                     ← MCP plugins
│   └── plugin.json
└── System/                  ← System plugins
    └── plugin.json
```

---

## Module Dependency Graph

```
                         ┌──────────┐
                         │   cmd/   │
                         │ main.go  │
                         └────┬─────┘
                              │
              ┌───────────────┼───────────────────┐
              │               │                   │
              ▼               ▼                   ▼
       ┌────────────┐  ┌────────────┐  ┌────────────────┐
       │   config   │  │  eventbus  │  │   database     │
       └────────────┘  └────────────┘  └────────────────┘
              │               │
              ▼               ▼
       ┌────────────┐  ┌────────────┐  ┌────────────┐
       │  project   │  │  workflow  │  │  logcenter │
       └────────────┘  └────────────┘  └────────────┘
              │               │               │
              ▼               ▼               ▼
       ┌────────────┐  ┌────────────┐  ┌────────────┐
       │  compiler  │  │   agent    │  │ diagnostic │
       │  registry  │  │  planner   │  │            │
       │  generators│  │  executor  │  └────────────┘
       └─────┬──────┘  │   tools    │
             │         └────────────┘
             ▼
       ┌────────────┐  ┌────────────┐  ┌────────────┐
       │  runtime   │  │  plugin    │  │environment │
       │  executor  │  │  manager   │  │  manager   │
       │  bundles   │  │  registry  │  └────────────┘
       └────────────┘  └────────────┘
             │
             ▼
       ┌────────────┐
       │  service   │
       │  (facade)  │
       └─────┬──────┘
             │
             ▼
       ┌────────────┐
       │  api/handlers│
       └────────────┘

External:
  packages/generators/ → imported by compiler
  packages/bundles/    → imported by runtime
  Plugins/             → discovered by plugin manager
```
