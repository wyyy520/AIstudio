# 系统联动检查报告

> 检查时间: 2026-07-09
> 检查范围: Frontend / Backend / Engine / Plugins
> 目标: 发现模块间断点，提供修复方案

---

## 一、系统架构总览

```
┌──────────────┐     HTTP/REST      ┌──────────────┐    subprocess     ┌──────────────┐
│   Frontend   │ ◄────────────────► │   Backend    │ ◄──────────────► │   Engine     │
│  (Vue 3 +    │     WebSocket      │   (Go/Gin)   │    stdout JSON   │  (Python)    │
│   TypeScript)│                    │              │                  │              │
└──────────────┘                    └──────────────┘                  └──────────────┘
                                           │                               │
                                           │ GORM                          │ plugin import
                                           ▼                               ▼
                                    ┌──────────────┐              ┌──────────────┐
                                    │   Database   │              │   Plugins    │
                                    │  (SQLite)    │              │  (YOLO等)    │
                                    └──────────────┘              └──────────────┘
```

**通信方式**:
- Frontend ↔ Backend: HTTP REST + WebSocket (同端口 8081)
- Backend → Engine: 子进程调用 `python runner.py --task task.json`，Engine 通过 stdout JSON-lines 返回结果
- Engine → Plugins: Python 模块直接 import 调用

---

## 二、已连接模块

### 2.1 Frontend ↔ Backend (HTTP REST)

**已连接的 API 模块**:

| 模块 | 状态 | 说明 |
|------|------|------|
| Auth (登录/注册/登出) | ✅ 已连接 | 路径、参数、响应格式一致 |
| Project CRUD | ✅ 已连接 | 路径一致，字段名有差异但可工作 |
| Workflow CRUD + Run | ✅ 已连接 | 路径一致，Run 返回 task_id 可跟踪 |
| Task CRUD + Status | ✅ 已连接 | 路径基本一致 |
| Plugin List/Install/Execute | ✅ 已连接 | 主要路径一致 |
| Agent Chat | ✅ 已连接 | 路径一致 |
| Log Query | ✅ 已连接 | 全局查询可用 |
| Health Check | ✅ 已连接 | 路径一致 |

**连接但存在问题**:

| 问题 | 位置 | 影响 |
|------|------|------|
| Task ID 字段名不匹配 | Backend `task_id` vs Frontend `id` | Task 列表可能无法正确显示 ID |
| Task 时间字段名不匹配 | Backend `start_time/end_time` vs Frontend `startedAt/completedAt` | 任务时间信息丢失 |
| Task 默认状态 `waiting` 未处理 | Frontend `mapTaskStatus()` | 新建任务显示为"运行中" |
| Plugin `type` vs `category` | Backend `type` vs Frontend `category` | 插件分类显示可能异常 |
| Workflow Run 返回值嵌套 | Backend `{code, data: {task_id}}` vs Frontend 期望直接 `task_id` | **运行工作流后无法获取 task_id** |

### 2.2 Frontend ↔ Backend (WebSocket)

**已连接**: WebSocket 连接和基础事件推送正常工作。

| 事件 | Backend 发送 | Frontend 接收 | 状态 |
|------|-------------|-------------|------|
| `task_status` | ✅ | ✅ | 正常 |
| `task_progress` | ✅ | ✅ | 正常 |
| `task_complete` | ✅ | ✅ | 正常 |
| `task_error` | ✅ | ✅ | 正常 |
| `task_log` | ❌ 不发送 | ✅ 期望接收 | **断点** |

### 2.3 Backend → Engine

**已连接**: Backend 通过子进程调用 Python Engine，Engine 通过 stdout JSON-lines 返回结果。

| 组件 | 状态 | 说明 |
|------|------|------|
| PythonRunner 调用 | ✅ 已连接 | `engine/runner.go` 正确调用 `python runner.py --task task.json` |
| stdout JSON 解析 | ✅ 已连接 | `engine/python.go` 正确解析 JSON-lines |
| 事件转发到 Task 系统 | ✅ 已连接 | Engine 事件通过 EventBus 广播到 WebSocket |
| 插件注册 | ✅ 已连接 | `PLUGIN_REGISTRY` 映射 plugin name → Python module |

**存在的问题**:
- Engine 的 `PLUGIN_REGISTRY` 只注册了 `yolo`，其他插件（SAM、Transformer、LSTM）未注册
- Backend Plugin Executor 有别名映射 `yolo-detector` → `yolo`，但其他插件名未映射

### 2.4 Backend ↔ Database

**已连接**: GORM 正确操作 SQLite 数据库。

| 表 | 模型 | 状态 |
|----|------|------|
| users | `models.User` | ✅ |
| projects | `models.Project` | ✅ |
| tasks | `models.Task` | ✅ |
| workflows | `models.Workflow` | ✅ |
| plugins | `models.Plugin` | ✅ |
| sessions | `models.Session` | ✅ |
| api_keys | `models.APIKey` | ✅ |
| permissions | `models.Permission` | ✅ |
| quotas | `models.Quota` | ✅ |

---

## 三、断点列表

### 3.1 高严重度断点 (功能不可用)

#### 断点 H1: Workflow Run 返回值解析失败

**位置**:
- Frontend: `src/store/workflow.ts:186` → `lastRunTaskId.value = result.task_id`
- Backend: `internal/api/handlers/workflow.go:200-209` → 返回 `{code, message, data: {task_id, ...}}`

**问题**: Frontend `runWorkflow()` 通过 `request.ts` 发起 POST，响应拦截器返回 `response.data`（即 `{code, message, data}`）。Frontend store 直接访问 `result.task_id`，但实际结构是 `result.data.task_id`。

**影响**: 运行工作流后无法获取 task_id，无法跟踪任务进度。

**修复方案**: 在 `src/store/workflow.ts` 中修改为 `result.data?.task_id`，或在 `src/api/workflow.ts` 中对 `runWorkflow()` 的返回值做解包处理。

---

#### 断点 H2: Task ID 字段名不匹配

**位置**:
- Backend domain model: `internal/task/models.go:53` → `ID string \`json:"task_id"\``
- Backend DB model: `internal/database/models/task.go:9` → `TaskID string \`json:"task_id"\``
- Frontend: `src/api/task.ts:4` → `id: string`

**问题**: Backend JSON 序列化字段名为 `task_id`，Frontend 期望 `id`。

**影响**: Task 列表和详情中 ID 字段为空或 undefined。

**修复方案**: 统一为 `id`（snake_case: `id` 或 camelCase: `taskId`），在 Backend 添加 `json:"id"` tag 或在 Frontend 添加映射。

---

#### 断点 H3: Task 时间字段名不匹配

**位置**:
- Backend: `start_time`, `end_time` (snake_case)
- Frontend ApiTask: `startedAt`, `completedAt` (camelCase)

**问题**: 字段名完全不匹配。

**影响**: 任务开始/完成时间信息在 Frontend 中丢失。

**修复方案**: 统一命名。Backend 使用 `started_at`/`completed_at` 或 Frontend 映射 `start_time` → `startedAt`。

---

#### 断点 H4: Plugin Update 路由不存在

**位置**:
- Frontend: `src/api/plugin.ts:76` → `PUT /api/plugins/${name}/update`
- Backend: `internal/api/router.go:117-125` → 无此路由

**问题**: Frontend 调用的 API 端点在 Backend 中不存在。

**影响**: 插件更新功能不可用。

**修复方案**: Backend 添加 `PUT /api/plugins/:name/update` 路由，或 Frontend 改用 `PUT /api/plugins/:name/status` 实现更新。

---

#### 断点 H5: Task Logs 路由不存在

**位置**:
- Frontend: `src/api/log.ts` → `GET /api/tasks/:taskId/logs`
- Backend: `internal/api/router.go:154-157` → 只有 `GET /api/logs`

**问题**: Frontend 期望按 taskId 获取日志，Backend 只提供全局日志查询。

**影响**: 任务日志查看功能不可用。

**修复方案**: Backend 在 LogHandler 中添加按 taskId 过滤的支持，或添加 `GET /api/tasks/:taskId/logs` 路由。

---

#### 断点 H6: Error Analysis API 未实现

**位置**:
- Frontend: `src/api/error.ts` → 4 个 API 端点
- Backend: 无对应 handler

**问题**: 整个错误分析模块的 Backend API 未实现。

**影响**: AI 错误分析和自动修复功能不可用。

**修复方案**: Backend 实现 `error_handler.go`，提供 `/api/error/*` 路由。

---

#### 断点 H7: Settings API 未实现

**位置**:
- Frontend: `src/api/settings.ts` → 5 个 API 端点
- Backend: 无对应 handler

**问题**: 整个设置模块的 Backend API 未实现。

**影响**: 应用设置和引擎配置功能不可用。

**修复方案**: Backend 实现 `settings_handler.go`，提供 `/api/settings/*` 路由。

---

### 3.2 中严重度断点 (功能降级)

#### 断点 M1: Task 默认状态 `waiting` 未处理

**位置**: Frontend `src/store/log.ts:42-50`

**问题**: Backend Task 默认状态为 `waiting`，Frontend `mapTaskStatus()` 未处理此值，默认返回 `running`。

**影响**: 新建的任务在 Frontend 中显示为"运行中"而非"等待中"。

**修复方案**: 在 `mapTaskStatus()` 中添加 `case 'waiting': return 'pending'`。

---

#### 断点 M2: Task `cancelled` 状态未处理

**位置**: Frontend `src/store/log.ts:42-50`

**问题**: Backend 支持 `cancelled` 状态，Frontend `mapTaskStatus()` 未处理。

**影响**: 取消的任务显示为"运行中"。

**修复方案**: 在 `mapTaskStatus()` 中添加 `case 'cancelled': return 'cancelled'`。

---

#### 断点 M3: WebSocket `task_log` 事件缺失

**位置**:
- Frontend: `src/api/websocket.ts:1` → 期望 `task_log` 事件
- Backend: `internal/api/handlers/websocket.go:187-276` → 不发送此事件

**问题**: Frontend 的日志实时推送功能依赖 `task_log` 事件，但 Backend 不发送。

**影响**: 实时日志流功能不可用，只能通过轮询获取日志。

**修复方案**: Backend 在 Task EventBus 中添加日志事件，并在 WebSocket handler 中映射为 `task_log` 类型。

---

#### 断点 M4: Plugin `type` vs `category` 字段名不一致

**位置**:
- Backend Plugin JSON: `type` 字段
- Frontend PluginStore types: `category` 字段

**问题**: 同一概念使用不同字段名。

**影响**: 插件分类信息可能无法正确显示。

**修复方案**: 统一为 `category` 或 `type`。

---

#### 断点 M5: Dual Store 系统

**位置**:
- `src/store/` (7 个 store): workflow, project, task, plugin, log, settings, theme
- `src/stores/` (6 个 store): user, chat, project, workflow, settings, index

**问题**: 两套 store 系统并存，功能重叠。

**影响**: 维护成本高，可能导致状态不一致。

**修复方案**: 将 `src/stores/` 中的功能合并到 `src/store/`，删除冗余 store。

---

### 3.3 低严重度问题

#### 断点 L1: Frontend 硬编码 Base URL

**位置**: `src/api/request.ts:5` → `baseURL: 'http://127.0.0.1:8081'`

**问题**: 忽略了 `.env` 中的 `VITE_API_BASE_URL` 配置。

**修复方案**: 改为 `import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:8081'`。

---

#### 断点 L2: Docs 与实际端口/前缀不一致

**位置**: `Docs/protocol/api-standard.md` → 端口 8080，前缀 `/api/v1`

**问题**: 设计文档与实际实现不符。

**修复方案**: 更新文档以反映实际实现。

---

#### 断点 L3: Backend `websocket.port` 配置无意义

**位置**: `Backend/config/default.yaml:42` → `websocket.port: "8082"`

**问题**: WebSocket 实际运行在 HTTP 端口 8081 上，此配置项无实际作用。

**修复方案**: 移除此配置项或添加注释说明。

---

#### 断点 L4: Config 双源问题

**位置**:
- `Config/app.yaml` (Launcher 读取)
- `Backend/config/default.yaml` (Backend 读取)

**问题**: 两处定义了重叠的配置（端口、数据库等），可能不一致。

**修复方案**: 统一配置源，Backend 读取 `Config/app.yaml` 或在 `Config/app.yaml` 中引用 Backend 配置。

---

## 四、Workflow 完整链路验证

### 4.1 链路图

```
Frontend 创建 Workflow
    │  POST /api/workflows {projectId, name, definition}
    ▼
Backend WorkflowHandler.Create()
    │  调用 WorkflowService.Create(projectID, name, definition)
    │  存储到 workflows 表 (definition 为 JSON 字符串)
    ▼
Frontend 运行 Workflow
    │  POST /api/workflows/:id/run
    ▼
Backend WorkflowHandler.Run()
    │  1. 获取 workflow 定义
    │  2. 解析 definition JSON
    │  3. 创建 Task (via TaskService.Create)
    │  4. 返回 {task_id, status}
    ▼
Backend TaskService.Create()
    │  1. 创建 Task (status=waiting)
    │  2. 立即 StartTask (status=running)
    │  3. 分发到注册的 TaskHandler
    ▼
Backend workflow.TaskHandler.Execute()
    │  1. 解析 workflow JSON
    │  2. 环境检查 (可选)
    │  3. 调用 Engine.Run(workflowJSON)
    ▼
Backend workflow.Engine.Run()
    │  1. 拓扑排序 DAG
    │  2. 按顺序执行节点
    │  3. 每个节点调用 PluginManager
    ▼
Backend plugin.SimpleExecutor.ExecuteNode()
    │  调用 engine.TaskHandler.RunPluginAction()
    ▼
Backend engine.PythonRunner.Run()
    │  写入 task.json 文件
    │  启动子进程: python runner.py --task task.json
    ▼
Engine runner.py
    │  1. 读取 task.json
    │  2. 验证 required fields
    │  3. 查找 PLUGIN_REGISTRY
    │  4. 动态 import handler module
    │  5. 调用 run_train(config) / run_predict(config)
    ▼
Engine handler (vision/yolo/train.py)
    │  执行 YOLO 训练
    │  输出 JSON-lines 到 stdout:
    │    {"type": "progress", "data": {...}}
    │    {"type": "log", "data": {...}}
    │    {"type": "result", "data": {...}}
    ▼
Backend engine.PythonRunner
    │  解析 stdout JSON-lines
    │  通过 EventBus 广播事件
    ▼
Backend task.EventBus → WebSocketHandler
    │  映射为 WebSocketEvent
    │  广播到所有连接的客户端
    ▼
Frontend WebSocketClient
    │  接收事件
    │  更新 store 状态
    │  UI 实时刷新
```

### 4.2 链路断点

| 步骤 | 断点 | 严重程度 |
|------|------|---------|
| Frontend → Backend (Run) | H1: 返回值解析失败 | 高 |
| Backend → Task 创建 | M1: `waiting` 状态未处理 | 中 |
| Backend → Engine | ✅ 正常 | - |
| Engine → Plugin | ⚠️ 只注册了 yolo | 低 |
| Engine → Backend (stdout) | ✅ 正常 | - |
| Backend → WebSocket | M3: `task_log` 事件缺失 | 中 |
| WebSocket → Frontend | M1/M2: 状态值未处理 | 中 |

---

## 五、服务通信统一性检查

### 5.1 通信方式统计

| 通信路径 | 方式 | 协议 | 状态 |
|---------|------|------|------|
| Frontend → Backend | HTTP REST | JSON over HTTP | ✅ 唯一 |
| Frontend ← Backend | WebSocket | JSON over WS | ✅ 唯一 |
| Backend → Engine | 子进程 | stdin/stdout JSON-lines | ✅ 唯一 |
| Backend → Database | GORM | SQL | ✅ 唯一 |

**结论**: 通信方式已统一，无重复。每对模块之间只有一种通信方式。

### 5.2 通信方式问题

| 问题 | 说明 |
|------|------|
| Backend 有 HTTP Server 模式 (server.py) 但未使用 | Engine 同时支持子进程和 HTTP 两种模式，Backend 只使用子进程模式 |
| Engine README 描述了 gRPC 但未实现 | gRPC 目录为空，实际使用 subprocess |

**建议**: 保持子进程模式，移除未实现的 gRPC 和 HTTP Server 相关代码，避免混淆。

---

## 六、错误处理统一性检查

### 6.1 各模块错误处理方式

| 模块 | 错误格式 | 状态码 | 错误码 |
|------|---------|--------|--------|
| Backend common | `{code: -1, message, data: APIError}` | HTTP status | ErrorCode (string) |
| Backend inline | `{code: -1, message}` | HTTP status | 无 |
| Frontend client.ts | `throw new Error(message)` | - | - |
| Engine stdout | `{"type": "error", "data": {message}}` | - | - |
| Engine exit code | `sys.exit(1)` | - | - |

### 6.2 问题

| 问题 | 说明 |
|------|------|
| Backend 错误码不统一 | `common.RespondError` 使用 `code: -1`，Docs 定义了数字错误码 (1001-7003) 但未实现 |
| Backend 两种错误返回方式 | `common.RespondError` vs 内联 `gin.H{"code": -1, ...}` |
| Frontend 不使用错误详情 | `client.ts` 只读 `message`，忽略 `data` 中的 `APIError` 结构 |
| Engine 错误码缺失 | 只有 `error` 事件，无结构化错误码 |

**建议**:
1. 统一 Backend 错误返回为 `common.RespondError()`
2. 实现 Docs 定义的数字错误码
3. Frontend `client.ts` 解析 `data` 中的错误详情

---

## 七、配置统一性检查

### 7.1 配置文件清单

| 文件 | 读取者 | 内容 |
|------|--------|------|
| `Config/app.yaml` | Launcher | 全局配置：端口、路径、启动参数 |
| `Backend/config/default.yaml` | Backend | Backend 默认配置 |
| `Backend/config/development.yaml` | Backend | 开发环境覆盖 |
| `Backend/config/production.yaml` | Backend | 生产环境覆盖 |
| `Frontend/.env` | Frontend (Vite) | `VITE_API_BASE_URL` |
| `Engine/requirements.txt` | Python | 依赖版本 |

### 7.2 配置冲突

| 配置项 | Config/app.yaml | Backend/config | Frontend/.env | 一致? |
|--------|----------------|---------------|--------------|------|
| Backend 端口 | 8081 | 8081 | localhost:8081 | ✅ |
| Engine 端口 | 8082 | - | - | Backend 未配置 |
| 数据库 | - | sqlite: aistudio.db | - | Backend 独立 |
| WebSocket 端口 | - | 8082 (无用) | - | 配置多余 |

### 7.3 问题

| 问题 | 说明 |
|------|------|
| Config/app.yaml 与 Backend/config 重叠 | 两处定义了 Backend 端口、数据库等 |
| Launcher 不传递配置给 Backend | 设计中 Launcher 应传 `AISTUDIO_CONFIG` 环境变量，但实际未使用 |
| Backend 不读取 Config/app.yaml | Backend 只读自己的 config 目录 |

**建议**: 以 `Config/app.yaml` 为单一配置源，Backend 读取此文件（或通过环境变量接收路径）。

---

## 八、修改方案优先级

### 第一优先级 (必须修复 - 功能断点)

| 编号 | 问题 | 修改位置 | 工作量 |
|------|------|---------|--------|
| H1 | Workflow Run 返回值解析 | `src/store/workflow.ts` 或 `src/api/workflow.ts` | 小 |
| H2 | Task ID 字段名不匹配 | Backend `task/models.go` + `database/models/task.go` | 中 |
| H3 | Task 时间字段名不匹配 | Backend `task/models.go` + `database/models/task.go` | 中 |
| H4 | Plugin Update 路由缺失 | Backend `api/router.go` + `handlers/plugin.go` | 小 |
| H5 | Task Logs 路由缺失 | Backend `api/router.go` + `handlers/log.go` | 中 |
| H6 | Error Analysis API 缺失 | Backend 新建 `handlers/error.go` | 大 |
| H7 | Settings API 缺失 | Backend 新建 `handlers/settings.go` | 大 |

### 第二优先级 (应该修复 - 功能降级)

| 编号 | 问题 | 修改位置 | 工作量 |
|------|------|---------|--------|
| M1 | Task `waiting` 状态未处理 | `src/store/log.ts` | 小 |
| M2 | Task `cancelled` 状态未处理 | `src/store/log.ts` | 小 |
| M3 | WebSocket `task_log` 事件缺失 | Backend `handlers/websocket.go` | 中 |
| M4 | Plugin `type` vs `category` | Backend 或 Frontend 统一 | 小 |
| M5 | Dual Store 合并 | `src/stores/` → `src/store/` | 大 |

### 第三优先级 (建议修复 - 代码质量)

| 编号 | 问题 | 修改位置 | 工作量 |
|------|------|---------|--------|
| L1 | Frontend 硬编码 Base URL | `src/api/request.ts` | 小 |
| L2 | Docs 端口/前缀不一致 | `Docs/protocol/api-standard.md` | 小 |
| L3 | WebSocket 端口配置无用 | `Backend/config/default.yaml` | 小 |
| L4 | Config 双源 | 统一配置读取 | 中 |

---

## 九、总结

### 已连接的模块

- ✅ Frontend ↔ Backend: Auth, Project, Workflow, Task, Plugin, Agent, Log, Health
- ✅ Backend ↔ Database: 全部 9 张表 CRUD
- ✅ Backend → Engine: 子进程调用 + stdout JSON-lines
- ✅ Engine → Plugins: YOLO 训练/推理
- ✅ WebSocket: 任务状态实时推送

### 存在的问题

- **7 个高严重度断点**: Workflow Run 返回值、Task 字段名、Plugin Update 路由、Task Logs 路由、Error Analysis API、Settings API
- **5 个中严重度断点**: Task 状态值、WebSocket 事件、Plugin 字段名、Dual Store
- **4 个低严重度问题**: 硬编码 URL、文档不一致、配置冗余

### 核心结论

系统的主要模块已经建立连接，但 **数据结构不统一** 是最大的系统性问题。Backend 使用 snake_case (`task_id`, `created_at`, `start_time`)，Frontend Api 层使用 camelCase (`id`, `createdAt`, `startedAt`)，导致字段映射断裂。建议统一命名规范后批量修复。
