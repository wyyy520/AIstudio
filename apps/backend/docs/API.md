# AIStudio REST API 文档

Base URL: `http://localhost:8081`

## 统一返回格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 错误码

| code | 说明 |
|------|------|
| 0 | 成功 |
| -1 | 通用错误 |

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 架构分层

```
Handler (HTTP 层)
    ↓
Service (业务逻辑层)
    ↓
Database / Task Manager / Plugin Manager (基础设施层)
```

Handler 不直接操作数据库，所有业务逻辑通过 Service 层封装。

---

## 1. 健康检查

### GET /api/health

检查服务状态，包括数据库连接。

**响应示例：**

```json
{
  "code": 0,
  "message": "AIStudio Backend is running",
  "data": {
    "status": "ok",
    "database": true
  }
}
```

---

## 2. 用户管理

### GET /api/users

获取所有用户。

### GET /api/users/:id

获取单个用户。

### POST /api/users

创建用户。

**请求体：**

```json
{
  "username": "string (必填)",
  "email": "string (必填)",
  "password": "string (必填)"
}
```

### PUT /api/users/:id

更新用户信息。所有字段可选，只更新提供的字段。

**请求体：**

```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

### DELETE /api/users/:id

删除用户。

---

## 3. 项目管理

### GET /api/projects

获取所有项目。

### GET /api/projects/:id

获取单个项目。

### POST /api/projects

创建项目。

**请求体：**

```json
{
  "name": "string (必填)",
  "description": "string",
  "ownerId": "number (必填)"
}
```

### PUT /api/projects/:id

更新项目。

**请求体：**

```json
{
  "name": "string",
  "description": "string",
  "status": "string (active/idle/running/error/archived)"
}
```

### DELETE /api/projects/:id

删除项目。

---

## 4. 工作流管理

### GET /api/workflows

获取所有工作流。支持按项目筛选：

```
GET /api/workflows?projectId=1
```

### GET /api/workflows/:id

获取单个工作流详情。返回包含 `definition`（JSON 定义）字段。

### POST /api/workflows

创建工作流。

**请求体：**

```json
{
  "projectId": "number (必填)",
  "name": "string (必填)",
  "definition": "string (JSON DAG 定义)"
}
```

### PUT /api/workflows/:id

更新工作流。

**请求体：**

```json
{
  "name": "string",
  "definition": "string",
  "status": "string (draft/active/running/completed/failed)"
}
```

### DELETE /api/workflows/:id

删除工作流。

### POST /api/workflows/:id/run

执行工作流。根据工作流 ID 加载其 Definition JSON，通过 Workflow Engine 执行。

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "uuid",
    "status": "completed",
    "nodes": [...],
    "edges": [...],
    "node_outputs": {}
  }
}
```

### GET /api/workflows/nodes

获取所有已注册的工作流节点类型列表。

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "type": "input",
      "plugin": "data-source",
      "name": "数据输入",
      "description": "数据输入节点",
      "inputs": [],
      "outputs": [{"id": "data", "label": "数据", "type": "any"}]
    }
  ]
}
```

---

## 5. 任务管理

### GET /api/tasks

获取所有任务。

### GET /api/tasks/:id

获取单个任务详情。

### POST /api/tasks

创建并提交任务到任务调度器。

**请求体：**

```json
{
  "name": "string (必填)",
  "description": "string",
  "handler": "string (必填, 如: workflow, agent)",
  "priority": "number (0=low, 1=normal, 2=high, 3=urgent, 默认1)",
  "payload": "object (任务数据)"
}
```

**响应：**

```json
{
  "code": 0,
  "message": "task submitted",
  "data": {
    "taskId": "uuid-string"
  }
}
```

### PUT /api/tasks/:id/cancel

取消任务（必须处于 pending 或 running 状态）。

### PUT /api/tasks/:id/status

更新任务状态。

**请求体：**

```json
{
  "status": "string (必填, pending/running/success/failed/cancelled)"
}
```

状态转换必须遵循状态机规则。

### DELETE /api/tasks/:id

删除任务。

### 任务状态机

```
Pending ──→ Running ──→ Success
                 ├──→ Failed
                 └──→ Cancelled
Pending ──────────────→ Cancelled
```

非法转换（如 Success → Running）会被拒绝。

### 注册的 Handler

| handler | 说明 |
|---------|------|
| `workflow` | 执行工作流 |
| `agent` | 执行 Agent 任务 |

---

## 6. 插件管理

### GET /api/plugins

获取所有已注册插件。

### GET /api/plugins/:name

获取插件详情。

### POST /api/plugins/install

安装插件（从 Plugins 目录扫描并注册）。

**请求体：**

```json
{
  "name": "string (必填, 插件名称)"
}
```

### PUT /api/plugins/:name/status

启用或禁用插件。

**请求体：**

```json
{
  "status": "string (必填, enabled 或 disabled)"
}
```

### DELETE /api/plugins/:name

卸载插件。

### POST /api/plugins/:name/execute

执行插件。

**请求体：**

```json
{
  "input": {
    "key": "value"
  }
}
```

### 插件状态

| 状态 | 说明 |
|------|------|
| `installed` | 已安装 |
| `enabled` | 已启用 |
| `disabled` | 已禁用 |
| `error` | 异常 |

---

## 7. Agent 对话

### POST /api/agent/chat

AI Agent 对话接口。接收用户消息并返回意图识别结果和待执行动作。

**请求体：**

```json
{
  "message": "string (必填, 用户消息)",
  "projectId": "string (项目ID)",
  "context": {}
}
```

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "reply": "I'll help you run the workflow. Creating a task now.",
    "actions": [
      {
        "type": "run_workflow",
        "description": "Execute workflow",
        "parameters": {
          "projectId": "1",
          "timestamp": 1700000000,
          "taskId": "uuid"
        }
      }
    ],
    "tools": ["workflow", "plugin", "task"]
  }
}
```

### 支持的意图

| 关键词 | 动作 |
|--------|------|
| run, execute, start, deploy | 创建工作流运行任务 |
| create, new, make, build | 创建新工作流 |
| status, progress, check | 查询任务状态 |
| help, what can you do | 显示帮助信息 |
| plugin, install | 列出可用插件 |

---

## 8. 日志服务

### GET /api/logs

查询系统日志。支持分页和多维度过滤。

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| level | string | 过滤级别: DEBUG, INFO, WARN, ERROR |
| source | string | 按来源过滤 |
| taskId | string | 按任务 ID 过滤 |
| keyword | string | 搜索消息内容 |
| start | string | 开始时间 (RFC3339) |
| end | string | 结束时间 (RFC3339) |
| page | int | 页码 (默认 1) |
| size | int | 每页数量 (默认 20, 最大 100) |

**请求示例：**

```
GET /api/logs?level=ERROR&page=1&size=10
GET /api/logs?keyword=task&source=task-manager
GET /api/logs?start=2024-01-01T00:00:00Z&end=2024-12-31T23:59:59Z
```

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "timestamp": "2024-01-01T12:00:00Z",
        "level": "INFO",
        "source": "task-manager",
        "message": "submitted task: abc-123",
        "taskId": "abc-123",
        "detail": ""
      }
    ],
    "total": 42,
    "page": 1,
    "size": 20
  }
}
```

---

## 完整端点列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/health | 健康检查 |
| GET | /api/users | 用户列表 |
| GET | /api/users/:id | 用户详情 |
| POST | /api/users | 创建用户 |
| PUT | /api/users/:id | 更新用户 |
| DELETE | /api/users/:id | 删除用户 |
| GET | /api/projects | 项目列表 |
| GET | /api/projects/:id | 项目详情 |
| POST | /api/projects | 创建项目 |
| PUT | /api/projects/:id | 更新项目 |
| DELETE | /api/projects/:id | 删除项目 |
| GET | /api/workflows | 工作流列表 |
| GET | /api/workflows/:id | 工作流详情 |
| POST | /api/workflows | 创建工作流 |
| PUT | /api/workflows/:id | 更新工作流 |
| DELETE | /api/workflows/:id | 删除工作流 |
| POST | /api/workflows/:id/run | 执行工作流 |
| GET | /api/workflows/nodes | 节点类型列表 |
| GET | /api/tasks | 任务列表 |
| GET | /api/tasks/:id | 任务详情 |
| POST | /api/tasks | 创建任务 |
| PUT | /api/tasks/:id/cancel | 取消任务 |
| PUT | /api/tasks/:id/status | 更新任务状态 |
| DELETE | /api/tasks/:id | 删除任务 |
| GET | /api/plugins | 插件列表 |
| GET | /api/plugins/:name | 插件详情 |
| POST | /api/plugins/install | 安装插件 |
| PUT | /api/plugins/:name/status | 更新插件状态 |
| DELETE | /api/plugins/:name | 卸载插件 |
| POST | /api/plugins/:name/execute | 执行插件 |
| POST | /api/agent/chat | Agent 对话 |
| GET | /api/logs | 查询日志 |