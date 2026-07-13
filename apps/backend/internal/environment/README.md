# AI Environment Detection & Management

## Overview

The `environment` package provides AI runtime environment detection and dependency management for AIStudio. It detects Python runtime, CUDA/GPU capabilities, and Python package dependencies.

## Components

- **detector.go** — Data types for environment status (Python, CUDA, GPU, dependencies)
- **python.go** — Python version/path/pip detection
- **cuda.go** — CUDA version and GPU enumeration via nvidia-smi
- **dependency.go** — requirements.txt parsing and pip package status check
- **manager.go** — Unified `Manager` interface: `CheckEnvironment()`, `InstallDependency()`, `GetStatus()`

## API

```
GET /api/environment/status
```

Returns:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "python": {
      "version": "3.11",
      "path": "/usr/bin/python3",
      "pip": true
    },
    "cuda": {
      "cuda": "12.1",
      "gpu": [
        { "name": "NVIDIA RTX 4090", "memory": "24564 MiB" }
      ]
    },
    "dependencies": [
      { "name": "torch", "version": "2.1.0", "status": "installed" },
      { "name": "transformers", "version": "", "status": "missing" }
    ]
  }
}
```

## Usage

```go
import "github.com/aistudio/backend/internal/environment"

mgr := environment.NewManager()
status := mgr.CheckEnvironment()
mgr.InstallDependency("torch")
```
