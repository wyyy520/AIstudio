# Agent 接口

## POST /agent/chat

发送对话消息，支持流式响应。

### 请求体

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

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| session_id | string | 是 | 对话会话 ID |
| message | string | 是 | 用户消息 |
| context | object | 否 | 附加上下文（当前项目/工作流） |

### 响应（流式 SSE）

```
data: {"type": "thinking", "content": "正在分析你的需求..."}

data: {"type": "text", "content": "我可以帮你创建一个基于 YOLO 的车辆检测工作流，"}

data: {"type": "text", "content": "包含以下步骤：\n1. 图像输入\n2. YOLO 目标检测\n3. 条件判断\n4. 结果输出"}

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

### 消息类型

| type | 说明 |
|------|------|
| thinking | Agent 思考过程 |
| text | 文本回复 |
| action | 执行动作（创建工作流/运行任务/修改节点等） |
| error | 错误信息 |
| done | 对话结束 |

---

## GET /agent/sessions

获取对话列表。

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "sess_001",
        "title": "创建车辆检测工作流",
        "message_count": 8,
        "project_id": "proj_001",
        "created_at": "2026-07-07T10:00:00Z",
        "updated_at": "2026-07-07T14:00:00Z"
      }
    ]
  }
}
```

---

## POST /agent/sessions

创建新对话。

### 请求体

```json
{
  "title": "新的对话",
  "project_id": "proj_001"
}
```

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "sess_002",
    "title": "新的对话",
    "project_id": "proj_001",
    "messages": []
  }
}
```

---

## GET /agent/sessions/:id

获取对话历史消息。

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "sess_001",
    "title": "创建车辆检测工作流",
    "messages": [
      {
        "id": "msg_001",
        "role": "user",
        "content": "帮我创建一个车辆检测的工作流",
        "timestamp": "2026-07-07T10:00:00Z"
      },
      {
        "id": "msg_002",
        "role": "assistant",
        "content": "我可以帮你创建一个基于 YOLO 的车辆检测工作流...",
        "actions": [
          {
            "type": "create_workflow",
            "status": "executed",
            "result": {"workflow_id": "wf_001"}
          }
        ],
        "timestamp": "2026-07-07T10:00:05Z"
      }
    ]
  }
}
```

---

## DELETE /agent/sessions/:id

删除对话。

### 响应

```json
{
  "code": 0,
  "message": "success"
}
```

---

## WebSocket: ws://localhost:8080/ws/agent/:sessionId

Agent 流式对话（WebSocket 版本，与 SSE 二选一）。

### 客户端发送

```json
{"type": "message", "content": "帮我优化这个工作流"}
```

### 服务端推送

```json
{"type": "thinking", "content": "正在分析工作流..."}
{"type": "text", "content": "建议在 YOLO 检测后添加..."}
{"type": "action", "action": "modify_node", "params": {...}}
{"type": "done"}
```
