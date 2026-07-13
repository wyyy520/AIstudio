# Task Schema

## 1. 概述

Task Schema 定义了 AIStudio 任务系统的完整数据协议。Task 是 Workflow 的一次执行实例，记录了从提交到完成的全生命周期数据，包括状态流转、节点执行进度、日志和结果。

**设计原则**：

- 完整记录执行过程，支持断点续跑
- 实时状态推送，前端可渲染进度
- 支持任务取消、重试、优先级调度
- 日志可追溯，结果可复用

---

## 2. 任务生命周期

```
                    ┌──────────┐
                    │ waiting  │  任务已提交，等待调度
                    └────┬─────┘
                         │
                    ┌────▼─────┐
                    │ running  │  正在执行
                    └────┬─────┘
                    ┌────┼────┐
               ┌────▼──┐ │ ┌──▼────┐
               │success│ │ │failed │
               └───────┘ │ └───────┘
                         │
                    ┌────▼──────┐
                    │ cancelled │  用户取消
                    └───────────┘
```

```typescript
type TaskStatus =
  | "waiting"     // 等待调度
  | "running"     // 执行中
  | "success"     // 执行成功
  | "failed"      // 执行失败
  | "cancelled";  // 用户取消
```

### 状态转换规则

| 当前状态 | 允许转换到 | 触发条件 |
|---------|-----------|---------|
| `waiting` | `running` | 调度器分配执行资源 |
| `waiting` | `cancelled` | 用户取消 |
| `running` | `success` | 所有节点执行完成 |
| `running` | `failed` | 某节点执行失败且无重试 |
| `running` | `cancelled` | 用户取消 |

---

## 3. Task 顶层结构

```json
{
  "schema_version": "1.0.0",
  "id": "task_20260707_001",
  "workflow_id": "wf_001",
  "workflow_version": 3,
  "project_id": "proj_001",
  "name": "车辆检测 #001",
  "status": "running",
  "progress": 0.45,
  "priority": "normal",
  "inputs": {
    "n1": {
      "image": "/storage/datasets/test.jpg"
    }
  },
  "config": {
    "device": "cuda",
    "debug": false,
    "max_retries": 1
  },
  "nodes": [],
  "result": null,
  "error": null,
  "created_at": "2026-07-07T14:00:00Z",
  "started_at": "2026-07-07T14:00:01Z",
  "finished_at": null,
  "duration_ms": null,
  "created_by": "user_001",
  "tags": ["production", "batch-001"],
  "metadata": {
    "trigger": "manual",
    "source": "api"
  }
}
```

### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `schema_version` | string | 是 | Schema 版本号 |
| `id` | string | 是 | 任务唯一标识，格式 `task_{timestamp}_{seq}` |
| `workflow_id` | string | 是 | 关联的工作流 ID |
| `workflow_version` | int | 是 | 执行时的工作流版本 |
| `project_id` | string | 是 | 关联的项目 ID |
| `name` | string | 否 | 任务显示名称 |
| `status` | TaskStatus | 是 | 任务状态 |
| `progress` | float | 是 | 总进度 0.0 ~ 1.0 |
| `priority` | string | 否 | 优先级 |
| `inputs` | object | 否 | 节点输入数据（按 node_id 索引） |
| `config` | object | 否 | 执行配置 |
| `nodes` | TaskNode[] | 是 | 节点执行记录 |
| `result` | object | 否 | 最终结果 |
| `error` | TaskError | 否 | 错误信息 |
| `created_at` | datetime | 是 | 创建时间 |
| `started_at` | datetime | 否 | 开始执行时间 |
| `finished_at` | datetime | 否 | 结束时间 |
| `duration_ms` | int | 否 | 总耗时（毫秒） |
| `created_by` | string | 否 | 创建者 ID |
| `tags` | string[] | 否 | 标签 |
| `metadata` | object | 否 | 扩展元数据 |

### Priority 枚举

```typescript
type TaskPriority =
  | "low"
  | "normal"
  | "high"
  | "urgent";
```

---

## 4. TaskNode 结构（节点执行记录）

```json
{
  "node_id": "n1",
  "node_name": "YOLO 目标检测",
  "plugin": "yolo-detector",
  "status": "success",
  "progress": 1.0,
  "started_at": "2026-07-07T14:00:01Z",
  "finished_at": "2026-07-07T14:00:03Z",
  "duration_ms": 2000,
  "input_snapshot": {
    "image": "/storage/datasets/test.jpg",
    "confidence": 0.5
  },
  "output_snapshot": {
    "detections": {
      "boxes": [[100, 200, 300, 400], [50, 60, 120, 180]],
      "scores": [0.95, 0.87],
      "classes": [2, 7]
    },
    "annotated_image": "/runtime/workspace/task_001/n1_annotated.jpg"
  },
  "config_snapshot": {
    "model": "yolov8n.pt",
    "device": "cuda"
  },
  "error": null,
  "retry_count": 0,
  "metrics": {
    "cpu_percent": 35.2,
    "memory_mb": 512,
    "gpu_memory_mb": 1024,
    "gpu_utilization": 78.5,
    "inference_time_ms": 45
  }
}
```

### TaskNode 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `node_id` | string | 节点 ID |
| `node_name` | string | 节点显示名称 |
| `plugin` | string | 使用的插件 |
| `status` | NodeStatus | 节点执行状态 |
| `progress` | float | 节点进度 0.0 ~ 1.0 |
| `started_at` | datetime | 开始时间 |
| `finished_at` | datetime | 结束时间 |
| `duration_ms` | int | 耗时（毫秒） |
| `input_snapshot` | object | 输入数据快照 |
| `output_snapshot` | object | 输出数据快照 |
| `config_snapshot` | object | 配置参数快照 |
| `error` | NodeError | 错误信息 |
| `retry_count` | int | 已重试次数 |
| `metrics` | NodeMetrics | 性能指标 |

### NodeStatus 枚举

```typescript
type NodeStatus =
  | "idle"       // 未执行
  | "pending"    // 等待输入
  | "running"    // 执行中
  | "success"    // 执行成功
  | "error"      // 执行失败
  | "cancelled"  // 已取消
  | "skipped";   // 已跳过
```

---

## 5. 错误结构

### 5.1 TaskError（任务级错误）

```json
{
  "code": "WORKFLOW_EXECUTION_FAILED",
  "message": "Node 'n2' execution failed: model not found",
  "node_id": "n2",
  "stack_trace": "Traceback (most recent call last)...",
  "recoverable": true,
  "suggestion": "请检查模型文件是否存在",
  "occurred_at": "2026-07-07T14:00:05Z"
}
```

### 5.2 NodeError（节点级错误）

```json
{
  "code": "PLUGIN_EXECUTION_FAILED",
  "message": "CUDA out of memory",
  "detail": "Tried to allocate 256.00 MiB (GPU 0; 8.00 GiB total capacity)",
  "stack_trace": "File main.py, line 42...",
  "recoverable": true,
  "max_retries": 3,
  "retry_count": 1,
  "retry_after_ms": 5000
}
```

### 5.3 错误码分类

| 错误码前缀 | 类别 | 说明 |
|-----------|------|------|
| `WORKFLOW_*` | 工作流级 | 工作流校验、DAG 错误 |
| `PLUGIN_*` | 插件级 | 插件加载、执行错误 |
| `ENGINE_*` | 引擎级 | Python Engine 错误 |
| `RESOURCE_*` | 资源级 | GPU/内存/磁盘不足 |
| `TIMEOUT_*` | 超时级 | 执行超时 |
| `INPUT_*` | 输入级 | 输入数据校验失败 |

---

## 6. 日志结构

```json
{
  "logs": [
    {
      "timestamp": "2026-07-07T14:00:00.123Z",
      "level": "info",
      "node_id": "n1",
      "source": "plugin",
      "message": "Loading model yolov8n.pt...",
      "context": {
        "device": "cuda",
        "pid": 12345
      }
    },
    {
      "timestamp": "2026-07-07T14:00:02.000Z",
      "level": "info",
      "node_id": "n1",
      "source": "plugin",
      "message": "Detection completed: 2 objects found",
      "context": {
        "object_count": 2,
        "inference_ms": 45
      }
    }
  ]
}
```

### 日志级别

```typescript
type LogLevel =
  | "debug"
  | "info"
  | "warn"
  | "error"
  | "fatal";
```

### 日志来源

```typescript
type LogSource =
  | "system"
  | "engine"
  | "plugin"
  | "scheduler"
  | "workflow";
```

---

## 7. TaskResult（任务结果）

```json
{
  "summary": {
    "total_nodes": 4,
    "executed_nodes": 3,
    "successful_nodes": 3,
    "failed_nodes": 0,
    "skipped_nodes": 1
  },
  "outputs": {
    "n4": {
      "message": "Detected 2 vehicles"
    }
  },
  "artifacts": [
    {
      "node_id": "n1",
      "name": "annotated_image",
      "type": "image",
      "path": "/runtime/workspace/task_001/n1_annotated.jpg",
      "size_bytes": 245760
    }
  ],
  "metrics": {
    "total_duration_ms": 5000,
    "peak_memory_mb": 1024,
    "peak_gpu_memory_mb": 2048
  }
}
```

---

## 8. TaskContext（执行上下文）

Backend 为每个任务创建的运行时上下文：

```json
{
  "task_id": "task_20260707_001",
  "workflow_id": "wf_001",
  "project_id": "proj_001",
  "work_dir": "/runtime/workspace/task_20260707_001/",
  "shared": {
    "n1.detections": {"boxes": [], "scores": []},
    "n2.result": true
  },
  "variables": {
    "confidence_threshold": 0.5
  },
  "created_at": "2026-07-07T14:00:00Z"
}
```

---

## 9. WebSocket 实时推送协议

### 9.1 连接

```
ws://localhost:8080/ws/task/{task_id}
```

### 9.2 消息类型汇总

| type | 说明 | 触发时机 |
|------|------|---------|
| `task_status` | 任务状态变更 | 任务状态机转换时 |
| `node_status` | 节点状态变更 | 节点开始/结束执行时 |
| `node_progress` | 节点进度更新 | 长时间执行节点周期上报 |
| `node_log` | 节点日志 | 插件/引擎输出日志时 |
| `node_done` | 节点完成 | 节点执行成功 |
| `node_error` | 节点失败 | 节点执行失败 |
| `task_done` | 任务完成 | 所有节点执行完毕 |
| `task_error` | 任务失败 | 不可恢复错误 |

### 9.3 消息示例

```json
{"type": "task_status", "task_id": "task_001", "status": "running", "progress": 0.45, "timestamp": "2026-07-07T14:00:05Z"}
{"type": "node_status", "task_id": "task_001", "node_id": "n1", "status": "running", "timestamp": "2026-07-07T14:00:01Z"}
{"type": "node_progress", "task_id": "task_001", "node_id": "n1", "progress": 0.65, "message": "Processing frame 65/100", "timestamp": "2026-07-07T14:00:02Z"}
{"type": "node_log", "task_id": "task_001", "node_id": "n1", "level": "info", "message": "Model loaded, device=cuda", "timestamp": "2026-07-07T14:00:01Z"}
{"type": "node_done", "task_id": "task_001", "node_id": "n1", "status": "success", "output": {"detections": {"boxes": [[100, 200, 300, 400]], "scores": [0.95]}}, "duration_ms": 2000, "timestamp": "2026-07-07T14:00:03Z"}
{"type": "node_error", "task_id": "task_001", "node_id": "n2", "error": {"code": "PLUGIN_EXECUTION_FAILED", "message": "Condition evaluation failed", "recoverable": true}, "timestamp": "2026-07-07T14:00:05Z"}
{"type": "task_done", "task_id": "task_001", "status": "success", "result": {"summary": {"total_nodes": 4, "executed_nodes": 3, "successful_nodes": 3}}, "duration_ms": 5000, "timestamp": "2026-07-07T14:00:05Z"}
{"type": "task_error", "task_id": "task_001", "status": "failed", "error": {"code": "WORKFLOW_EXECUTION_FAILED", "message": "Node n2 execution failed", "node_id": "n2"}, "timestamp": "2026-07-07T14:00:05Z"}
```

---

## 10. 完整 JSON 示例

### 10.1 成功执行的任务

```json
{
  "schema_version": "1.0.0",
  "id": "task_20260707_001",
  "workflow_id": "wf_001",
  "workflow_version": 3,
  "project_id": "proj_001",
  "name": "车辆检测 #001",
  "status": "success",
  "progress": 1.0,
  "priority": "normal",
  "inputs": {
    "n1": {"image": "/storage/datasets/test.jpg"}
  },
  "config": {"device": "cuda", "debug": false, "max_retries": 1},
  "nodes": [
    {
      "node_id": "n1",
      "node_name": "图像输入",
      "plugin": "data-source",
      "status": "success",
      "progress": 1.0,
      "started_at": "2026-07-07T14:00:00Z",
      "finished_at": "2026-07-07T14:00:00Z",
      "duration_ms": 10,
      "input_snapshot": {},
      "output_snapshot": {"image": "/storage/datasets/test.jpg"},
      "error": null,
      "retry_count": 0
    },
    {
      "node_id": "n2",
      "node_name": "YOLO 车辆检测",
      "plugin": "yolo-detector",
      "status": "success",
      "progress": 1.0,
      "started_at": "2026-07-07T14:00:00Z",
      "finished_at": "2026-07-07T14:00:02Z",
      "duration_ms": 2000,
      "input_snapshot": {"image": "/storage/datasets/test.jpg", "confidence": 0.5},
      "output_snapshot": {
        "detections": {
          "boxes": [[100, 200, 300, 400], [50, 60, 120, 180]],
          "scores": [0.95, 0.87],
          "classes": [2, 7]
        },
        "annotated_image": "/runtime/workspace/task_001/n2_annotated.jpg"
      },
      "error": null,
      "retry_count": 0,
      "metrics": {"memory_mb": 512, "gpu_memory_mb": 1024}
    },
    {
      "node_id": "n3",
      "node_name": "是否有车辆",
      "plugin": "if-else",
      "status": "success",
      "progress": 1.0,
      "started_at": "2026-07-07T14:00:02Z",
      "finished_at": "2026-07-07T14:00:02Z",
      "duration_ms": 5,
      "input_snapshot": {"detections": {"boxes": [[100, 200, 300, 400]], "scores": [0.95]}},
      "output_snapshot": {"true": true, "false": false},
      "error": null,
      "retry_count": 0
    },
    {
      "node_id": "n4",
      "node_name": "输出检测数量",
      "plugin": "terminal",
      "status": "success",
      "progress": 1.0,
      "started_at": "2026-07-07T14:00:02Z",
      "finished_at": "2026-07-07T14:00:02Z",
      "duration_ms": 50,
      "input_snapshot": {"trigger": true},
      "output_snapshot": {"message": "Detected 2 vehicles"},
      "error": null,
      "retry_count": 0
    }
  ],
  "result": {
    "summary": {
      "total_nodes": 4,
      "executed_nodes": 4,
      "successful_nodes": 4,
      "failed_nodes": 0,
      "skipped_nodes": 0
    },
    "outputs": {
      "n4": {"message": "Detected 2 vehicles"}
    },
    "artifacts": [
      {
        "node_id": "n2",
        "name": "annotated_image",
        "type": "image",
        "path": "/runtime/workspace/task_001/n2_annotated.jpg",
        "size_bytes": 245760
      }
    ],
    "metrics": {
      "total_duration_ms": 2065,
      "peak_memory_mb": 512,
      "peak_gpu_memory_mb": 1024
    }
  },
  "error": null,
  "created_at": "2026-07-07T14:00:00Z",
  "started_at": "2026-07-07T14:00:00Z",
  "finished_at": "2026-07-07T14:00:02Z",
  "duration_ms": 2065,
  "created_by": "user_001"
}
```

### 10.2 失败执行的任务

```json
{
  "schema_version": "1.0.0",
  "id": "task_20260707_002",
  "workflow_id": "wf_001",
  "workflow_version": 3,
  "project_id": "proj_001",
  "name": "车辆检测 #002",
  "status": "failed",
  "progress": 0.5,
  "priority": "normal",
  "inputs": {
    "n1": {"image": "/storage/datasets/broken.jpg"}
  },
  "config": {"device": "cuda", "max_retries": 1},
  "nodes": [
    {
      "node_id": "n1",
      "node_name": "图像输入",
      "plugin": "data-source",
      "status": "success",
      "progress": 1.0,
      "started_at": "2026-07-07T15:00:00Z",
      "finished_at": "2026-07-07T15:00:00Z",
      "duration_ms": 10,
      "output_snapshot": {"image": "/storage/datasets/broken.jpg"}
    },
    {
      "node_id": "n2",
      "node_name": "YOLO 车辆检测",
      "plugin": "yolo-detector",
      "status": "error",
      "progress": 0.0,
      "started_at": "2026-07-07T15:00:00Z",
      "finished_at": "2026-07-07T15:00:03Z",
      "duration_ms": 3000,
      "input_snapshot": {"image": "/storage/datasets/broken.jpg", "confidence": 0.5},
      "output_snapshot": null,
      "error": {
        "code": "PLUGIN_EXECUTION_FAILED",
        "message": "CUDA out of memory",
        "detail": "Tried to allocate 256.00 MiB",
        "recoverable": true,
        "max_retries": 1,
        "retry_count": 1
      },
      "retry_count": 1
    },
    {
      "node_id": "n3",
      "node_name": "是否有车辆",
      "plugin": "if-else",
      "status": "skipped",
      "progress": 0.0
    },
    {
      "node_id": "n4",
      "node_name": "输出检测数量",
      "plugin": "terminal",
      "status": "skipped",
      "progress": 0.0
    }
  ],
  "result": null,
  "error": {
    "code": "WORKFLOW_EXECUTION_FAILED",
    "message": "Node n2 execution failed after 1 retry",
    "node_id": "n2",
    "recoverable": false,
    "suggestion": "请减少输入图像尺寸或使用更大的 GPU"
  },
  "created_at": "2026-07-07T15:00:00Z",
  "started_at": "2026-07-07T15:00:00Z",
  "finished_at": "2026-07-07T15:00:03Z",
  "duration_ms": 3010,
  "created_by": "user_001"
}
```
