# API Interface - 实际实现对照表

> 本文档记录 Frontend 实际调用的 API 路径与 Backend 实际注册的路由的对照关系。
> 用于排查模块间接口不一致问题。

## 1. 基础信息

| 项目 | Docs/protocol 设计 | Frontend 实际 | Backend 实际 |
|------|-------------------|--------------|-------------|
| Base URL | `http://localhost:8080/api/v1` | `http://127.0.0.1:8081` (request.ts 硬编码) | `http://0.0.0.0:8081` |
| API 前缀 | `/api/v1` | `/api` (无 v1) | `/api` (无 v1) |
| 端口 | 8080 | 8081 | 8081 |
| WebSocket | `ws://localhost:8080/ws/task/:taskId` | `ws://localhost:8081/api/ws` | `/api/ws` (同端口) |

**问题**: Docs 设计的 `/api/v1` 前缀与实际 `/api` 不一致。Docs 设计的端口 8080 与实际 8081 不一致。

---

## 2. 响应格式对照

### 2.1 成功响应

```json
// Frontend 期望 (client.ts unwrap 逻辑)
{ "code": 0, "message": "success", "data": <T> }

// Backend 实际返回
{ "code": 0, "message": "success", "data": <T> }
```

**一致**: Frontend `client.ts:32` 检查 `res.code !== 0` 判断错误，Backend 统一返回 `code: 0` 表示成功。

### 2.2 错误响应

```json
// Frontend 期望
{ "code": -1, "message": "错误描述" }

// Backend common.RespondError() 实际返回
{ "code": -1, "message": "错误描述", "data": { "code": "BAD_REQUEST", "module": "...", "message": "...", "solution": "..." } }

// Backend 内联 handler 实际返回
{ "code": -1, "message": "错误描述" }
```

**问题**: Backend 有两种错误返回方式。`common.RespondError()` 在 `data` 中嵌套了 `APIError` 结构体，但 Frontend `client.ts` 只读取顶层 `message`，忽略了 `data` 中的详细错误信息。Docs/protocol 定义的数字错误码 (1001, 1002...) 未被实现。

---

## 3. 各模块 API 路由对照

### 3.1 Auth 模块

| 功能 | Frontend 路径 | Backend 路径 | 状态 |
|------|-------------|-------------|------|
| 登录 | `POST /api/auth/login` | `POST /api/auth/login` | 一致 |
| 注册 | `POST /api/auth/register` | `POST /api/auth/register` | 一致 |
| 刷新 Token | `POST /api/auth/refresh` | `POST /api/auth/refresh` | 一致 |
| 登出 | `POST /api/auth/logout` | `POST /api/auth/logout` | 一致 |
| 获取用户信息 | `GET /api/user/profile` | `GET /api/user/profile` | 一致 |
| 更新用户信息 | `PUT /api/user/profile` | `PUT /api/user/profile` | 一致 |

### 3.2 Project 模块

| 功能 | Frontend 路径 | Backend 路径 | 状态 |
|------|-------------|-------------|------|
| 项目列表 | `GET /api/projects` | `GET /api/projects` | 一致 |
| 项目详情 | `GET /api/projects/:id` | `GET /api/projects/:id` | 一致 |
| 创建项目 | `POST /api/projects` | `POST /api/projects` | 一致 |
| 更新项目 | `PUT /api/projects/:id` | `PUT /api/projects/:id` | 一致 |
| 删除项目 | `DELETE /api/projects/:id` | `DELETE /api/projects/:id` | 一致 |

**数据结构问题**:
- Backend Project JSON: `{"id": 1, "name": "...", "ownerId": 1, "createdAt": "...", "updatedAt": "..."}`
- Frontend ApiProject: `id: number` (匹配), 但 `createdAt/updatedAt` (camelCase) 与 Backend 一致
- Frontend domain Project: `created_at/updated_at` (snake_case) 与 Backend 不一致

### 3.3 Workflow 模块

| 功能 | Frontend 路径 | Backend 路径 | 状态 |
|------|-------------|-------------|------|
| 工作流列表 | `GET /api/workflows` | `GET /api/workflows` | 一致 |
| 工作流详情 | `GET /api/workflows/:id` | `GET /api/workflows/:id` | 一致 |
| 创建工作流 | `POST /api/workflows` | `POST /api/workflows` | 一致 |
| 更新工作流 | `PUT /api/workflows/:id` | `PUT /api/workflows/:id` | 一致 |
| 删除工作流 | `DELETE /api/workflows/:id` | `DELETE /api/workflows/:id` | 一致 |
| 运行工作流 | `POST /api/workflows/:id/run` | `POST /api/workflows/:id/run` | 一致 |
| 节点类型列表 | `GET /api/workflows/nodes` | `GET /api/workflows/nodes` | 一致 |

**缺失的 Backend 路由** (Docs/protocol 中定义但未实现):
- `GET /api/workflows/:id/status` - 工作流运行状态
- `POST /api/workflows/:id/stop` - 停止运行
- `GET /api/workflows/:id/nodes/:nodeId/output` - 节点输出

**数据结构问题**:
- Backend Workflow JSON: `{"id": 1, "projectId": 1, "name": "...", "definition": "...", "status": "draft", "createdAt": "...", "updatedAt": "..."}`
- Frontend ApiWorkflow: `id: number`, `projectId: number` (匹配)
- Frontend domain Workflow: `id: string` (类型不匹配), `created_at/updated_at` (snake_case 不匹配)

### 3.4 Task 模块

| 功能 | Frontend 路径 | Backend 路径 | 状态 |
|------|-------------|-------------|------|
| 任务列表 | `GET /api/tasks` | `GET /api/tasks` | 一致 |
| 任务详情 | `GET /api/tasks/:id` | `GET /api/tasks/:id` | 一致 |
| 创建任务 | `POST /api/tasks` | `POST /api/tasks` | 一致 |
| 取消任务 | `PUT /api/tasks/:id/cancel` | `PUT /api/tasks/:id/cancel` | 一致 |
| 删除任务 | `DELETE /api/tasks/:id` | `DELETE /api/tasks/:id` | 一致 |
| 任务状态 | `GET /api/task/:id/status` | `GET /api/task/:id/status` | 一致 |
| 更新状态 | - | `PUT /api/tasks/:id/status` | Backend 独有 |
| 创建任务(v2) | - | `POST /api/task/create` | Backend 独有 |

**数据结构问题**:
- Backend Task JSON (DB model): `{"task_id": "...", "created_at": "...", "updated_at": "...", "start_time": "...", "end_time": "..."}`
- Backend Task JSON (domain): `{"task_id": "...", "created_at": "...", "updated_at": "...", "start_time": "...", "end_time": "..."}`
- Frontend ApiTask: `id: string`, `createdAt: string`, `updatedAt: string` (camelCase)
- **字段名不匹配**: Backend 返回 `task_id`，Frontend 期望 `id`
- **字段名不匹配**: Backend 返回 `created_at`，Frontend 期望 `createdAt`
- **字段名不匹配**: Backend 返回 `start_time`/`end_time`，Frontend 期望 `startedAt`/`completedAt`
- **状态值不匹配**: Backend 默认 `waiting`，Frontend `mapTaskStatus()` 未处理此值

**缺失的 Backend 路由**:
- `GET /api/tasks/:taskId/logs` - 按任务获取日志 (Frontend `fetchTaskLogs()` 调用)
- `POST /api/tasks/:id/retry` - 重试任务

### 3.5 Plugin 模块

| 功能 | Frontend 路径 | Backend 路径 | 状态 |
|------|-------------|-------------|------|
| 插件列表 | `GET /api/plugins` | `GET /api/plugins` | 一致 |
| 插件详情 | `GET /api/plugins/:name` | `GET /api/plugins/:name` | 一致 |
| 安装插件 | `POST /api/plugins/install` | `POST /api/plugins/install` | 一致 |
| 卸载插件 | `DELETE /api/plugins/:name` | `DELETE /api/plugins/:name` | 一致 |
| 更新插件 | `PUT /api/plugins/:name/update` | **不存在** | **断点** |
| 更新状态 | `PUT /api/plugins/:name/status` | `PUT /api/plugins/:name/status` | 一致 |
| 执行插件 | `POST /api/plugins/:name/execute` | `POST /api/plugins/:name/execute` | 一致 |

**Backend 独有路由** (Frontend 未调用):
- `POST /api/plugin/install` - 旧版安装接口
- `POST /api/plugin/remove` - 旧版移除接口
- `GET /api/plugin/:id` - 旧版详情接口
- `GET /api/plugins/market` - 插件市场列表
- `POST /api/plugins/market/install` - 从市场安装

**断点**: Frontend `updatePlugin()` 调用 `PUT /api/plugins/${name}/update`，但 Backend 没有此路由。

### 3.6 Agent 模块

| 功能 | Frontend 路径 | Backend 路径 | 状态 |
|------|-------------|-------------|------|
| 发送对话 | `POST /api/agent/chat` | `POST /api/agent/chat` | 一致 |
| 仅规划 | `POST /api/agent/plan` | `POST /api/agent/plan` | 一致 |
| 生成工作流 | - | `POST /api/agent/generate-workflow` | Backend 独有 |
| 生成并运行 | - | `POST /api/agent/generate-and-run` | Backend 独有 |

### 3.7 Log 模块

| 功能 | Frontend 路径 | Backend 路径 | 状态 |
|------|-------------|-------------|------|
| 查询日志 | `GET /api/logs` | `GET /api/logs` | 一致 |
| 任务日志 | `GET /api/tasks/:taskId/logs` | **不存在** | **断点** |

**断点**: Frontend `fetchTaskLogs()` 调用 `GET /api/tasks/:taskId/logs`，但 Backend 只有全局 `GET /api/logs`，不支持按 taskId 过滤。

### 3.8 Error Analysis 模块

| 功能 | Frontend 路径 | Backend 路径 | 状态 |
|------|-------------|-------------|------|
| 分析错误 | `POST /api/error/analyze` | **不存在** | **断点** |
| 修复错误 | `POST /api/error/repair` | **不存在** | **断点** |
| 获取分析结果 | `GET /api/error/analysis/:taskId` | **不存在** | **断点** |
| 获取修复状态 | `GET /api/error/fix/:fixId/status` | **不存在** | **断点** |

**断点**: 整个 Error Analysis 模块的 Backend API 未实现。

### 3.9 Settings 模块

| 功能 | Frontend 路径 | Backend 路径 | 状态 |
|------|-------------|-------------|------|
| 获取设置 | `GET /api/settings` | **不存在** | **断点** |
| 更新设置 | `PUT /api/settings` | **不存在** | **断点** |
| 引擎配置 | `GET /api/settings/engine` | **不存在** | **断点** |
| 更新引擎配置 | `PUT /api/settings/engine` | **不存在** | **断点** |
| 测试引擎连接 | `POST /api/settings/engine/test` | **不存在** | **断点** |

**断点**: 整个 Settings 模块的 Backend API 未实现。

### 3.10 Health 模块

| 功能 | Frontend 路径 | Backend 路径 | 状态 |
|------|-------------|-------------|------|
| 健康检查 | `GET /api/health` | `GET /api/health` | 一致 |

### 3.11 WebSocket

| 功能 | Frontend | Backend | 状态 |
|------|---------|---------|------|
| 连接 | `ws://localhost:8081/api/ws` | `GET /api/ws` (upgrade) | 一致 |
| 事件类型 | `task_status, task_log, task_progress, task_error, task_complete` | `task_status, task_progress, task_complete, task_error` | **缺少 task_log** |

**问题**: Frontend 定义了 `task_log` 事件类型，但 Backend 不发送此事件类型。Backend 的 `mapTaskEvent()` 只处理 `task.created`, `task.started`, `task.progress`, `task.completed`, `task.failed`, `task.cancelled`。

---

## 4. 数据结构字段名对照

### 4.1 Project

| 字段 | Backend DB Model (json tag) | Frontend ApiProject | Frontend domain Project | 一致? |
|------|---------------------------|--------------------|-----------------------|------|
| ID | `id` (uint) | `id: number` | `id: string` | 部分 |
| Name | `name` | `name` | `name` | 一致 |
| Description | `description` | `description` | `description` | 一致 |
| OwnerID | `ownerId` | `ownerId` | - | 一致 |
| Status | `status` | `status` | `status` | 一致 |
| CreatedAt | `createdAt` | `createdAt` | `created_at` | **不一致** |
| UpdatedAt | `updatedAt` | `updatedAt` | `updated_at` | **不一致** |

### 4.2 Workflow

| 字段 | Backend DB Model (json tag) | Frontend ApiWorkflow | Frontend domain Workflow | 一致? |
|------|---------------------------|--------------------|-----------------------|------|
| ID | `id` (uint) | `id: number` | `id: string` | 部分 |
| ProjectID | `projectId` | `projectId: number` | - | 一致 |
| Name | `name` | `name` | `name` | 一致 |
| Definition | `definition` (string) | `definition: string` | `nodes/edges` (parsed) | 不同层 |
| Status | `status` | `status` | `status` | 一致 |
| CreatedAt | `createdAt` | `createdAt` | `created_at` | **不一致** |
| UpdatedAt | `updatedAt` | `updatedAt` | `updated_at` | **不一致** |

### 4.3 Task

| 字段 | Backend Domain (json tag) | Backend DB (json tag) | Frontend ApiTask | Frontend domain Task | 一致? |
|------|--------------------------|---------------------|-----------------|-------------------|------|
| ID | `task_id` | `task_id` | `id` | `id` | **不一致** |
| Name | `name` | `name` | `name` | `name` | 一致 |
| Type | `type` | `type` | `handler` | `type` | **不一致** |
| Status | `status` | `status` | `status` | `status` | 一致 |
| Progress | `progress` | `progress` | `progress` | `progress` | 一致 |
| Priority | `priority` | `priority` | `priority` | - | 一致 |
| Handler | `handler` | `handler` | `handler` | - | 一致 |
| Result | `result` | `result` | `result` | - | 一致 |
| Error | `error` | `error` | `error` | - | 一致 |
| CreatedAt | `created_at` | `created_at` | `createdAt` | `created_at` | **不一致** |
| UpdatedAt | `updated_at` | `updated_at` | `updatedAt` | `updated_at` | **不一致** |
| StartTime | `start_time` | `start_time` | `startedAt` | `startedAt` | **不一致** |
| EndTime | `end_time` | `end_time` | `completedAt` | `completedAt` | **不一致** |
| ProjectID | `project_id` | `project_id` | - | `projectId` | **不一致** |
| WorkflowID | `workflow_id` | `workflow_id` | - | `workflowId` | **不一致** |

### 4.4 Plugin

| 字段 | Backend DB Model (json tag) | Frontend ApiPluginSummary | Frontend PluginStore Plugin | 一致? |
|------|---------------------------|-------------------------|--------------------------|------|
| ID | `id` (uint) / `plugin_id` | `id: string` | `id` | 部分 |
| Name | `name` | `name` | `name` | 一致 |
| Version | `version` | `version` | `version` | 一致 |
| Author | `author` | `author` | `author` | 一致 |
| Type | `type` | `type` | `category` | **不一致** |
| Status | `status` | `status` | `status` | 一致 |
| Enabled | `enabled` | `enabled` | - | 一致 |
| CreatedAt | `created_at` | `createdAt` | `installedAt` | **不一致** |
| UpdatedAt | `updated_at` | `updatedAt` | `updatedAt` | **不一致** |
| Source | `source` | - | `source` | 一致 |
| Nodes | - | - | `workflowNodes` | 仅 PluginStore |
| Dependencies | - | - | `dependencies` | 仅 PluginStore |

### 4.5 Task Status 值对照

| Backend 域值 | Backend DB 默认值 | Frontend mapTaskStatus() | Frontend 日志类型 | 一致? |
|-------------|-----------------|------------------------|----------------|------|
| `waiting` | `waiting` | **未处理 (返回 'running')** | - | **不一致** |
| `running` | - | `running` | `running` | 一致 |
| `success` | - | `completed` -> `success` | `success` | **不一致** |
| `failed` | - | `failed` | `failed` | 一致 |
| `cancelled` | - | **未处理 (返回 'running')** | - | **不一致** |
| - | - | `warning` | `warning` | Frontend 独有 |

**问题**: Backend 默认状态 `waiting` 和 `cancelled` 在 Frontend `mapTaskStatus()` 中未正确处理。Backend 使用 `success` 但 Frontend 映射 `completed` -> `success`，说明存在中间状态转换不匹配。

---

## 5. 断点汇总

### 5.1 前端调用但后端未实现的 API

| 序号 | Frontend 调用 | 严重程度 | 影响 |
|------|-------------|---------|------|
| 1 | `PUT /api/plugins/:name/update` | 高 | 插件更新功能不可用 |
| 2 | `GET /api/tasks/:taskId/logs` | 高 | 任务日志查看不可用 |
| 3 | `POST /api/error/analyze` | 中 | AI 错误分析不可用 |
| 4 | `POST /api/error/repair` | 中 | AI 自动修复不可用 |
| 5 | `GET /api/error/analysis/:taskId` | 中 | 获取分析结果不可用 |
| 6 | `GET /api/error/fix/:fixId/status` | 中 | 修复状态查询不可用 |
| 7 | `GET /api/settings` | 中 | 设置读取不可用 |
| 8 | `PUT /api/settings` | 中 | 设置保存不可用 |
| 9 | `GET /api/settings/engine` | 低 | 引擎配置读取不可用 |
| 10 | `PUT /api/settings/engine` | 低 | 引擎配置保存不可用 |
| 11 | `POST /api/settings/engine/test` | 低 | 引擎连接测试不可用 |

### 5.2 数据结构不匹配

| 序号 | 问题 | 严重程度 |
|------|------|---------|
| 1 | Task `id` vs `task_id` 字段名不匹配 | 高 |
| 2 | Task `createdAt` vs `created_at` 字段名不匹配 | 高 |
| 3 | Task `startedAt`/`completedAt` vs `start_time`/`end_time` 不匹配 | 高 |
| 4 | Task 默认状态 `waiting` 未被 Frontend 处理 | 中 |
| 5 | Task `cancelled` 状态未被 Frontend 处理 | 中 |
| 6 | Plugin `type` vs `category` 字段名不匹配 | 中 |
| 7 | Plugin `createdAt` vs `created_at` 字段名不匹配 | 中 |
| 8 | Frontend domain 类型使用 snake_case，API 层使用 camelCase | 低 |

### 5.3 WebSocket 事件不匹配

| 序号 | 问题 | 严重程度 |
|------|------|---------|
| 1 | Frontend 期望 `task_log` 事件，Backend 不发送 | 中 |
| 2 | Backend `task.cancelled` 映射为 `task_error` 类型，Frontend 处理为错误 | 低 |

### 5.4 配置不一致

| 序号 | 问题 | 严重程度 |
|------|------|---------|
| 1 | Docs 设计端口 8080，实际 8081 | 低 |
| 2 | Docs 设计 `/api/v1` 前缀，实际 `/api` | 低 |
| 3 | Frontend `request.ts` 硬编码地址，忽略 `.env` 配置 | 低 |
| 4 | Backend `websocket.port: 8082` 配置项无实际用途 (WS 走 HTTP 端口) | 低 |
| 5 | `Config/app.yaml` 与 `Backend/config/default.yaml` 配置重叠 | 低 |
