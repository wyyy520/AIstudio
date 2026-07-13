# 工作流接口

## GET /workflows

获取工作流列表。

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| project_id | string | 否 | 按项目筛选 |
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 42,
    "items": [
      {
        "id": "wf_001",
        "name": "车辆检测工作流",
        "project_id": "proj_001",
        "status": "idle",
        "version": 3,
        "created_at": "2026-07-01T10:00:00Z",
        "updated_at": "2026-07-07T14:00:00Z"
      }
    ]
  }
}
```

---

## POST /workflows

创建工作流。

### 请求体

```json
{
  "name": "车辆检测工作流",
  "project_id": "proj_001",
  "description": "基于 YOLO 的车辆检测流程"
}
```

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "wf_002",
    "name": "车辆检测工作流",
    "project_id": "proj_001",
    "status": "idle",
    "version": 1
  }
}
```

---

## GET /workflows/:id

获取工作流详情（含完整图数据）。

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "wf_001",
    "name": "车辆检测工作流",
    "project_id": "proj_001",
    "status": "idle",
    "version": 3,
    "graph": {
      "nodes": [
        {
          "id": "n1",
          "type": "vision",
          "plugin": "yolo",
          "position": {"x": 100, "y": 200},
          "config": {
            "model": "yolov8n.pt",
            "confidence": 0.5
          },
          "ports": {
            "inputs": [
              {"name": "image", "type": "image", "required": true}
            ],
            "outputs": [
              {"name": "detections", "type": "json"}
            ]
          }
        },
        {
          "id": "n2",
          "type": "logic",
          "plugin": "if",
          "position": {"x": 400, "y": 200},
          "config": {
            "condition": "len(detections.boxes) > 0"
          }
        }
      ],
      "edges": [
        {
          "id": "e1",
          "from": "n1",
          "to": "n2",
          "from_port": "detections",
          "to_port": "input"
        }
      ]
    }
  }
}
```

---

## PUT /workflows/:id

更新工作流（保存编辑）。

### 请求体

```json
{
  "name": "车辆检测工作流 v2",
  "graph": {
    "nodes": [...],
    "edges": [...]
  }
}
```

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "wf_001",
    "version": 4
  }
}
```

---

## POST /workflows/:id/run

运行工作流。

### 请求体

```json
{
  "inputs": {
    "n1": {
      "image": "/storage/datasets/test.jpg"
    }
  },
  "config": {
    "device": "cuda",
    "debug": false
  }
}
```

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "task_20260707_001",
    "status": "running"
  }
}
```

---

## GET /workflows/:id/status

获取工作流运行状态。

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "task_20260707_001",
    "status": "running",
    "progress": 0.66,
    "nodes": [
      {"id": "n1", "status": "success", "duration": 120},
      {"id": "n2", "status": "running", "duration": 30}
    ]
  }
}
```

---

## POST /workflows/:id/stop

停止正在运行的工作流。

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "task_20260707_001",
    "status": "cancelled"
  }
}
```

---

## GET /workflows/:id/nodes/:nodeId/output

获取某个节点的输出数据。

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "node_id": "n1",
    "status": "success",
    "output": {
      "detections": {
        "boxes": [[100, 200, 300, 400], [50, 60, 120, 180]],
        "scores": [0.95, 0.87],
        "classes": [2, 7]
      }
    },
    "metrics": {
      "duration_ms": 120,
      "memory_mb": 256
    }
  }
}
```

---

## WebSocket: ws://localhost:8080/ws/task/:taskId

实时推送任务执行状态。

### 消息格式

```json
// 节点状态变更
{"type": "node_status", "node_id": "n1", "status": "running"}

// 节点日志
{"type": "node_log", "node_id": "n1", "message": "Loading model yolov8n.pt..."}

// 节点完成
{"type": "node_status", "node_id": "n1", "status": "success", "output": {...}}

// 工作流完成
{"type": "workflow_done", "status": "success"}

// 错误
{"type": "error", "node_id": "n2", "message": "Condition evaluation failed"}
```
