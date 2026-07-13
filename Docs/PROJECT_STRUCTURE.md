# Project Structure

## Directory Tree

```
AIstudio/
├── .github/
│   └── workflows/
│       ├── ci.yml              # CI pipeline (lint, test, build)
│       └── release.yml         # Release workflow (binaries, GitHub release)
│
├── apps/
│   ├── backend/                # Restructured Go backend
│   │   ├── cmd/main.go         # Entry point
│   │   └── internal/           # Internal packages
│   └── desktop/                # Vue3 + Tauri desktop app (WIP)
│
├── Backend/                    # Go backend server (primary)
│   ├── cmd/main.go             # Server entry point
│   ├── config/                 # Default configuration files
│   ├── internal/               # Application internals
│   │   ├── agent/              # AI Agent (planning, execution, memory)
│   │   ├── api/                # HTTP handlers and middleware
│   │   ├── auth/               # Authentication & authorization
│   │   ├── common/             # Shared utilities
│   │   ├── config/             # Configuration loader
│   │   ├── database/           # Database layer & models
│   │   ├── engine/             # Python engine bridge
│   │   ├── environment/        # Environment detection
│   │   ├── launcher/           # Module lifecycle management
│   │   ├── mcp/                # MCP protocol implementation
│   │   ├── plugin/             # Plugin system
│   │   ├── service/            # Business service layer
│   │   ├── task/               # Task scheduling
│   │   └── workflow/           # Workflow engine
│   └── pkg/                    # Shared backend packages
│
├── packages/                   # Shared Go modules
│   ├── agent/                  # AI Agent SDK
│   ├── bundles/                # Runtime bundle definitions
│   ├── cloud/                  # Cloud deployment support
│   ├── common/                 # Common types & utilities
│   ├── compiler/               # Workflow compiler
│   ├── diagnostic/             # Error analysis engine
│   ├── environment/            # Environment detection
│   ├── event/                  # Event bus
│   ├── generators/             # Code generators
│   │   ├── common/             # Generator interface & base types
│   │   └── python/             # Python code generator
│   ├── logger/                 # Structured logging
│   ├── plugin/                 # Plugin loading & management
│   ├── plugins/                # Built-in plugins
│   ├── project/                # Project management
│   ├── runtime/                # Execution runtime
│   ├── sdk/                    # Developer SDK
│   ├── security/               # Security utilities
│   ├── skill/                  # Skill template management
│   ├── storage/                # File & data storage
│   └── workflow/               # Workflow types & validation
│
├── Frontend/                   # Vue3 + TypeScript frontend
│   ├── src/
│   │   ├── api/                # API client layer
│   │   ├── components/         # Shared components
│   │   ├── composables/        # Vue composables
│   │   ├── layouts/            # Page layouts
│   │   ├── pages/              # Page-level components
│   │   ├── router/             # Vue Router config
│   │   ├── stores/             # Pinia stores
│   │   └── types/              # TypeScript type definitions
│   └── tests/                  # Frontend tests
│
├── Engine/                     # Python AI execution engine
│   ├── server.py               # HTTP server mode
│   ├── runner.py               # Task execution mode
│   ├── dataset/                # Data loading & processing
│   ├── inference/              # Model inference
│   ├── model/                  # Model management
│   ├── runtime/                # Python runtime management
│   ├── trainer/                # Model training
│   └── vision/                 # Computer vision modules
│
├── Config/                     # Global configuration YAML files
├── docs/                       # Developer documentation
├── Docs/                       # Architecture & design documents
├── scripts/                    # Development scripts
│   ├── build/build.sh          # Build script
│   └── dev/dev.sh              # Dev server launcher
├── tests/                      # Test suites
│   ├── backend/                # Backend integration tests
│   ├── benchmark/              # Performance benchmarks
│   ├── e2e/                    # End-to-end tests
│   ├── frontend/               # Frontend tests
│   └── integration/            # Pipeline integration tests
│
├── Storage/                    # User data persistence
├── Runtime/                    # Runtime files (cache, logs, temp)
├── Plugins/                    # Plugin directory
├── Launcher/                   # Startup orchestration
├── Makefile                    # Top-level build orchestration
├── aistudio.ps1                # PowerShell management script
├── start.bat                   # Windows startup
└── start.py                    # Python launcher
```

## Package Dependencies

```
Backend cmd/main.go
├── Backend/internal/config         (configuration loading)
├── Backend/internal/database       (GORM database layer)
├── Backend/internal/eventbus       (internal event bus)
├── Backend/internal/logcenter      (log aggregation)
├── Backend/internal/project        (project CRUD)
├── Backend/internal/compiler       (workflow compilation bridge)
│   └── Backend/internal/compiler/generators/python (Python generator)
├── Backend/internal/runtime        (execution bridge)
├── Backend/internal/plugin         (plugin management)
├── Backend/internal/skill          (skill templates)
├── Backend/internal/diagnostic     (error analysis)
├── Backend/internal/auth           (JWT auth)
├── Backend/internal/agent          (AI agent)
├── Backend/internal/service        (service layer)
├── Backend/internal/mcp            (MCP protocol)
├── Backend/internal/task           (task scheduling)
└── Backend/internal/api            (HTTP handlers)

packages/ (shared modules)
├── workflow        → types, validation, I/O
├── compiler        → compilation engine, generator registry
├── generators      → Generator interface, Python/MATLAB etc.
├── project         → Project CRUD, export
├── runtime         → Execution orchestration, bundle management
├── event           → Event bus
├── environment     → Host environment detection
├── agent           → Agent SDK
├── diagnostic      → Error analysis engine
├── security        → JWT, encryption
├── plugin          → Plugin interfaces
├── skill           → Skill template engine
├── storage         → Data storage
├── logger          → Structured logging
├── common          → Shared types
├── bundles         → Runtime bundle specs
├── cloud           → Cloud deployment
└── sdk             → Developer SDK
```

## Data Flow

```
User Input (UI / API)
    │
    ▼
Workflow DSL (JSON/YAML)
    │
    ▼
Compiler ──► Plan ──► Generator ──► Project Files
    │                                    │
    ▼                                    ▼
Runtime ◄────────────────────────── Source Code
    │
    ├──► Python Engine (local/subprocess)
    ├──► Docker Container
    └──► SSH Remote
    │
    ▼
Results ──► Logs / Events / Metrics
```

## Key Architectural Decisions

| Decision | Rationale |
|----------|-----------|
| Workflow as single source of truth | All state is derived from the workflow JSON, enabling reproducibility |
| Go for backend | Strong typing, fast compilation, excellent concurrency for task scheduling |
| Python for engine | Rich AI/ML ecosystem (PyTorch, Ultralytics, etc.) |
| Plugin system via Go plugin + RPC | Extensible without modifying core |
| Event-driven architecture | Decouples modules, enables real-time UI updates |
| MCP protocol support | Interoperability with AI coding assistants |
