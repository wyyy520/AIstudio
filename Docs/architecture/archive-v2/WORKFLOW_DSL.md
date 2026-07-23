# Workflow DSL Specification

## Overview

The Workflow DSL is the **Single Source of Truth** for AIStudio V2. Every project has a `workflow.json` file at its root that declares the entire engineering workflow as a directed acyclic graph (DAG) of typed nodes with typed ports connected by edges.

### Design Principles

1. **Pure Declaration** — Workflow describes WHAT, not HOW. No runtime state, no execution status.
2. **No Runtime State** — No status, progress, logs, or errors in `workflow.json`.
3. **Self-Describing** — Contains all info needed for compilation to any target.
4. **Versioned Schema** — `schema_version` for backward compatibility.
5. **Target-Aware** — Declares target platform for compilation.
6. **File-Backed** — Always stored as `workflow.json` on disk.
7. **Atomic Writes** — Writes to temp file first, then rename.
8. **Manually Editable** — Users can edit with any text editor.

---

## Schema

### Top-Level Structure

```json
{
  "schema_version": "2.0.0",
  "id": "wf-unique-id",
  "name": "Workflow Name",
  "description": "What this workflow does",
  "version": 1,
  "author": "Author Name",
  "tags": ["vision", "training"],
  "metadata": {},
  "variables": {
    "dataset_path": "Storage/datasets/coco128",
    "epochs": 100
  },
  "target": "python",
  "nodes": [],
  "edges": [],
  "created_at": "2026-07-12T00:00:00Z",
  "updated_at": "2026-07-12T00:00:00Z"
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `schema_version` | string | yes | Schema version (currently `"2.0.0"`) |
| `id` | string | yes | Unique workflow identifier |
| `name` | string | yes | Human-readable name |
| `description` | string | no | Detailed description |
| `version` | int | yes | Workflow revision number |
| `author` | string | no | Creator name |
| `tags` | string[] | no | Classification tags |
| `metadata` | object | no | Arbitrary key-value metadata |
| `variables` | object | no | Template variables for parameterization |
| `target` | Target | yes | Target platform (see below) |
| `nodes` | Node[] | yes | Workflow DAG nodes |
| `edges` | Edge[] | yes | Data flow connections |
| `created_at` | datetime | no | Creation timestamp |
| `updated_at` | datetime | no | Last update timestamp |

---

## Target Platforms

| Target | Constant | Description |
|--------|----------|-------------|
| `python` | `TargetPython` | Python 3.x project |
| `matlab` | `TargetMATLAB` | MATLAB project |
| `ros2` | `TargetROS2` | ROS2 workspace |
| `docker` | `TargetDocker` | Docker container |
| `stm32` | `TargetSTM32` | STM32 embedded project |
| `cpp` | `TargetCPP` | C++ project |
| `unity` | `TargetUnity` | Unity project |
| `java` | `TargetJava` | Java project |

---

## Node Types

### DSL Native Control Nodes (built-in)

| Type | Constant | Description |
|------|----------|-------------|
| `control.condition` | `NodeTypeCondition` | Conditional branching |
| `control.loop` | `NodeTypeLoop` | Iteration loop |
| `control.switch` | `NodeTypeSwitch` | Multi-branch switch |
| `control.retry` | `NodeTypeRetry` | Retry with backoff |

### Algorithm/Processing Nodes (may be provided by plugins)

| Type | Constant | Description |
|------|----------|-------------|
| `data_loader` | `NodeTypeDataLoader` | Load dataset |
| `data_preprocessor` | `NodeTypeDataPreprocess` | Preprocess data |
| `data_augmentation` | `NodeTypeDataAugment` | Augment data |
| `model_trainer` | `NodeTypeModelTrainer` | Train model |
| `model_evaluator` | `NodeTypeModelEvaluator` | Evaluate model |
| `model_exporter` | `NodeTypeModelExporter` | Export model |
| `model_inference` | `NodeTypeModelInference` | Run inference |
| `data_split` | `NodeTypeDataSplit` | Split dataset |
| `feature_extractor` | `NodeTypeFeatureExtract` | Extract features |
| `hyperparameter_tuning` | `NodeTypeHyperparamTune` | Tune hyperparameters |
| `visualization` | `NodeTypeVisualization` | Visualize results |
| `metric_computation` | `NodeTypeMetricCompute` | Compute metrics |
| `custom` | `NodeTypeCustom` | Custom node |

### Node Structure

```json
{
  "id": "node-unique-id",
  "type": "model_trainer",
  "name": "Train Model",
  "description": "Train YOLOv8 on custom dataset",
  "position": { "x": 400, "y": 100 },
  "size": { "width": 200, "height": 100 },
  "config": {
    "model": "yolov8n.pt",
    "epochs": 100,
    "batch": 16
  },
  "inputs": [
    { "id": "in_data", "name": "dataset", "type": "dataset", "required": true }
  ],
  "outputs": [
    { "id": "out_model", "name": "model", "type": "model", "required": true }
  ],
  "constraints": {
    "min_inputs": 1,
    "max_inputs": 5,
    "required_config": ["model", "epochs"]
  }
}
```

### Port Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | yes | Port identifier |
| `name` | string | yes | Display name |
| `type` | DataType | yes | Data type |
| `description` | string | no | Description |
| `required` | bool | no | Whether port must be connected |

### Data Types

| Type | Constant | Description |
|------|----------|-------------|
| `image` | `DataTypeImage` | Image data |
| `tensor` | `DataTypeTensor` | Tensor data |
| `dataset` | `DataTypeDataset` | Dataset reference |
| `model` | `DataTypeModel` | Model artifact |
| `text` | `DataTypeText` | Text data |
| `number` | `DataTypeNumber` | Numeric value |
| `boolean` | `DataTypeBoolean` | Boolean value |
| `json` | `DataTypeJSON` | JSON data |
| `file` | `DataTypeFile` | File reference |
| `stream` | `DataTypeStream` | Stream data |
| `any` | `DataTypeAny` | Any type |

---

## Edge Types

```json
{
  "id": "edge-unique-id",
  "source": { "node_id": "node-dataset", "port_id": "out_image" },
  "target": { "node_id": "node-train", "port_id": "in_image" }
}
```

### Edge Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | yes | Edge identifier |
| `source` | EdgeEndpoint | yes | Source node + port |
| `target` | EdgeEndpoint | yes | Target node + port |

### EdgeEndpoint

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `node_id` | string | yes | Source/target node ID |
| `port_id` | string | yes | Source/target port ID |

---

## Control Node Configs

### Condition

```json
{
  "type": "control.condition",
  "config": {
    "expression": "accuracy > 0.9",
    "true_branch": "node-deploy",
    "false_branch": "node-retrain"
  }
}
```

### Loop

```json
{
  "type": "control.loop",
  "config": {
    "iterations": 10,
    "iterator_var": "epoch",
    "break_expression": "loss < 0.01"
  }
}
```

### Switch

```json
{
  "type": "control.switch",
  "config": {
    "expression": "model_type",
    "cases": [
      { "value": "yolo", "branch_id": "node-yolo" },
      { "value": "sam", "branch_id": "node-sam" }
    ],
    "default_case": "node-generic"
  }
}
```

### Retry

```json
{
  "type": "control.retry",
  "config": {
    "max_retries": 3,
    "backoff_ms": 1000,
    "retry_on_any": false
  }
}
```

---

## Variables and Parameterization

Variables are declared at the top level and referenced in node configs using `${variable_name}` syntax:

```json
{
  "variables": {
    "dataset_path": "Storage/datasets/coco128",
    "epochs": 100,
    "batch_size": 16
  },
  "nodes": [
    {
      "id": "node-train",
      "config": {
        "epochs": "${epochs}",
        "batch": "${batch_size}",
        "data": "${dataset_path}"
      }
    }
  ]
}
```

---

## Schema Versioning

Current schema version: **2.0.0**

The `SchemaMigrator` handles automatic migration between versions:

```go
migrator.Register("1.0.0", func(wf *Workflow) error {
    wf.SchemaVersion = "2.0.0"
    return nil
})
```

---

## Example: Full workflow.json

```json
{
  "schema_version": "2.0.0",
  "id": "wf-yolo-pipeline",
  "name": "YOLOv8 Training Pipeline",
  "description": "End-to-end dataset loading and model training",
  "version": 1,
  "author": "AIStudio",
  "tags": ["vision", "yolo", "training"],
  "variables": {
    "dataset_path": "data/coco128",
    "epochs": 100,
    "batch_size": 16
  },
  "target": "python",
  "nodes": [
    {
      "id": "node-dataset",
      "type": "data_loader",
      "name": "Dataset",
      "position": { "x": 100, "y": 100 },
      "config": { "path": "${dataset_path}", "format": "yolo" },
      "outputs": [
        { "id": "out_data", "name": "data", "type": "dataset" }
      ]
    },
    {
      "id": "node-train",
      "type": "model_trainer",
      "name": "Trainer",
      "position": { "x": 400, "y": 100 },
      "config": {
        "model": "yolov8n.pt",
        "epochs": "${epochs}",
        "batch": "${batch_size}"
      },
      "inputs": [
        { "id": "in_data", "name": "data", "type": "dataset", "required": true }
      ],
      "outputs": [
        { "id": "out_model", "name": "model", "type": "model" }
      ]
    }
  ],
  "edges": [
    {
      "id": "edge-data-to-train",
      "source": { "node_id": "node-dataset", "port_id": "out_data" },
      "target": { "node_id": "node-train", "port_id": "in_data" }
    }
  ]
}
```

---

## File I/O

- `workflow.LoadFromFile(path)` — Read and parse workflow.json
- `workflow.SaveToFile(wf, path)` — Write workflow.json atomically (temp file + rename)
- `workflow.WorkflowManager` — File-based CRUD with automatic schema migration

### Project Directory Layout

```
<project-root>/
├── workflow.json           ← Single Source of Truth
├── .aistudio/
│   └── project.json        ← Project metadata (id, name, target, status)
├── src/                    ← Generated source code
├── config/                 ← Configuration files
├── data/                   ← Dataset references
├── models/                 ← Model artifacts
├── outputs/                ← Output / logs
└── tests/                  ← Test suite
```
