# API Standard

## 1. 概述

API Standard 定义了 AIStudio Frontend 与 Backend 之间的通信规范。所有 REST API 和 WebSocket 通信必须遵循此标准。

**设计原则**：RESTful 风格、统一响应格式、版本化 API、支持分页/过滤/排序。

---

## 2. 基础信息

| 项目 | 值 |
|------|-----|
| Base URL | `http://localhost:8080/api/v1` |
| 协议 | HTTP / WebSocket |
| 数据格式 | JSON |
| 认证方式 | Bearer Token |
| 时区 | UTC (ISO 8601) |

---

## 3. 请求规范

### 通用请求头

```
Content-Type: application/json
Authorization: Bearer {token}
X-Request-ID: {uuid}
```

### 查询参数规范

| 参数 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `page` | int | 页码，从 1 开始 | `?page=1` |
| `page_size` | int | 每页条数，默认 20，最大 100 | `?page_size=20` |
| `sort` | string | 排序字段 | `?sort=created_at` |
| `order` | string | 排序方向 | `?order=desc` |
| `keyword` | string | 搜索关键词 | `?keyword=车辆` |
| `status` | string | 状态过滤 | `?status=running` |

---

## 4. 响应规范

### 通用响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "meta": {},
  "request_id": "req_abc123"
}
```

### 分页响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 42,
    "page": 1,
    "page_size": 20,
    "total_pages": 3,
    "items": []
  }
}
```

---

## 5. 错误规范

### 错误响应格式

```json
{
  "code": 1001,
  "message": "参数错误",
  "error": {
    "type": "validation_error",
    "details": [
      {"field": "name", "message": "名称不能为空"}
    ]
  },
  "request_id": "req_abc123"
}
```

### HTTP 状态码与业务错误码

| HTTP 状态码 | 业务错误码 | 说明 |
|------------|-----------|------|
| 200 | 0 | 成功 |
| 400 | 1001 | 参数错误 |
| 401 | 1002 | 未认证 |
| 403 | 1003 | 权限不足 |
| 404 | 1004 | 资源不存在 |
| 409 | 1005 | 资源冲突 |
| 422 | 1006 | 数据校验失败 |
| 500 | 2001 | 服务器内部错误 |

### 业务错误码

| 错误码 | 说明 |
|--------|------|
| 3001 | 工作流不存在 |
| 3002 | 工作流校验失败（DAG 环路、端口不兼容） |
| 3003 | 工作流版本冲突 |
| 3004 | 工作流正在运行中 |
| 4001 | 插件不存在 |
| 4002 | 插件加载失败 |
| 4003 | 插件执行失败 |
| 4004 | 插件依赖缺失 |
| 5001 | 任务不存在 |
| 5002 | 任务超时 |
| 5003 | 任务已取消 |
| 6001 | Python Engine 连接失败 |
| 6002 | 模型加载失败 |
| 6003 | GPU 显存不足 |
| 7001 | 项目不存在 |
| 7003 | 磁盘空间不足 |

---

## 6. Project API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /projects | 项目列表 |
| POST | /projects | 创建项目 |
| GET | /projects/:id | 项目详情 |
| PUT | /projects/:id | 更新项目 |
| DELETE | /projects/:id | 删除项目 |
| POST | /projects/:id/archive | 归档项目 |

**GET /projects 响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 5,
    "items": [
      {
        "id": "proj_001",
        "name": "智能交通检测",
        "description": "基于视觉的车辆检测与追踪",
        "path": "/storage/projects/proj_001",
        "status": "active",
        "workflow_count": 3,
        "dataset_count": 5,
        "model_count": 2,
        "storage_used_mb": 1024,
        "created_at": "2026-07-01T10:00:00Z",
        "updated_at": "2026-07-07T14:00:00Z"
      }
    ]
  }
}
```

**POST /projects 请求体**：

```json
{
  "name": "智能交通检测",
  "description": "基于视觉的车辆检测与追踪",
  "template": "blank"
}
```

**GET /projects/:id 响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "proj_001",
    "name": "智能交通检测",
    "description": "基于视觉的车辆检测与追踪",
    "path": "/storage/projects/proj_001",
    "settings": {
      "python_env": "python3.11",
      "cuda_device": 0,
      "work_dir": "/storage/projects/proj_001/workspace"
    },
    "stats": {
      "workflow_count": 3,
      "dataset_count": 5,
      "model_count": 2,
      "storage_used_mb": 1024
    },
    "created_at": "2026-07-01T10:00:00Z",
    "updated_at": "2026-07-07T14:00:00Z"
  }
}
```

---

## 7. Workflow API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /workflows | 工作流列表 |
| POST | /workflows | 创建工作流 |
| GET | /workflows/:id | 工作流详情（含 graph） |
| PUT | /workflows/:id | 更新工作流 |
| DELETE | /workflows/:id | 删除工作流 |
| POST | /workflows/:id/run | 运行工作流 |
| GET | /workflows/:id/status | 运行状态 |
| POST | /workflows/:id/stop | 停止运行 |
| GET | /workflows/:id/nodes/:nodeId/output | 节点输出 |

**POST /workflows/:id/run 请求体**：

```json
{
  "inputs": {
    "n1": {"image": "/storage/datasets/test.jpg"}
  },
  "config": {
    "device": "cuda",
    "debug": false
  }
}
```

**POST /workflows/:id/run 响应**：

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

**GET /workflows/:id/status 响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "task_20260707_001",
    "status": "running",
    "progress": 0.66,
    "nodes": [
      {"id": "n1", "status": "success", "duration_ms": 120},
      {"id": "n2", "status": "running", "duration_ms": 30}
    ]
  }
}
```

**PUT /workflows/:id 请求体**：

```json
{
  "name": "车辆检测工作流 v2",
  "graph": {
    "nodes": [],
    "edges": []
  }
}
```

---

## 8. Plugin API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /plugins | 插件列表 |
| GET | /plugins/:name | 插件详情 |
| POST | /plugins/install | 安装插件 |
| DELETE | /plugins/:name | 卸载插件 |
| GET | /plugins/:name/config-schema | 插件配置 Schema |
| POST | /plugins/:name/test | 测试插件执行 |
| POST | /plugins/:name/enable | 启用插件 |
| POST | /plugins/:name/disable | 禁用插件 |

**GET /plugins 查询参数**：`type` (vision/nlp/logic/system/simulation/mcp)

**POST /plugins/install 请求体**：

```json
{
  "source": "local",
  "path": "/storage/plugins/custom-plugin/"
}
```

| source | 说明 |
|--------|------|
| `local` | 本地目录 |
| `url` | 远程 URL |
| `registry` | 插件市场 |

**GET /plugins/:name/config-schema 响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "fields": [
      {
        "name": "model",
        "type": "string",
        "label": "模型文件",
        "default": "yolov8n.pt",
        "description": "YOLO 模型权重文件路径"
      },
      {
        "name": "device",
        "type": "select",
        "label": "推理设备",
        "default": "auto",
        "options": [
          {"value": "auto", "label": "自动"},
          {"value": "cpu", "label": "CPU"},
          {"value": "cuda", "label": "GPU (CUDA)"}
        ]
      }
    ]
  }
}
```

**POST /plugins/:name/test 请求体**：

```json
{
  "inputs": {
    "image": "/storage/datasets/test.jpg",
    "confidence": 0.5
  },
  "config": {
    "model": "yolov8n.pt",
    "device": "cuda"
  }
}
```

---

## 9. Task API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /tasks | 任务列表 |
| GET | /tasks/:id | 任务详情 |
| GET | /tasks/:id/logs | 任务日志 |
| POST | /tasks/:id/cancel | 取消任务 |
| POST | /tasks/:id/retry | 重试任务 |

**GET /tasks 查询参数**：`workflow_id`, `status`, `page`, `page_size`

**GET /tasks/:id 响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "task_20260707_001",
    "workflow_id": "wf_001",
    "status": "completed",
    "progress": 1.0,
    "inputs": {"n1": {"image": "/storage/datasets/test.jpg"}},
    "config": {"device": "cuda", "debug": false},
    "nodes": [
      {
        "node_id": "n1",
        "status": "success",
        "started_at": "2026-07-07T14:00:00Z",
        "finished_at": "2026-07-07T14:00:02Z",
        "duration_ms": 2000,
        "output": {"detections": {"boxes": [[100, 200, 300, 400]], "scores": [0.95]}}
      }
    ],
    "started_at": "2026-07-07T14:00:00Z",
    "finished_at": "2026-07-07T14:02:30Z",
    "duration_ms": 150000
  }
}
```

**GET /tasks/:id/logs 查询参数**：`node_id`, `level`, `since`, `page`, `page_size`

**GET /tasks/:id/logs 响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 10,
    "items": [
      {
        "timestamp": "2026-07-07T14:00:00Z",
        "node_id": "n1",
        "level": "info",
        "source": "plugin",
        "message": "Loading model yolov8n.pt..."
      }
    ]
  }
}
```

**POST /tasks/:id/retry 响应**：

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

---

## 10. Agent API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /agent/chat | 发送对话（SSE 流式） |
| GET | /agent/sessions | 对话列表 |
| POST | /agent/sessions | 创建对话 |
| GET | /agent/sessions/:id | 对话历史 |
| DELETE | /agent/sessions/:id | 删除对话 |

**POST /agent/chat 请求体**：

```json
{
  "session_id": "sess_001",
  "message": "帮我创建一个车辆检测的工作流",
  "context": {
    "project_id": "proj_001",
    "workflow_id": "wf_001"
  }
}
```

**POST /agent/chat 响应（SSE 流式）**：

```
data: {"type": "thinking", "content": "正在分析你的需求..."}

data: {"type": "text", "content": "我可以帮你创建一个基于 YOLO 的车辆检测工作流"}

data: {"type": "action", "action": "create_workflow", "params": {
  "name": "车辆检测工作流",
  "nodes": [
    {"type": "vision", "plugin": "yolo"},
    {"type": "logic", "plugin": "if"},
    {"type": "system", "plugin": "terminal"}
  ]
}}

data: {"type": "done"}
```

**消息类型**：

| type | 说明 |
|------|------|
| `thinking` | Agent 思考过程 |
| `text` | 文本回复 |
| `action` | 执行动作（创建/修改工作流） |
| `error` | 错误信息 |
| `done` | 对话结束 |

---

## 11. WebSocket 协议

### 11.1 任务实时状态

```
ws://localhost:8080/ws/task/{task_id}
```

**消息类型**：

| type | 说明 |
|------|------|
| `task_status` | 任务状态变更 |
| `node_status` | 节点状态变更 |
| `node_progress` | 节点进度更新 |
| `node_log` | 节点日志 |
| `node_done` | 节点完成 |
| `node_error` | 节点失败 |
| `task_done` | 任务完成 |
| `task_error` | 任务失败 |

**消息示例**：

```json
{"type": "task_status", "task_id": "task_001", "status": "running", "progress": 0.45, "timestamp": "2026-07-07T14:00:05Z"}
{"type": "node_status", "task_id": "task_001", "node_id": "n1", "status": "running", "timestamp": "2026-07-07T14:00:01Z"}
{"type": "node_log", "task_id": "task_001", "node_id": "n1", "level": "info", "message": "Model loaded", "timestamp": "2026-07-07T14:00:01Z"}
{"type": "node_done", "task_id": "task_001", "node_id": "n1", "status": "success", "output": {}, "duration_ms": 2000, "timestamp": "2026-07-07T14:00:03Z"}
{"type": "task_done", "task_id": "task_001", "status": "success", "result": {}, "duration_ms": 5000, "timestamp": "2026-07-07T14:00:05Z"}
```

### 11.2 Agent 流式对话

```
ws://localhost:8080/ws/agent/{session_id}
```

**客户端发送**：

```json
{"type": "message", "content": "帮我优化这个工作流"}
```

**服务端推送**：

```json
{"type": "thinking", "content": "正在分析工作流..."}
{"type": "text", "content": "建议在 YOLO 检测后添加..."}
{"type": "action", "action": "modify_node", "params": {}}
{"type": "done"}
```

---

## 12. API 版本管理

| 版本 | 说明 |
|------|------|
| `/api/v1` | 当前稳定版本 |
| `/api/v2` | 下一版本（开发中） |

**版本兼容规则**：

- 新增可选字段：向后兼容
- 删除字段：不兼容，新版本处理
- 修改字段类型：不兼容，新版本处理
- 新增端点：向后兼容
- 删除端点：不兼容

---

## 13. 限流与安全

| 策略 | 说明 |
|------|------|
| 限流 | 1000 请求/分钟/IP |
| 认证 | JWT Bearer Token |
| CORS | 仅允许 localhost:5173 |
| 请求大小限制 | 最大 50MB |
| WebSocket 心跳 | 30 秒间隔 |
