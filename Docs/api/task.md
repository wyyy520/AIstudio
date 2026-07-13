# 任务接口

## GET /tasks

获取任务列表。

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| workflow_id | string | 否 | 按工作流筛选 |
| status | string | 否 | 按状态筛选（pending/running/completed/failed/cancelled） |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页条数 |

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 128,
    "items": [
      {
        "id": "task_20260707_001",
        "workflow_id": "wf_001",
        "workflow_name": "车辆检测工作流",
        "status": "completed",
        "progress": 1.0,
        "started_at": "2026-07-07T14:00:00Z",
        "finished_at": "2026-07-07T14:02:30Z",
        "duration_ms": 150000
      }
    ]
  }
}
```

---

## GET /tasks/:id

获取任务详情。

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "task_20260707_001",
    "workflow_id": "wf_001",
    "status": "completed",
    "progress": 1.0,
    "inputs": {
      "n1": {"image": "/storage/datasets/test.jpg"}
    },
    "config": {
      "device": "cuda",
      "debug": false
    },
    "nodes": [
      {
        "node_id": "n1",
        "status": "success",
        "started_at": "2026-07-07T14:00:00Z",
        "finished_at": "2026-07-07T14:00:02Z",
        "duration_ms": 2000,
        "output": {
          "detections": {"boxes": [[100, 200, 300, 400]], "scores": [0.95]}
        }
      },
      {
        "node_id": "n2",
        "status": "success",
        "started_at": "2026-07-07T14:00:02Z",
        "finished_at": "2026-07-07T14:00:02Z",
        "duration_ms": 5,
        "output": {"result": true}
      }
    ],
    "started_at": "2026-07-07T14:00:00Z",
    "finished_at": "2026-07-07T14:02:30Z",
    "duration_ms": 150000
  }
}
```

---

## GET /tasks/:id/logs

获取任务执行日志。

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| node_id | string | 否 | 按节点筛选 |
| level | string | 否 | 日志级别（info/warn/error） |
| since | string | 否 | 起始时间（ISO 8601） |

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "logs": [
      {
        "timestamp": "2026-07-07T14:00:00Z",
        "node_id": "n1",
        "level": "info",
        "message": "Loading model yolov8n.pt..."
      },
      {
        "timestamp": "2026-07-07T14:00:01Z",
        "node_id": "n1",
        "level": "info",
        "message": "Model loaded, device=cuda"
      },
      {
        "timestamp": "2026-07-07T14:00:02Z",
        "node_id": "n1",
        "level": "info",
        "message": "Detection completed: 1 object found"
      },
      {
        "timestamp": "2026-07-07T14:00:02Z",
        "node_id": "n2",
        "level": "info",
        "message": "Condition evaluated: true"
      }
    ]
  }
}
```

---

## POST /tasks/:id/cancel

取消正在运行的任务。

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "task_20260707_001",
    "status": "cancelled"
  }
}
```

---

## POST /tasks/:id/retry

重试失败的任务（从失败节点开始，跳过已成功的节点）。

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "task_20260707_002",
    "status": "running",
    "resume_from": "n2"
  }
}
```
