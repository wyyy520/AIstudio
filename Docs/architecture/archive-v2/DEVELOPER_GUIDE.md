# Developer Guide

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.21+ | Backend server |
| Node.js | 18+ | Frontend build |
| Python | 3.9+ | Runtime execution |
| Git | Any | Version control |

## Build Instructions

### Backend

```bash
# Clone the repository
git clone <repo-url>
cd AIStudio

# Install Go dependencies
cd Backend
go mod download

# Build the server
go build -o bin/aistudio-server ./cmd/main.go

# Run tests
go test ./... -v

# Run the server (development)
go run ./cmd/main.go

# Or use the start script
./aistudio.ps1
```

### Frontend

```bash
cd Frontend
npm install
npm run dev    # Development server
npm run build  # Production build
```

### Desktop (Tauri)

```bash
cd Frontend
npm install
npx tauri dev      # Development
npx tauri build    # Production build
```

## Configuration

Configuration is handled via `viper` and YAML files:

```yaml
# Backend/config/config.yaml
server:
  address: ":8080"
  cors_origins: ["http://localhost:5173"]
  rate_limit: 100
  read_timeout: 30s
  write_timeout: 30s

database:
  driver: "sqlite"
  path: "./data/aistudio.db"

storage:
  projects_dir: "./Storage/projects"

plugin:
  directory: "./Plugins"

auth:
  jwt_secret: "your-secret-key"
  token_expiry: 24h
```

## Running the Application

```bash
# Start backend
cd Backend && go run ./cmd/main.go

# Start frontend (separate terminal)
cd Frontend && npm run dev

# Or use the all-in-one script
./aistudio.ps1
```

The backend starts on `http://localhost:8080` by default. The frontend dev server starts on `http://localhost:5173`.

## Testing

### Backend Tests

```bash
cd Backend

# Run all tests
go test ./... -v

# Run specific package tests
go test ./internal/agent/... -v
go test ./internal/compiler/... -v
go test ./internal/runtime/... -v
go test ./internal/plugin/... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Frontend Tests

```bash
cd Frontend
npm run test
```

## Adding a New Generator

### 1. Create the generator package

```
packages/generators/<name>/
├── generator.go
└── templates/
    └── ...
```

### 2. Implement the Generator interface

```go
package mygenerator

import (
    "context"
    "embed"
    "text/template"

    "github.com/aistudio/backend/packages/generators/common"
)

//go:embed templates/*
var templateFS embed.FS

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
    // 1. Create output directory
    // 2. Load and execute templates with workflow data
    // 3. Write generated files
    // 4. Return GenerateResult
}

func (g *MyGenerator) Validate(wf *common.Workflow) error {
    // Validate workflow for this target
    return nil
}

func (g *MyGenerator) RuntimeRequirement(wf *common.Workflow) (*common.RuntimeRequirement, error) {
    return &common.RuntimeRequirement{
        Name:     "my-bundle",
        Version:  "1.0.0",
        Python:   ">=3.9",
        Commands: []string{"my-tool"},
    }, nil
}
```

### 3. Create the internal adapter

```
Backend/internal/compiler/generators/<name>/
└── <name>_adapter.go
```

### 4. Register in main.go

```go
// Backend/cmd/main.go
func registerGenerators(compilerEngine compiler.Compiler) {
    compilerEngine.RegisterGenerator(compilerPython.NewGenerator())
    compilerEngine.RegisterGenerator(mygenerator.NewGenerator()) // Your new generator
}
```

## Adding a New Bundle

### 1. Create bundle directory and spec

```
packages/bundles/<name>/
└── bundle.json
```

### 2. Write bundle.json

```json
{
  "name": "my-bundle",
  "version": "1.0.0",
  "python": ">=3.9",
  "packages": ["mypackage>=1.0", "another-dep"],
  "commands": ["python3", "my-tool"],
  "gpu_optional": false,
  "description": "My AI bundle"
}
```

### 3. Bundle is auto-discovered

The `LoadBundleSpecsFromDir()` function scans `packages/bundles/*/bundle.json` and loads all specs at startup.

## Plugin Development

### 1. Create plugin.json

```json
{
  "id": "my-plugin",
  "name": "My Plugin",
  "version": "1.0.0",
  "min_schema_version": "2.0.0",
  "kind": "algorithm",
  "description": "My plugin description",
  "author": "Me",
  "nodes": [
    {
      "type": "my_processor",
      "name": "My Processor",
      "description": "Processes data",
      "inputs": [
        { "id": "in", "name": "Input Data", "type": "dataset", "required": true }
      ],
      "outputs": [
        { "id": "out", "name": "Output Data", "type": "dataset", "required": true }
      ],
      "config_schema": {
        "type": "object",
        "properties": {
          "param1": { "type": "string", "default": "value" },
          "param2": { "type": "integer", "minimum": 1, "default": 10 }
        },
        "required": ["param1"]
      }
    }
  ],
  "runtime_bundle": "my-bundle",
  "supported_targets": ["python"]
}
```

### 2. Place in Plugins directory

```
Plugins/<Category>/
└── plugin.json
```

### 3. Plugin is auto-discovered at startup

The `PluginManager.DiscoverPlugins()` method scans `Plugins/` recursively.

## CI/CD Pipeline

### GitHub Actions Workflow

```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run backend tests
        run: cd Backend && go test ./... -v
      - name: Run frontend tests
        run: cd Frontend && npm ci && npm run test

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build backend
        run: cd Backend && go build -o bin/aistudio-server ./cmd/main.go
      - name: Build frontend
        run: cd Frontend && npm ci && npm run build
```

## Coding Standards

### Go

- Follow standard Go formatting (`gofmt`)
- Use interfaces for testability
- Embed `BaseGenerator` for generator implementations
- Use `sync.RWMutex` for concurrent access
- Publish events via EventBus for cross-module communication
- Log via `log.Printf` or event bus topics

### File Naming

- Use `snake_case` for Go files
- Use dot-notation for node types (`model_trainer.yolo`)
- Use kebab-case for plugin IDs (`my-plugin`)

### Import Conventions

```go
// Standard library first
import (
    "context"
    "fmt"
    "os"
)

// External dependencies next
import (
    "github.com/gin-gonic/gin"
    "github.com/spf13/viper"
)

// Internal imports last
import (
    "github.com/aistudio/backend/internal/compiler"
    "github.com/aistudio/backend/internal/eventbus"
    "github.com/aistudio/backend/internal/workflow"
)
```

## Key Development Workflows

### Adding a feature to the Compiler

1. Modify `Backend/internal/compiler/compiler.go` (Compiler interface/implementation)
2. If adding new generator capabilities, modify `generators.go`
3. Add/update `CompileEventData` in `eventbus/topics.go` if new events needed
4. Add API routes in `Backend/internal/api/router.go`
5. Update service layer in `Backend/internal/service/`

### Creating a new API endpoint

1. Add handler in `Backend/internal/api/handlers/<name>.go`
2. Add route in `Backend/internal/api/router.go`
3. Add service method in `Backend/internal/service/<name>_service.go`
4. Document in API docs

### Debugging

- Enable trace mode: `eventbus.WithTrace(true)` in `main.go`
- Check logs: `Backend/logs/` directory
- Use `envStatus` from environment manager for agent debugging
- Dry-run compilation: `opts.DryRun = true` for preview without file writes
