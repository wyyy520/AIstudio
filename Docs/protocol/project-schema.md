# Project Schema

## 1. 概述

Project Schema 定义了 AIStudio 项目的完整数据结构。Project 是所有工作流、数据集、模型、实验的顶层容器，提供了统一的组织和管理框架。

**设计原则**：

- 一个 Project 包含完整的工作空间
- 支持项目级别的环境隔离
- 支持模板化项目创建
- 支持项目导入/导出

---

## 2. Project 顶层结构

```json
{
  "schema_version": "1.0.0",
  "id": "proj_001",
  "name": "智能交通检测",
  "description": "基于视觉的车辆检测与追踪系统",
  "template": "blank",
  "path": "/storage/projects/proj_001",
  "status": "active",
  "settings": {
    "python_env": "python3.11",
    "cuda_device": 0,
    "work_dir": "/storage/projects/proj_001/workspace",
    "log_level": "info",
    "auto_save": true,
    "auto_save_interval_ms": 60000
  },
  "workflows": [
    {"id": "wf_001", "name": "车辆检测", "status": "idle"},
    {"id": "wf_002", "name": "轨迹追踪", "status": "idle"}
  ],
  "datasets": [
    {"id": "ds_001", "name": "训练集", "type": "image", "count": 1000, "size_mb": 512}
  ],
  "models": [
    {"id": "model_001", "name": "yolov8n-custom", "type": "yolo", "size_mb": 6.2, "status": "ready"}
  ],
  "plugins": [
    {"name": "yolo-detector", "version": "1.2.0", "enabled": true},
    {"name": "sumo-bridge", "version": "1.0.0", "enabled": false}
  ],
  "experiments": [
    {
      "id": "exp_001",
      "name": "YOLOv8 对比实验",
      "workflow_id": "wf_001",
      "task_count": 15,
      "best_task_id": "task_001",
      "created_at": "2026-07-01T10:00:00Z"
    }
  ],
  "environment": {
    "python_version": "3.11.4",
    "cuda_version": "12.1",
    "gpu_count": 1,
    "gpu_name": "NVIDIA RTX 4090",
    "installed_packages": {
      "torch": "2.1.0",
      "ultralytics": "8.0.120"
    }
  },
  "logs": {
    "total_count": 1250,
    "total_size_mb": 45.2,
    "retention_days": 30
  },
  "metadata": {
    "author": "user_001",
    "tags": ["vision", "traffic", "production"],
    "version": "1.0.0",
    "license": ""
  },
  "created_at": "2026-07-01T10:00:00Z",
  "updated_at": "2026-07-07T14:00:00Z",
  "last_accessed_at": "2026-07-07T14:00:00Z"
}
```

---

## 3. 字段说明

### 3.1 基本信息

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `schema_version` | string | 是 | Schema 版本号 |
| `id` | string | 是 | 项目唯一标识，格式 `proj_{nanoid}` |
| `name` | string | 是 | 项目名称 |
| `description` | string | 否 | 项目描述 |
| `template` | string | 否 | 创建时使用的模板 |
| `path` | string | 是 | 项目文件系统路径 |
| `status` | ProjectStatus | 是 | 项目状态 |
| `created_at` | datetime | 是 | 创建时间 |
| `updated_at` | datetime | 是 | 最后更新时间 |
| `last_accessed_at` | datetime | 否 | 最后访问时间 |

```typescript
type ProjectStatus =
  | "active"     // 活跃
  | "archived"   // 已归档
  | "readonly";  // 只读
```

### 3.2 ProjectTemplate 枚举

```typescript
type ProjectTemplate =
  | "blank"        // 空白项目
  | "vision"       // 视觉项目模板
  | "nlp"          // NLP 项目模板
  | "timeseries"   // 时序项目模板
  | "simulation"   // 仿真项目模板
  | "mcp";         // MCP 集成模板
```

---

## 4. Settings（项目设置）

```json
{
  "python_env": "python3.11",
  "cuda_device": 0,
  "work_dir": "/storage/projects/proj_001/workspace",
  "log_level": "info",
  "auto_save": true,
  "auto_save_interval_ms": 60000,
  "max_concurrent_tasks": 2,
  "default_timeout_ms": 300000,
  "data_retention_days": 30
}
```

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `python_env` | string | `python3.11` | Python 环境路径 |
| `cuda_device` | int | `0` | CUDA 设备编号 |
| `work_dir` | string | - | 工作目录 |
| `log_level` | string | `info` | 日志级别 |
| `auto_save` | boolean | `true` | 自动保存 |
| `auto_save_interval_ms` | int | `60000` | 自动保存间隔 |
| `max_concurrent_tasks` | int | `2` | 最大并发任务数 |
| `default_timeout_ms` | int | `300000` | 默认超时时间 |
| `data_retention_days` | int | `30` | 数据保留天数 |

---

## 5. Workflow 引用

```json
{
  "id": "wf_001",
  "name": "车辆检测",
  "description": "基于 YOLO 的车辆检测流程",
  "status": "idle",
  "version": 3,
  "node_count": 4,
  "last_task_id": "task_001",
  "last_run_status": "success",
  "created_at": "2026-07-01T10:00:00Z",
  "updated_at": "2026-07-07T14:00:00Z"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 工作流 ID |
| `name` | string | 工作流名称 |
| `description` | string | 工作流描述 |
| `status` | string | 工作流状态（idle/running） |
| `version` | int | 版本号 |
| `node_count` | int | 节点数量 |
| `last_task_id` | string | 最近一次任务 ID |
| `last_run_status` | string | 最近一次运行状态 |

---

## 6. Dataset 引用

```json
{
  "id": "ds_001",
  "name": "训练集",
  "description": "车辆检测训练数据",
  "type": "image",
  "format": "yolo",
  "count": 1000,
  "size_mb": 512,
  "path": "/storage/projects/proj_001/datasets/train",
  "labels": {
    "classes": ["car", "truck", "bus", "motorcycle"],
    "class_count": 4,
    "annotated_count": 1000
  },
  "created_at": "2026-07-01T10:00:00Z",
  "updated_at": "2026-07-05T14:00:00Z"
}
```

```typescript
type DatasetType =
  | "image"
  | "text"
  | "audio"
  | "video"
  | "timeseries"
  | "tabular"
  | "mixed";

type DatasetFormat =
  | "yolo"       // YOLO 格式
  | "coco"       // COCO 格式
  | "pascal_voc" // Pascal VOC 格式
  | "csv"        // CSV
  | "jsonl"      // JSON Lines
  | "parquet"    // Parquet
  | "custom";    // 自定义格式
```

---

## 7. Model 引用

```json
{
  "id": "model_001",
  "name": "yolov8n-custom",
  "description": "自定义训练的 YOLOv8n 模型",
  "type": "yolo",
  "framework": "pytorch",
  "size_mb": 6.2,
  "status": "ready",
  "path": "/storage/projects/proj_001/models/yolov8n-custom.pt",
  "metrics": {
    "mAP50": 0.89,
    "mAP50-95": 0.67,
    "inference_ms": 45,
    "trained_on": "ds_001",
    "epochs": 100
  },
  "created_at": "2026-07-05T14:00:00Z",
  "updated_at": "2026-07-05T14:00:00Z"
}
```

```typescript
type ModelType =
  | "yolo"
  | "rt-detr"
  | "sam"
  | "ocr"
  | "transformer"
  | "llm"
  | "lstm"
  | "cnn"
  | "custom";

type ModelFramework =
  | "pytorch"
  | "tensorflow"
  | "onnx"
  | "tensorrt"
  | "opencv";
```

---

## 8. Plugin 引用

```json
{
  "name": "yolo-detector",
  "display_name": "YOLO 目标检测",
  "version": "1.2.0",
  "type": "vision",
  "language": "python",
  "enabled": true,
  "loaded": true,
  "installed_at": "2026-07-01T10:00:00Z",
  "config": {
    "model": "yolov8n.pt",
    "device": "cuda"
  }
}
```

---

## 9. Experiment（实验）

实验是对工作流多次运行结果的对比分析。

```json
{
  "id": "exp_001",
  "name": "YOLOv8 对比实验",
  "description": "对比不同 YOLOv8 模型大小对检测精度的影响",
  "workflow_id": "wf_001",
  "task_count": 15,
  "best_task_id": "task_001",
  "best_metric": {
    "metric": "mAP50",
    "value": 0.89
  },
  "parameters": {
    "model_variants": ["yolov8n.pt", "yolov8s.pt", "yolov8m.pt"],
    "confidence_range": [0.3, 0.5, 0.7]
  },
  "status": "completed",
  "created_at": "2026-07-01T10:00:00Z",
  "completed_at": "2026-07-03T14:00:00Z"
}
```

### ExperimentTask（实验任务记录）

```json
{
  "task_id": "task_001",
  "parameters": {
    "model": "yolov8n.pt",
    "confidence": 0.5
  },
  "metrics": {
    "mAP50": 0.89,
    "mAP50-95": 0.67,
    "inference_ms": 45,
    "memory_mb": 512
  },
  "status": "success",
  "duration_ms": 5000,
  "created_at": "2026-07-01T10:00:00Z"
}
```

---

## 10. Environment（项目环境）

```json
{
  "python_version": "3.11.4",
  "python_path": "/usr/bin/python3.11",
  "cuda_version": "12.1",
  "cuda_device": 0,
  "gpu_count": 1,
  "gpu_name": "NVIDIA RTX 4090",
  "gpu_memory_mb": 24576,
  "os": "windows",
  "os_version": "11",
  "installed_packages": {
    "torch": "2.1.0",
    "torchvision": "0.16.0",
    "ultralytics": "8.0.120",
    "opencv-python": "4.8.1.78",
    "numpy": "1.26.2"
  },
  "missing_packages": [
    {"name": "sumo-rl", "required_by": "sumo-plugin"}
  ]
}
```

---

## 11. Logs（项目日志概览）

```json
{
  "total_count": 1250,
  "total_size_mb": 45.2,
  "retention_days": 30,
  "log_dir": "/storage/projects/proj_001/logs",
  "recent_errors": [
    {
      "task_id": "task_002",
      "node_id": "n2",
      "error_code": "PLUGIN_EXECUTION_FAILED",
      "message": "CUDA out of memory",
      "timestamp": "2026-07-07T15:00:03Z"
    }
  ]
}
```

---

## 12. 完整 JSON 示例

### 12.1 空白项目

```json
{
  "schema_version": "1.0.0",
  "id": "proj_new",
  "name": "新项目",
  "description": "",
  "template": "blank",
  "path": "/storage/projects/proj_new",
  "status": "active",
  "settings": {
    "python_env": "python3.11",
    "cuda_device": 0,
    "work_dir": "/storage/projects/proj_new/workspace",
    "log_level": "info",
    "auto_save": true,
    "auto_save_interval_ms": 60000,
    "max_concurrent_tasks": 2,
    "default_timeout_ms": 300000
  },
  "workflows": [],
  "datasets": [],
  "models": [],
  "plugins": [],
  "experiments": [],
  "environment": {
    "python_version": "3.11.4",
    "cuda_version": "12.1",
    "gpu_count": 1,
    "gpu_name": "NVIDIA RTX 4090"
  },
  "logs": {"total_count": 0, "total_size_mb": 0, "retention_days": 30},
  "metadata": {
    "author": "user_001",
    "tags": [],
    "version": "1.0.0"
  },
  "created_at": "2026-07-07T14:00:00Z",
  "updated_at": "2026-07-07T14:00:00Z"
}
```

### 12.2 完整项目

```json
{
  "schema_version": "1.0.0",
  "id": "proj_001",
  "name": "智能交通检测",
  "description": "基于视觉的车辆检测与追踪系统",
  "template": "vision",
  "path": "/storage/projects/proj_001",
  "status": "active",
  "settings": {
    "python_env": "python3.11",
    "cuda_device": 0,
    "work_dir": "/storage/projects/proj_001/workspace",
    "log_level": "info",
    "auto_save": true,
    "auto_save_interval_ms": 60000,
    "max_concurrent_tasks": 2,
    "default_timeout_ms": 300000
  },
  "workflows": [
    {
      "id": "wf_001",
      "name": "车辆检测",
      "description": "YOLO 车辆检测",
      "status": "idle",
      "version": 3,
      "node_count": 4,
      "last_task_id": "task_001",
      "last_run_status": "success",
      "created_at": "2026-07-01T10:00:00Z",
      "updated_at": "2026-07-07T14:00:00Z"
    },
    {
      "id": "wf_002",
      "name": "轨迹追踪",
      "description": "基于检测结果的车辆追踪",
      "status": "idle",
      "version": 1,
      "node_count": 3,
      "last_task_id": null,
      "last_run_status": null,
      "created_at": "2026-07-05T10:00:00Z",
      "updated_at": "2026-07-05T10:00:00Z"
    }
  ],
  "datasets": [
    {
      "id": "ds_001",
      "name": "车辆训练集",
      "type": "image",
      "format": "yolo",
      "count": 1000,
      "size_mb": 512,
      "path": "/storage/projects/proj_001/datasets/train",
      "labels": {
        "classes": ["car", "truck", "bus", "motorcycle"],
        "class_count": 4,
        "annotated_count": 1000
      },
      "created_at": "2026-07-01T10:00:00Z"
    },
    {
      "id": "ds_002",
      "name": "车辆测试集",
      "type": "image",
      "format": "yolo",
      "count": 200,
      "size_mb": 100,
      "path": "/storage/projects/proj_001/datasets/test",
      "labels": {
        "classes": ["car", "truck", "bus", "motorcycle"],
        "class_count": 4,
        "annotated_count": 200
      },
      "created_at": "2026-07-01T10:00:00Z"
    }
  ],
  "models": [
    {
      "id": "model_001",
      "name": "yolov8n-custom",
      "type": "yolo",
      "framework": "pytorch",
      "size_mb": 6.2,
      "status": "ready",
      "path": "/storage/projects/proj_001/models/yolov8n-custom.pt",
      "metrics": {
        "mAP50": 0.89,
        "mAP50-95": 0.67,
        "inference_ms": 45,
        "trained_on": "ds_001",
        "epochs": 100
      },
      "created_at": "2026-07-05T14:00:00Z"
    }
  ],
  "plugins": [
    {"name": "yolo-detector", "version": "1.2.0", "type": "vision", "enabled": true, "loaded": true},
    {"name": "if-else", "version": "1.0.0", "type": "logic", "enabled": true, "loaded": true},
    {"name": "terminal", "version": "1.0.0", "type": "system", "enabled": true, "loaded": true}
  ],
  "experiments": [
    {
      "id": "exp_001",
      "name": "YOLOv8 对比实验",
      "description": "对比不同模型大小的检测精度",
      "workflow_id": "wf_001",
      "task_count": 15,
      "best_task_id": "task_001",
      "best_metric": {"metric": "mAP50", "value": 0.89},
      "status": "completed",
      "created_at": "2026-07-01T10:00:00Z",
      "completed_at": "2026-07-03T14:00:00Z"
    }
  ],
  "environment": {
    "python_version": "3.11.4",
    "python_path": "C:\\Python311\\python.exe",
    "cuda_version": "12.1",
    "cuda_device": 0,
    "gpu_count": 1,
    "gpu_name": "NVIDIA GeForce RTX 4090",
    "gpu_memory_mb": 24576,
    "os": "windows",
    "os_version": "11",
    "installed_packages": {
      "torch": "2.1.0",
      "ultralytics": "8.0.120",
      "opencv-python": "4.8.1.78",
      "numpy": "1.26.2"
    },
    "missing_packages": []
  },
  "logs": {
    "total_count": 1250,
    "total_size_mb": 45.2,
    "retention_days": 30,
    "log_dir": "/storage/projects/proj_001/logs",
    "recent_errors": []
  },
  "metadata": {
    "author": "user_001",
    "tags": ["vision", "traffic", "production"],
    "version": "1.0.0"
  },
  "created_at": "2026-07-01T10:00:00Z",
  "updated_at": "2026-07-07T14:00:00Z",
  "last_accessed_at": "2026-07-07T14:00:00Z"
}
```

---

## 13. 项目模板

### 13.1 空白模板

```json
{
  "template": "blank",
  "name": "空白项目",
  "description": "从零开始创建项目",
  "workflows": [],
  "datasets": [],
  "models": [],
  "plugins": ["if-else", "terminal"]
}
```

### 13.2 视觉模板

```json
{
  "template": "vision",
  "name": "视觉项目",
  "description": "预配置的计算机视觉项目",
  "workflows": [
    {
      "name": "目标检测",
      "nodes": [
        {"type": "input", "plugin": "data-source"},
        {"type": "vision", "plugin": "yolo-detector"},
        {"type": "output", "plugin": "result-sink"}
      ]
    }
  ],
  "plugins": ["yolo-detector", "sam-segmenter", "ocr-reader", "if-else"]
}
```

### 13.3 NLP 模板

```json
{
  "template": "nlp",
  "name": "NLP 项目",
  "description": "预配置的自然语言处理项目",
  "workflows": [
    {
      "name": "文本分类",
      "nodes": [
        {"type": "input", "plugin": "data-source"},
        {"type": "nlp", "plugin": "transformer"},
        {"type": "output", "plugin": "result-sink"}
      ]
    }
  ],
  "plugins": ["transformer", "llm-chat", "if-else"]
}
```

### 13.4 仿真模板

```json
{
  "template": "simulation",
  "name": "仿真项目",
  "description": "预配置的仿真集成项目",
  "workflows": [
    {
      "name": "交通仿真",
      "nodes": [
        {"type": "input", "plugin": "data-source"},
        {"type": "vision", "plugin": "yolo-detector"},
        {"type": "mcp", "plugin": "sumo-bridge"},
        {"type": "output", "plugin": "result-sink"}
      ]
    }
  ],
  "plugins": ["yolo-detector", "sumo-bridge", "if-else"]
}
```
