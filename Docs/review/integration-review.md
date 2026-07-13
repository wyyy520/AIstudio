# AIStudio 集成联调审查报告

> **审查日期**: 2026-07-10  
> **审查范围**: 完整调用链 - Frontend → API → Backend → Workflow → Task → Plugin → Engine → Result → Logs → UI  
> **审查目标**: 确认每个模块真正连接，无断点

---

## 一、调用链全景

```
Frontend (Vue 3 + Pinia)
    │
    ▼  HTTP POST /api/workflows/:id/run
API (Axios → Gin)
    │
    ▼  services.WorkflowService
Backend (Service Layer)
    │
    ▼  taskMgr.CreateTask → taskMgr.StartTask
Workflow Engine (task.TaskHandler)
    │
    ▼  engine.Run(ctx, workflowJSON)
Task Manager (Worker Pool)
    │
    ▼  handler.Execute(ctx, task)
Plugin Executor (SimpleExecutor)
    │
    ▼  engineRunner.RunPluginAction()
Python Engine (subprocess)
    │
    ▼  stdout JSON lines
Result Parsing
    │
    ▼  taskMgr.FinishTask / taskMgr.FailTask
Logs (LogService + TaskLogger)
    │
    ▼  WebSocket broadcast
UI Update (Pinia store → Vue component)
```

---

## 二、断点分析

### 2.1 ✅ 正常连接部分

| 链路 | 状态 | 验证文件 |
|------|------|---------|
| Frontend → API | ✅ | [router.ts](file:///d:/AIstudio-master/AIstudio-master/Frontend/src/router/index.ts) → [request.ts](file:///d:/AIstudio-master/AIstudio-master/Frontend/src/api/request.ts) |
| API → Backend Service | ✅ | [router.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/router.go) → [service.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/service/service.go) |
| Backend → Task Manager | ✅ | [workflow.go handler](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/handlers/workflow.go) → [task_service.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/service/task_service.go) |
| Task → Worker | ✅ | [manager.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/task/manager.go) → [worker.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/task/worker.go) |
| Worker → Engine Handler | ✅ | [worker.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/task/worker.go) → [runner.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/engine/runner.go) |
| Engine → Python subprocess | ✅ | [python.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/engine/python.go) → [runner.py](file:///d:/AIstudio-master/AIstudio-master/Engine/runner.py) |
| Result → Task Manager | ✅ | [python.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/engine/python.go) L156-L175 |
| Task → WebSocket | ✅ | [websocket.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/handlers/websocket.go) |
| WebSocket → Frontend | ✅ | [websocket.ts](file:///d:/AIstudio-master/AIstudio-master/Frontend/src/api/websocket.ts) |

### 2.2 ❌ 断点/问题

#### 断点1：Plugin Executor 不真正调用 Python Engine（CRITICAL）

**位置**: [plugin/executor.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/plugin/executor.go) L57-L69

```go
func (e *SimpleExecutor) Execute(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error) {
    // ...
    if runner != nil {
        // 这里使用了 ctx.Value("request_id") 但未设置
        taskID := fmt.Sprintf("plugin-%s-%d", name, ctx.Value("request_id"))
        // taskID 可能为空
        result, err := runner.RunPluginAction(ctx, taskID, name, action, input)
    }
    // Fallback: mock execution
}
```

**问题**:
1. `ctx.Value("request_id")` 从未在上下文中设置，导致 `taskID` 为 `plugin-<name>-%!d(<nil>)`
2. 如果 `engineRunner` 为 nil，直接退回 mock 模式，不报错也不提示
3. 没有超时控制

**影响**: 插件执行在不经意间返回假数据，用户无法察觉。

#### 断点2：Workflow 节点不执行真实操作（CRITICAL）

**位置**: [node.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/workflow/node.go)

**问题**: 所有节点（`YOLODetectorNode`, `PyTorchNode`, `TransformerNode` 等）返回硬编码模拟数据，不调用 `PluginExecutor` 或 `PythonEngine`。

**影响**: 工作流执行结果完全是假的，无法用于实际训练/推理。

#### 断点3：Python Engine 返回结果未被 LogService 记录（HIGH）

**位置**: [python.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/engine/python.go) L132-L148

**问题**: `handleLog` 将日志写入 `TaskLogger`（内存），但未同步写入 `LogService`（应用日志系统）。这意味着通过 `/api/logs` 查不到 Engine 执行日志。

**影响**: 用户无法通过日志中心查看 Engine 执行日志。

#### 断点4：Frontend 工作流页面未连接真实数据（HIGH）

**位置**: [Frontend/src/views/Workflow.vue](file:///d:/AIstudio-master/AIstudio-master/Frontend/src/views/Workflow.vue)

**问题**: 从 `Workflow.vue` 的组件结构和路由配置看，工作流页面依赖 `WorkflowStore` 但未从 API 获取真实数据。

**建议修复**: 连接 `WorkflowService` 的 API 调用。

#### 断点5：Environment 检查结果未传递到前端（MEDIUM）

**位置**: [environment.go handler](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/handlers/environment.go)

**问题**: `GET /api/environment/status` 存在但前端 Dashboard 的健康检查使用的是 `/api/health`，两者未整合。

**影响**: Dashboard 显示的环境状态可能不准确。

#### 断点6：Task 日志未关联到 Workflow 执行（MEDIUM）

**问题**: Workflow 执行时，`TaskHandler` 调用 `engine.Run()` 但不记录每个节点的执行日志到 `LogService`，导致日志中心无法按 Workflow 查询。

---

## 三、修复计划

### 3.1 立即修复（CRITICAL）

| 断点 | 修复方案 | 涉及文件 |
|------|---------|---------|
| 1 | 使用 UUID 生成 taskID，移除 ctx.Value 依赖 | [executor.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/plugin/executor.go) |
| 2 | 让 Workflow 节点通过 PluginExecutor 调用真实引擎 | [node.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/workflow/node.go) |
| 3 | Python Engine 日志同时写入 TaskLogger 和 LogService | [python.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/engine/python.go) |

### 3.2 短期修复（HIGH）

| 断点 | 修复方案 |
|------|---------|
| 4 | 连接 Workflow 页面到真实 API |
| 5 | 整合 `/api/environment/status` 和 `/api/health` |
| 6 | Workflow 执行时记录每个节点的日志到 LogService |

### 3.3 验证方法

修复后执行以下验证：

```
1. 创建 Workflow → 点击 Run
2. 确认 Task 状态从 waiting → running → success/failed
3. 确认 WebSocket 收到实时事件
4. 确认日志中心可按 taskId 查询到执行日志
5. 确认 Python subprocess 实际执行（非 mock）