# Plugin System V2 — Pure Declaration

## Overview

AIStudio Plugin System V2 is a **pure declarative** plugin system. Plugin manifests contain **zero executable code** — they are pure declarations of capabilities. The Generator reads plugin manifests to know what code to generate.

```
plugin.json (v2 manifest)
    │
    ▼
Plugin Manager
    │
    ├── Registry (in-memory)
    ├── Dynamic Discovery (from Plugins/ directory)
    └── Node Type Registration → Engine.NodeRegistry
```

## Design Principles

1. **Pure Declarations** — Manifests contain only metadata, schemas, and type info. No code, no callbacks.
2. **Zero Runtime** — Plugins are not loaded into memory as code. They describe what code to generate.
3. **JSON Schema Validation** — Node configs are validated against JSON Schemas in manifests.
4. **Dynamic Discovery** — Plugins are auto-discovered from the `Plugins/` directory.
5. **Versioned** — Each manifest has a version and min schema version.

## Plugin Manifest Schema V2

### Top-Level Structure

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "AIStudio Plugin Manifest v2",
  "required": ["id", "name", "version", "min_schema_version", "kind", "nodes"]
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | yes | Unique plugin identifier (kebab-case, e.g. `yolo`) |
| `name` | string | yes | Human-readable name |
| `version` | string | yes | Semantic version (`\d+\.\d+\.\d+`) |
| `min_schema_version` | string | yes | Minimum schema version (`"2.0.0"`) |
| `kind` | string | yes | Plugin capability kind (see below) |
| `description` | string | no | Description |
| `author` | string | no | Author/organization |
| `nodes` | PluginNode[] | yes | Node type definitions |
| `runtime_bundle` | string | no | Required runtime bundle name |
| `supported_targets` | string[] | no | Supported target platforms |

### Manifest Kinds

| Kind | Value | Description |
|------|-------|-------------|
| Algorithm | `algorithm` | Algorithm implementation (e.g., YOLO, SAM) |
| Runtime | `runtime` | Runtime environment definition |
| System | `system` | System-level capability |
| Adapter | `adapter` | Third-party adapter |
| Tool | `tool` | Utility tool |

### Plugin Node Definition

```json
{
  "type": "model_trainer.yolo",
  "name": "YOLO Trainer",
  "description": "Train YOLOv8 model on custom dataset",
  "inputs": [
    { "id": "images", "name": "Images", "type": "dataset", "required": true }
  ],
  "outputs": [
    { "id": "model", "name": "Trained Model", "type": "model", "required": true }
  ],
  "config_schema": {
    "type": "object",
    "properties": {
      "model": {
        "type": "string",
        "default": "yolov8n.pt",
        "description": "Base model name"
      },
      "epochs": {
        "type": "integer",
        "default": 100,
        "minimum": 1,
        "maximum": 1000
      },
      "batch": {
        "type": "integer",
        "default": 16
      },
      "device": {
        "type": "string",
        "enum": ["cuda", "cpu", "mps"],
        "default": "cuda"
      }
    },
    "required": ["model", "epochs"]
  }
}
```

### Port Info

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | yes | Port identifier |
| `name` | string | yes | Display name |
| `type` | string | yes | Data type (image, tensor, dataset, model, etc.) |
| `required` | boolean | no | Whether connection is mandatory |

### Config Schema

Uses JSON Schema (draft-07) for node configuration validation:

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Always `"object"` |
| `properties` | object | Map of property name → SchemaProperty |
| `required` | string[] | Required property names |

### SchemaProperty

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | JSON type (string, integer, number, boolean, array) |
| `default` | any | Default value |
| `description` | string | Description |
| `minimum` | number | Minimum value |
| `maximum` | number | Maximum value |
| `enum` | string[] | Allowed string values |
| `items` | object | Array item schema (if type=array) |

---

## Plugin Interface

```go
// Manifest is the interface for reading plugin declarations.
type Manifest interface {
    ID() string
    GetManifest() *ManifestV2
}

// Registry is the interface for plugin registration and lookup.
type Registry interface {
    Register(p *Plugin) error
    Unregister(name string) error
    Get(name string) (*Plugin, bool)
    GetByID(id string) (*Plugin, bool)
    List() []*Plugin
    ListByType(pluginType PluginType) []*Plugin
    ListEnabled() []*Plugin
    UpdateEnabled(name string, enabled bool) error
    Count() int
}

// Discovery is the interface for dynamic plugin discovery.
type Discovery interface {
    Discover() ([]*Plugin, error)
}
```

## Plugin Types

```go
type PluginType string

const (
    PluginTypeVision     PluginType = "vision"
    PluginTypeNLP        PluginType = "nlp"
    PluginTypeTimeseries PluginType = "timeseries"
    PluginTypeSimulation PluginType = "simulation"
    PluginTypeMCP        PluginType = "mcp"
    PluginTypeSystem     PluginType = "system"
)
```

## Plugin Lifecycle

```
StatusNotInstalled
    │
    │ Discover (from Plugins/ directory)
    ▼
StatusInstalled
    │
    │ Enable
    ▼
StatusEnabled ◄────► StatusDisabled
    │
    │ Error
    ▼
StatusError
```

## Dynamic Discovery

Plugins are auto-discovered from the `Plugins/` directory:

```
Plugins/
├── Vision/
│   └── plugin.json     ← Manifest v2
├── NLP/
│   └── plugin.json
├── TimeSeries/
│   └── plugin.json
├── Simulation/
│   └── plugin.json
├── MCP/
│   └── plugin.json
└── System/
    └── plugin.json
```

Each subdirectory contains a `plugin.json` file with the manifest.

## Plugin Manager

```go
type Plugin struct {
    ID          string        `json:"id"`
    Name        string        `json:"name"`
    Version     string        `json:"version"`
    Author      string        `json:"author"`
    Type        PluginType    `json:"category"`
    Description string        `json:"description"`
    Status      Status        `json:"status"`
    Enabled     bool          `json:"enabled"`
    Nodes       []PluginNode  `json:"nodes"`
    Manifest    *ManifestV2   `json:"-"`
    SourceDir   string        `json:"-"`
    CreatedAt   time.Time     `json:"createdAt"`
    UpdatedAt   time.Time     `json:"updatedAt"`
}
```

### PluginSummary

A lightweight representation for list views:

```go
type PluginSummary struct {
    ID          string     `json:"id"`
    Name        string     `json:"name"`
    Version     string     `json:"version"`
    Author      string     `json:"author"`
    Type        PluginType `json:"category"`
    Description string     `json:"description"`
    Status      Status     `json:"status"`
    Enabled     bool       `json:"enabled"`
    NodeCount   int        `json:"nodeCount"`
    Kind        string     `json:"kind"`
}
```

## Node Type Registration

Discovered plugin nodes are registered in the `workflow.Engine.NodeRegistry`:

```go
// Each plugin node type is registered as a NodeDefinition
engine.Registry().Register(NodeDefinition{
    Type:        "model_trainer.yolo",
    Plugin:      "yolo",
    Name:        "YOLO Trainer",
    Description: "Train YOLOv8 model",
    Inputs:      []Port{...},
    Outputs:     []Port{...},
    Factory:     func() ExecutableNode { return &YOLONode{} },
})
```

---

## Migration from V1 to V2

### V1 (Legacy)

V1 plugins were Go plugins (`.so` files) loaded at runtime:

```go
// V1: Executable plugin with code
type PluginV1 interface {
    Name() string
    Execute(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)
}
```

### V2 (Current)

V2 plugins are pure JSON manifests:

```json
{
  "id": "yolo",
  "name": "YOLO",
  "version": "1.0.0",
  "kind": "algorithm",
  "nodes": [
    { "type": "model_trainer.yolo", ... }
  ]
}
```

### Migration Steps

1. Extract node type definitions from V1 code into `plugin.json`
2. Define input/output port schemas
3. Define config JSON Schema
4. Add runtime bundle reference
5. Remove executable code from plugin
6. Move `plugin.json` to `Plugins/<Category>/`

### Key Differences

| Aspect | V1 | V2 |
|--------|----|----|
| Code | Go `.so` plugins | Zero code (pure JSON) |
| Runtime | Loaded at runtime | Parsed at compile time |
| Schema | No schema | JSON Schema for configs |
| Ports | Implicit | Explicit typed ports |
| Discovery | Manual registration | Automatic from `Plugins/` dir |
| Bundle | Bundled with code | Declared as `runtime_bundle` ref |

---

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/plugins` | List all plugins |
| GET | `/api/plugins/:name` | Get plugin details |
| GET | `/api/plugins/nodes` | Get all plugin node types |
| PUT | `/api/plugins/:name/status` | Update plugin status |
