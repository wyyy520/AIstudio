# Quickstart Guide

This guide walks through creating a YOLO object detection workflow from scratch.

## Prerequisites

Ensure you have installed:
- Go 1.25+
- Node.js 20+
- Python 3.9+
- (Optional) CUDA 11.8+ for GPU acceleration

## Step 1: Start the Services

```bash
# Development mode (auto-reload)
make dev
```

This starts:
- Backend API at `http://localhost:8081`
- Frontend UI at `http://localhost:5173`

## Step 2: Create a Project

### Using the API

```bash
# Create a new project
curl -X POST http://localhost:8081/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "yolo-training",
    "target": "python",
    "description": "YOLOv8 object detection training pipeline"
  }'
```

### Using the Frontend

1. Open `http://localhost:5173` in your browser
2. Click "New Project"
3. Enter project name: `yolo-training`
4. Select target: `Python`
5. Click "Create"

## Step 3: Build a Workflow

A YOLO training workflow consists of the following nodes:

```
[Data Loader] → [Data Split] → [Model Trainer] → [Model Evaluator] → [Model Exporter]
```

### Using the API

```bash
# Update workflow with nodes and edges
curl -X PUT http://localhost:8081/api/v1/projects/<project-id>/workflow \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "nodes": [
      {"id": "n1", "type": "data_loader", "name": "load_yolo_dataset", "position": {"x": 0, "y": 0}},
      {"id": "n2", "type": "data_split", "name": "split_dataset", "position": {"x": 300, "y": 0}},
      {"id": "n3", "type": "model_trainer", "name": "train_yolov8", "position": {"x": 600, "y": 0}},
      {"id": "n4", "type": "model_evaluator", "name": "evaluate_model", "position": {"x": 900, "y": 0}},
      {"id": "n5", "type": "model_exporter", "name": "export_onnx", "position": {"x": 1200, "y": 0}}
    ],
    "edges": [
      {"id": "e1", "source": {"node_id": "n1", "port_id": "dataset"}, "target": {"node_id": "n2", "port_id": "dataset"}},
      {"id": "e2", "source": {"node_id": "n2", "port_id": "train_dataset"}, "target": {"node_id": "n3", "port_id": "dataset"}},
      {"id": "e3", "source": {"node_id": "n3", "port_id": "model"}, "target": {"node_id": "n4", "port_id": "model"}},
      {"id": "e4", "source": {"node_id": "n4", "port_id": "model"}, "target": {"node_id": "n5", "port_id": "model"}}
    ],
    "target": "python"
  }'
```

### Using the Frontend

1. Open your project in the Workflow Editor
2. Drag nodes from the Node Panel onto the canvas:
   - Data Loader → Data Split → Model Trainer → Model Evaluator → Model Exporter
3. Connect the nodes by dragging edges between ports
4. Configure each node's properties (e.g., dataset path, model name, epochs)
5. Click "Save"

## Step 4: Compile the Workflow

```bash
curl -X POST http://localhost:8081/api/v1/projects/<project-id>/compile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>"
```

This generates a runnable Python project in the project directory:

```
project-dir/
├── src/
│   ├── load_yolo_dataset.py
│   ├── split_dataset.py
│   ├── train_yolov8.py
│   ├── evaluate_model.py
│   └── export_onnx.py
├── requirements.txt
├── pyproject.toml
├── workflow.json
└── .gitignore
```

## Step 5: Run the Workflow

```bash
curl -X POST http://localhost:8081/api/v1/projects/<project-id>/run \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>"
```

Monitor progress via the WebSocket:

```bash
# Connect to WebSocket (use wscat or similar)
wscat -c ws://localhost:8081/api/v1/ws?token=<token>
```

You will receive real-time events:
- `compile.started` / `compile.completed` / `compile.failed`
- `runtime.log` / `runtime.completed` / `runtime.failed`

## Step 6: View Results

- **Logs**: `GET /api/v1/logs?project=<project-id>`
- **Run status**: `GET /api/v1/runtime/status/<run-id>`
- **Generated files**: Browse the project directory

## Example: Full Script

```bash
#!/usr/bin/env bash
set -euo pipefail

TOKEN="your-auth-token"
BASE="http://localhost:8081/api/v1"

# 1. Login
TOKEN=$(curl -s -X POST "${BASE}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# 2. Create project
PROJECT_ID=$(curl -s -X POST "${BASE}/projects" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"name":"yolo-training","target":"python"}' | jq -r '.id')

# 3. Update workflow
curl -s -X PUT "${BASE}/projects/${PROJECT_ID}/workflow" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{
    "nodes": [
      {"id":"n1","type":"data_loader","name":"load_data","position":{"x":0,"y":0}},
      {"id":"n2","type":"model_trainer","name":"train_yolo","position":{"x":300,"y":0}}
    ],
    "edges": [
      {"id":"e1","source":{"node_id":"n1","port_id":"dataset"},"target":{"node_id":"n2","port_id":"dataset"}}
    ],
    "target":"python"
  }'

# 4. Compile
curl -s -X POST "${BASE}/projects/${PROJECT_ID}/compile" \
  -H "Authorization: Bearer ${TOKEN}"

# 5. Run
RUN_ID=$(curl -s -X POST "${BASE}/projects/${PROJECT_ID}/run" \
  -H "Authorization: Bearer ${TOKEN}" | jq -r '.runId')

echo "Workflow running. Run ID: ${RUN_ID}"
echo "Check status: ${BASE}/runtime/status/${RUN_ID}"
```
