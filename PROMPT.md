# AIStudio 项目实施 — Prompt 工程文档

> **基于**: [ROADMAP.md](ROADMAP.md) v1.0
> **日期**: 2026-07-13
> **用途**: 将每个实施步骤转化为可交给 Claude Code 或其他 AI 编程助手的独立 Prompt
> **结构**: 每个 Prompt 均可直接复制使用，包含上下文、目标、文件清单、验收标准

---

## 使用说明

### 如何使用此文档

1. 从下方复制对应 **Phase N — Step X.Y.Z** 的完整 Prompt 块
2. 在 Claude Code 中粘贴执行，或作为 `/` 指令输入
3. 完成后对照该 Prompt 末尾的验收标准逐项确认
4. 全部通过后进入下一步

### 提示词结构说明

每个 Prompt 统一包含以下模块：

| 模块 | 说明 |
|------|------|
| **任务标题** | 简短描述该步骤做什么 |
| **背景** | 为什么需要这一步，当前状态是什么 |
| **目标** | 这一步完成后应该达到的状态 |
| **文件清单** | 新建文件和修改文件的详细说明，含代码片段 |
| **参考代码** | 项目中可参考的现有实现 |
| **验收标准** | 可逐项勾选的 Checklist |

### 实施顺序建议

```
第 1 周              第 2 周              第 3-4 周             第 5-6 周
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ Prompt 1.1   │    │ Prompt 1.2   │    │ Prompt 2.1   │    │ Prompt 3.1   │
│ Engine Bridge│───→│ 节点执行器    │    │ 安装流程完善  │    │ LLM 流式     │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘

┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ Prompt 1.3   │    │ Prompt 1.4   │    │ Prompt 2.2   │    │ Prompt 3.2   │
│ WebSocket    │    │ 插件执行     │    │ 多语言支持   │    │ Agent 流式   │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
                    (与 1.3 并行)                              │
                                                              └──────────────┐
                                                    ┌──────────────┐    ┌──────────────┐
                                                    │ Prompt 5.1   │    │ Prompt 3.3   │
                                                    │ Makefile     │    │ 上下文感知   │
                    ┌──────────────┐                └──────────────┘    └──────────────┘
                    │ Prompt 5.2   │                                     (或 5.x 并行)
                    │ Docker       │
                    └──────────────┘
```

---

## Phase 1 — 引擎打通：让工作流真正运行 (P0)

> **目标**: 用户创建项目 → 设计工作流 → 节点能调用真实的 Python 引擎 → 看到真实结果
> **预估**: 2-3 周
> **并行度**: Step 1.1 ~ 1.4 可以并行

---

### Prompt 1.1 — 创建 Go→Python HTTP Bridge

```
## 任务：创建 Go→Python HTTP Bridge

### 背景
AIStudio 的 Python AI Engine 已有 HTTP Server (`Engine/server.py`)，提供 `/health` (GET) 和 `/task` (POST) 端点。Go 后端目前未调用它，所有 Workflow 节点返回占位结果。需要创建 Go 端的 Engine Bridge 客户端。

### 目标
在 `apps/backend/internal/engine/` 下创建 Engine Bridge 包，实现 Go ↔ Python 的 HTTP 通信。

### 文件清单

#### 新建文件

1. **`apps/backend/internal/engine/client.go`** — EngineClient 接口

   定义接口 `EngineClient`，包含方法：
   - `Infer(ctx, req InferRequest) (*InferResponse, error)`
   - `Train(ctx, req TrainRequest) (*TrainResponse, error)`
   - `Health(ctx) (*HealthResponse, error)`
   - `LoadModel(ctx, req LoadModelRequest) (*LoadModelResponse, error)`

   定义请求/响应结构体：

   | 结构体 | 字段 |
   |--------|------|
   | `InferRequest` | `ModelName string` + `Input map[string]interface{}` + `Config map[string]interface{}` |
   | `InferResponse` | `Output map[string]interface{}` + `DurationMs int64` + `Error string` |
   | `TrainRequest` | `Dataset string` + `Config map[string]interface{}` + `ModelName string` |
   | `TrainResponse` | `ModelPath string` + `Metrics map[string]float64` + `DurationMs int64` + `Error string` |
   | `HealthResponse` | `Status string` + `Version string` + `Uptime int64` |
   | `LoadModelRequest` | `ModelName string` + `ModelPath string` |
   | `LoadModelResponse` | `Success bool` + `Error string` |

2. **`apps/backend/internal/engine/http_client.go`** — HTTP 客户端实现

   - 实现 `EngineClient` 接口，使用 `net/http` 标准库
   - `Health()` — GET 请求 `{baseURL}/health`，解析 JSON 返回
   - `Infer()` — POST 请求 `{baseURL}/task`，设置 `Content-Type: application/json`
   - `Train()` — POST 请求 `{baseURL}/task`
   - `LoadModel()` — POST 请求 `{baseURL}/task`
   - 所有方法需处理：连接超时、非 200 状态码、JSON 解析错误
   - 超时时间从 config 读取

3. **`apps/backend/internal/engine/config.go`** — Engine 配置

   ```go
   type Config struct {
       BaseURL    string        // 默认 http://localhost:8082
       Timeout    time.Duration // 默认 30s
       RetryCount int           // 默认 3
       RetryDelay time.Duration // 默认 1s
   }
   ```

   方法：`DefaultConfig() *Config` 返回默认值

4. **`apps/backend/internal/engine/engine.go`** — 包入口

   - 提供 `NewClient(config *Config) EngineClient` 工厂函数
   - 可选的日志包装器

#### 修改文件

5. **`apps/backend/internal/service/service.go`** — 注册到 Service Container

   - 添加 `EngineService() EngineClient` 方法到 Container
   - 参考 `RuntimeService()` 的实现模式

6. **`apps/backend/cmd/main.go`** — 初始化 Engine Bridge

   - 在 `initServices()` 中初始化 Engine Client
   - 配置从全局 Config 读取

### 参考代码

Python Engine 端点 (`Engine/server.py` line 43-56):
```python
@app.get("/health")
async def health():
    return {"status": "ok", "version": "1.0.0", "uptime": time.time() - start_time}

@app.post("/task")
async def execute_task(request: TaskRequest):
    # Route to appropriate handler based on request.action
    ...
```

现有 Service 模式 (`apps/backend/internal/service/runtime_service.go`):
```go
type ServiceContainer struct {
    runtime *RuntimeService
    // ...
}

func (c *Container) RuntimeService() *RuntimeService {
    return c.runtime
}
```

### 验收标准
- [ ] `apps/backend/internal/engine/` 包编译通过
- [ ] `NewClient()` 返回非 nil 的 EngineClient
- [ ] `client.Health()` 能成功调用 Python Engine 的 `/health` 端点并返回解析结果
- [ ] HTTP 超时、重试逻辑正确工作
- [ ] `ServiceContainer.EngineService()` 返回有效的 EngineClient 实例
- [ ] `go build ./apps/backend/...` 通过
```

---

### Prompt 1.2 — 实现 Workflow 节点真实执行器

```
## 任务：替换 noOpExecutor 为真实节点执行器

### 背景
当前 `apps/backend/internal/workflow/builtin_nodes.go` 中 40+ 个节点的 Factory 全部指向 `noOpExecutor`，该函数返回占位 mock 结果。需要将关键节点替换为真实执行器，调用 Engine Bridge 或内部逻辑。

### 目标
创建 `apps/backend/internal/workflow/executors/` 目录，实现 10 个关键节点的真实执行器，并更新 `builtin_nodes.go` 中的工厂函数。

### 文件清单

#### 新建文件

1. **`apps/backend/internal/workflow/executors/yolo_train.go`**
   - 实现 `func YOLOTrainExecutor(engine engine.EngineClient) ExecutableNode`
   - 从 inputs 提取 `dataset`，从 config 提取训练参数
   - 调用 `engine.Train(ctx, ...)`
   - 将 TrainResponse 中的 Metrics 和 ModelPath 写入 outputs

2. **`apps/backend/internal/workflow/executors/yolo_predict.go`**
   - 实现 `func YOLOPredictExecutor(engine engine.EngineClient) ExecutableNode`
   - 从 inputs 提取 `image` 或 `images`，从 config 提取推理参数
   - 调用 `engine.Infer(ctx, ...)`
   - 将 InferResponse 中的 Output 写入 outputs

3. **`apps/backend/internal/workflow/executors/condition.go`**
   - 实现 `func ConditionExecutor() ExecutableNode`
   - 从 config 读取 `expression` (如 `input > 0.5`)
   - 从 inputs 读取变量值
   - 使用 `goja` (Go JS 引擎) 或简单表达式引擎计算
   - 返回 `{"result": true/false, "branch": "true"/"false"}`

4. **`apps/backend/internal/workflow/executors/loop.go`**
   - 实现 `func LoopExecutor() ExecutableNode`
   - 从 config 读取 `maxIterations` (默认 10)
   - 从上下文获取当前迭代计数（通过 ctx 携带的迭代状态）
   - 每次执行递增计数器
   - 未达上限时输出 `{"continue": true, "currentIteration": n}`，否则 `{"continue": false}`

5. **`apps/backend/internal/workflow/executors/switch.go`**
   - 实现 `func SwitchExecutor() ExecutableNode`
   - 从 config 读取 `cases` 数组（每个 case 有 `value` 和 `outputPort`）
   - 从 inputs 读取 `switchValue`
   - 匹配 case 值，输出对应的 port 标识
   - 无匹配时走 `default` 分支
   - 返回 `{"selectedCase": "...", "selectedPort": "..."}`

6. **`apps/backend/internal/workflow/executors/retry.go`**
   - 实现 `func RetryExecutor(inner ExecutableNode) ExecutableNode`
   - 装饰器模式：包裹内部节点
   - 从 config 读取 `maxRetries` (默认 3) 和 `backoffMs` (默认 1000)
   - 执行内部节点，失败时按指数退避重试
   - 返回 `{"success": true/false, "attempts": n, "lastError": "..."}`

7. **`apps/backend/internal/workflow/executors/data_loader.go`**
   - 实现 `func DataLoaderExecutor() ExecutableNode`
   - 从 config 读取 `source` (文件路径) 和 `format` (csv/json/image)
   - 读取本地文件
   - CSV 用 `encoding/csv` 解析，JSON 用 `encoding/json` 解析
   - 返回 `{"data": [...], "rowCount": n, "columns": [...]}`

8. **`apps/backend/internal/workflow/executors/nlp.go`**
   - 实现 `func NLPExecutor(engine engine.EngineClient) ExecutableNode`
   - 从 inputs 提取 `text`，从 config 提取 `task` (分类/实体抽取/摘要)
   - 调用 `engine.Infer()`（未来可以通过特定 NLP 端点）
   - 返回 `{"result": "...", "confidence": 0.95}`

9. **`apps/backend/internal/workflow/executors/http_request.go`**
   - 实现 `func HTTPRequestExecutor() ExecutableNode`
   - 从 config 读取 `method`, `url`, `headers`, `body`
   - 使用 `net/http` 发送 HTTP 请求
   - 返回 `{"statusCode": 200, "body": "...", "headers": {...}}`

#### ExecutableNode 接口参考

```go
// 参考现有 types.go 中的 ExecutableNode 定义
type ExecutableNode interface {
    Execute(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error)
}
```

#### 修改文件

10. **`apps/backend/internal/workflow/builtin_nodes.go`** — 更新节点 Factory

    修改以下节点的 Factory 函数指向真实执行器（不再指向 `noOpExecutor`）：

    | 节点类型 | 新执行器 |
    |---------|---------|
    | `YOLOTrainNode` | `YOLOTrainExecutor(engine)` |
    | `YOLOPredictNode` | `YOLOPredictExecutor(engine)` |
    | `ConditionNode` | `ConditionExecutor()` |
    | `LoopNode` | `LoopExecutor()` |
    | `SwitchNode` | `SwitchExecutor()` |
    | `RetryNode` | `RetryExecutor(inner)` |
    | `DataLoaderNode` | `DataLoaderExecutor()` |
    | `NLPAnalyzerNode` | `NLPExecutor(engine)` |
    | `HTTPRequestNode` | `HTTPRequestExecutor()` |
    | `StartNode` / `EndNode` | 保留 noOp（标记节点） |

### 关键代码变更

当前（builtin_nodes.go）:
```go
func noOpExecutor(nodeID string) ExecutableNode {
    return executableFunc(func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
        return map[string]interface{}{
            "result":  "mock_result",
            "node_id": nodeID,
        }, nil
    })
}
```

改为:
```go
func yoloTrainExecutor(engineClient engine.EngineClient) ExecutableNode {
    return executableFunc(func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
        resp, err := engineClient.Train(ctx, engine.TrainRequest{
            Dataset:   toString(inputs["dataset"]),
            Config:    config,
        })
        if err != nil {
            return nil, fmt.Errorf("yolo train failed: %w", err)
        }
        if resp.Error != "" {
            return nil, fmt.Errorf("engine error: %s", resp.Error)
        }
        return map[string]interface{}{
            "model_path":  resp.ModelPath,
            "metrics":     resp.Metrics,
            "duration_ms": resp.DurationMs,
        }, nil
    })
}
```

### 验收标准
- [ ] `apps/backend/internal/workflow/executors/` 编译通过
- [ ] YOLO 训练执行器调用 `engine.Train()` 并正确传递参数
- [ ] IF 条件节点能正确评估 `>`, `<`, `==`, `>=`, `<=`, `&&`, `||` 表达式
- [ ] Loop 节点能正确维护迭代计数并在达到上限后停止
- [ ] Switch 节点能匹配多路分支
- [ ] Retry 节点在子节点失败时重试并遵循退避策略
- [ ] 所有修改后的节点在 Workflow 运行中返回真实数据（非占位符）
- [ ] `go test ./apps/backend/internal/workflow/...` 通过
```

---

### Prompt 1.3 — 实现 WebSocket 实时状态推送

```
## 任务：实现 WebSocket 实时状态推送系统

### 背景
Task EventBus 已有事件系统（`internal/eventbus/eventbus.go`），但事件仅在 Go 内部流转，未推送到前端。需要实现 WebSocket Hub 将 Task/Workflow 状态变化实时推送到前端。

### 目标
创建 `apps/backend/internal/api/ws/` 包，实现 WebSocket 连接管理、房间订阅、事件转发，并更新路由和前端对接。

### 文件清单

#### 新建文件

1. **`apps/backend/internal/api/ws/hub.go`** — WebSocket Hub

   ```go
   type Hub struct {
       rooms      map[string]map[*Client]bool // roomID → clients
       register   chan *Client
       unregister chan *Client
   }
   ```

   方法：
   - `NewHub() *Hub`
   - `Run()` — 主循环，处理 register/unregister/broadcast
   - `BroadcastToRoom(roomID string, msg []byte)` — 向房间内所有客户端发送
   - `BroadcastToUser(userID string, msg []byte)` — 向用户所有会话发送
   - `CreateRoom(roomID string)` / `DeleteRoom(roomID string)`

   房间命名规范：`task:{taskID}`、`workflow:{workflowID}`、`user:{userID}`

2. **`apps/backend/internal/api/ws/client.go`** — WebSocket 客户端

   ```go
   type Client struct {
       hub    *Hub
       conn   *websocket.Conn
       send   chan []byte
       userID string
       rooms  []string
   }
   ```

   方法：
   - `ReadPump()` — 读取消息循环（处理客户端发来的 subscribe/unsubscribe 消息）
   - `WritePump()` — 写入消息循环（从 send chan 读取并写入 WebSocket）

   客户端消息格式:
   ```json
   {"type": "subscribe", "room": "task:xxx"}
   {"type": "unsubscribe", "room": "task:xxx"}
   ```

3. **`apps/backend/internal/api/ws/messages.go`** — 消息类型定义

   消息类型常量：
   - `MsgTypeTaskStatus` = `"task_status"`
   - `MsgTypeNodeStatus` = `"node_status"`
   - `MsgTypeNodeLog` = `"node_log"`
   - `MsgTypeTaskDone` = `"task_done"`
   - `MsgTypeWorkflowProgress` = `"workflow_progress"`

   消息结构体:
   ```go
   type WsMessage struct {
       Type      string      `json:"type"`
       TaskID    string      `json:"taskId,omitempty"`
       NodeID    string      `json:"nodeId,omitempty"`
       Status    string      `json:"status,omitempty"`
       Progress  float64     `json:"progress,omitempty"`
       Payload   interface{} `json:"payload,omitempty"`
       Timestamp string      `json:"timestamp"`
   }
   ```

#### 修改文件

4. **`apps/backend/cmd/main.go`** — 订阅 Task EventBus

   在 `setupEventSubscriptions()` 函数中：
   ```go
   func setupEventSubscriptions(hub *ws.Hub, eventBus *eventbus.EventBus) {
       eventBus.Subscribe("task:status", func(event eventbus.Event) {
           msg := ws.NewMessage("task_status", event.Data)
           data, _ := msg.ToJSON()
           hub.BroadcastToRoom("task:"+event.Data["taskID"], data)
       })
   }
   ```

5. **`apps/backend/internal/api/router.go`** — 更新 WebSocket 路由

   添加路由：`ws.GET("/ws", wsHandler.HandleWebSocket)`

6. **`apps/desktop/src/api/websocket.ts`** — 前端 WebSocket 客户端对接

   实现 `WebSocketClient` 类：
   - `connect()` — 建立连接，自动重连（指数退避）
   - `subscribe(room)` / `unsubscribe(room)`
   - `onMessage(type, callback)` — 按消息类型注册回调
   - `disconnect()`

### 消息格式参考

```json
{
  "type": "node_status",
  "taskId": "task_001",
  "nodeId": "n2",
  "status": "running",
  "progress": 0.45,
  "timestamp": "2026-07-13T10:30:00Z"
}
```

### 验收标准
- [ ] WebSocket Hub 能管理多房间多客户端
- [ ] 客户端可通过消息订阅/取消订阅房间
- [ ] Task 状态变化通过 EventBus → WebSocket 推送到前端
- [ ] 前端 `websocket.ts` 能正确连接、重连和处理消息
- [ ] 前端 LogViewer/WorkflowConsole 显示实时状态
- [ ] 服务端 `go build` 通过
- [ ] 前端 `npm run build` 通过
```

---

### Prompt 1.4 — 实现插件执行系统

```
## 任务：为 Plugin V2 Manager 添加 Execute 方法和安装管道

### 背景
Plugin V2 Manager 能发现、注册、启用/禁用插件，但 `internal/plugin/manager.go` 没有 `Execute()` 方法，插件无法在工作流中执行。需要添加执行接口、执行器实现和安装管道。

### 目标
为插件系统添加执行能力，支持 Python 脚本执行和进程隔离模式，并实现可用的安装/卸载管道。

### 文件清单

#### 新建文件

1. **`apps/backend/internal/plugin/executors/interfaces.go`** — 执行器接口

   ```go
   type PluginExecutor interface {
       Execute(ctx context.Context, plugin *Plugin, input map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error)
       Language() string
   }
   ```

2. **`apps/backend/internal/plugin/executors/python_executor.go`** — Python 插件执行器

   - 从 Plugin Manifest 获取 entry point（如 `main.py`）
   - 使用 `os/exec` 调用 `python {plugin_path}/main.py`
   - 通过 stdin 传入 JSON 格式的 input
   - 从 stdout 读取 JSON 格式的 output
   - 设置超时（从 config 读取，默认 60s）
   - 错误时合并 stderr 到 error message

3. **`apps/backend/internal/plugin/executors/process_executor.go`** — 进程隔离执行器

   - 支持任意语言的可执行文件
   - stdin/stdout JSON 通信协议
   - 功能：进程启动/停止/超时控制、资源限制、日志收集、退出码检查
   - 方法：`NewProcessExecutor(path string, opts ProcessOptions)`、`Execute(ctx, input, config)`

4. **`apps/backend/internal/plugin/installer.go`** — 插件安装管道

   方法：
   - `Install(ctx, manifestURL string) (*InstallTask, error)` — 异步安装
   - `InstallSync(ctx, manifestURL string) error` — 同步安装
   - `Uninstall(ctx, pluginName string) error`
   - `GetInstallStatus(taskID string) *InstallStatus`

   安装流程：
   1. 解析 Manifest（从 URL 或本地文件）
   2. 检查依赖（Python 版本、CUDA、系统包）
   3. 创建插件目录 `plugins/{name}/`
   4. 下载代码（git clone 或 zip 解压）
   5. pip install requirements.txt（如果有）
   6. 下载模型权重（如果 manifest 指定）
   7. 执行 install.py（如果有）
   8. 在 Plugin Registry 注册

#### 修改文件

5. **`apps/backend/internal/plugin/manager.go`** — 添加 Execute 和注册方法

   - 添加 `executors map[string]PluginExecutor` 字段
   - 方法：
     - `RegisterExecutor(language string, executor PluginExecutor)`
     - `Execute(ctx, pluginName string, input, config) (map[string]interface{}, error)`
       - 查找 Plugin → 检查是否 Enabled → 获取执行器 → 执行并返回

6. **`apps/backend/internal/api/router.go`** — 添加插件安装 API

   | 方法 | 路由 | 说明 |
   |------|------|------|
   | POST | `/api/plugins/install` | 安装插件 |
   | DELETE | `/api/plugins/:name` | 卸载插件 |
   | GET | `/api/plugins/:name/status` | 获取安装状态 |
   | POST | `/api/plugins/:name/execute` | 执行插件 |

### 验收标准
- [ ] `Manager.RegisterExecutor()` 能注册新的语言执行器
- [ ] `Manager.Execute()` 能调用已注册的 Python 插件并返回结果
- [ ] 插件安装管道能解析 Manifest、安装依赖、下载模型
- [ ] 安装失败时返回详细错误信息
- [ ] `POST /api/plugins/install` 端点返回 `InstallTask` ID
- [ ] 前端可查询安装进度
- [ ] `go test ./apps/backend/internal/plugin/...` 通过
```

---

### Prompt 1.5 — Phase 1 集成验收

```
## 任务：Phase 1 集成验收测试

### 背景
Phase 1 的 4 个 Step 应已全部完成。需要编写集成测试验证端到端链路。

### 目标
编写集成测试脚本，验证整个 Phase 1 的端到端可用性。

### 测试场景

1. **Engine Bridge 连通性测试**
   - 启动 Python Engine (`cd Engine && python server.py --port 8082`)
   - 启动 Go 后端
   - 验证后端可调用 `GET /health` 返回正常
   - 验证 `/api/v1/engine/health` 返回 Python Engine 状态

2. **Workflow 执行测试**
   - 创建工作流 JSON（YOLO 训练 → YOLO 推理）
   - 通过 API `POST /api/workflows` 创建
   - 通过 `POST /api/workflows/:id/run` 执行
   - 轮询 `GET /api/tasks/:id` 直到完成
   - 验证节点状态依次为 running → success
   - 验证输出包含真实结果

3. **WebSocket 推送测试**
   - 建立 WebSocket 连接
   - 订阅 room `task:{taskID}`
   - 触发工作流执行
   - 验证 WebSocket 能收到 `node_status` 事件
   - 验证收到 `task_done` 事件

4. **插件执行测试**
   - 创建测试 Python 插件（含 manifest.json）
   - 通过 `POST /api/plugins/install` 安装
   - 执行插件 `POST /api/plugins/:name/execute`
   - 验证返回正确结果

### 验收标准
- [ ] `make test-phase-1` 脚本可运行全部 4 个测试场景
- [ ] Engine Bridge 连通性测试通过
- [ ] Workflow 执行测试通过（节点返回真实数据）
- [ ] WebSocket 推送测试通过
- [ ] 插件执行测试通过
- [ ] 测试失败时有清晰的错误信息
```

---

## Phase 2 — 插件系统：让节点可扩展 (P1)

> **目标**: 第三方开发者能编写、安装、发布插件
> **预估**: 2 周
> **前置**: Phase 1 Step 1.4

---

### Prompt 2.1 — 完善插件安装流程

```
## 任务：完善插件安装流程 — 依赖检查、模型下载、安装 UI

### 背景
Phase 1 已实现基本的插件安装管道。Phase 2 需要完善依赖检查、模型权重下载、安装脚本执行和前端安装 UI。

### 目标
增强安装器功能，确保安装过程可靠、可追踪，前端可显示安装进度。

### 文件清单

#### 修改文件

1. **`apps/backend/internal/plugin/installer.go`** — 完善安装流程

   实现依赖检查函数：
   ```go
   func checkDependencies(ctx, manifest) ([]DependencyCheck, error)
   ```
   - Python 版本检查：`python --version`
   - CUDA 可用性：`python -c "import torch; print(torch.cuda.is_available())"`
   - 系统包检查（Linux: `dpkg -l`, Windows: `where`）

   实现模型下载：
   ```go
   func downloadModel(ctx, url, destPath string, progressCh chan<- DownloadProgress) error
   ```
   - HTTP Range 请求支持断点续传
   - 下载进度推送到 progressCh
   - 文件校验（SHA256 checksum）

   安装状态追踪：
   ```go
   type InstallStatus struct {
       TaskID      string
       PluginName  string
       Phase       InstallPhase  // check_deps / download / install / register
       Progress    float64
       Log         []string
       Error       string
       StartedAt   time.Time
       CompletedAt *time.Time
   }
   ```

2. **`apps/desktop/src/components/plugin/`** — 前端安装 UI 对接

   确保 `PluginInstallTask.vue` 组件能展示：
   - 安装阶段标签（检查依赖 → 下载 → 安装 → 注册）
   - 进度条
   - 实时日志输出
   - 错误状态和重试按钮

### 验收标准
- [ ] 安装前自动检查 Python 版本和 CUDA 可用性
- [ ] 模型下载有进度报告和断点续传支持
- [ ] install.py 脚本执行并捕获日志
- [ ] 安装过程中可通过 API 查询详细状态
- [ ] 前端插件安装 UI 显示实时进度
- [ ] 安装失败时可查看详细错误日志
- [ ] `go test ./apps/backend/internal/plugin/...` 通过
```

---

### Prompt 2.2 — 多语言插件支持

```
## 任务：添加多语言插件执行器支持

### 背景
当前仅实现了 Python 插件执行器。需要添加 Go 原生插件加载器和进程隔离执行器，以便支持多种语言的插件执行。

### 目标
实现 Go 插件加载器，完善进程隔离执行器，添加执行 Metrics。

### 文件清单

#### 新建文件

1. **`apps/backend/internal/plugin/executors/go_executor.go`** — Go 插件加载器

   - 使用 `plugin.Open()` 加载 `.so` 文件
   - 约定插件导出符号：`Execute(input, config) (output, error)`
   - 类型断言和错误处理
   - 资源管理：插件生命周期追踪

2. **`apps/backend/internal/plugin/executors/metrics.go`** — 执行 Metrics

   ```go
   type ExecutionMetrics struct {
       Duration      time.Duration
       MemoryBytes   uint64
       CPUPercent    float64
       GPUMemoryBytes uint64 // 可选
       OutputSize    int
   }
   ```

   提供 `CollectMetrics(execFn func() error) (*ExecutionMetrics, error)` 包装函数

#### 修改文件

3. **`apps/backend/internal/plugin/executors/process_executor.go`** — 完善进程执行器

   添加功能：
   - 超时控制：`context.WithTimeout`
   - 资源限制：Windows 用 Job Object，Linux 用 cgroups/rlimit
   - 日志收集：实时读取 stdout/stderr，写入缓冲区
   - 标准协议版本协商：握手 → 数据 → 结束

   错误类型：
   - `ErrExecutionTimeout` — 超时错误
   - `ErrOutOfMemory` — OOM 错误
   - `ErrProcessCrash` — 进程崩溃

### 验收标准
- [ ] Go `.so` 插件可通过 `plugin.Open()` 加载并执行
- [ ] 进程隔离执行器支持超时和资源限制
- [ ] 执行 Metrics 正确记录耗时和内存使用
- [ ] 进程崩溃时返回详细错误原因
- [ ] `go build` 通过
```

---

## Phase 3 — Agent 完善与 LLM 集成 (P1)

> **目标**: AI Agent 能真正通过 LLM 理解用户意图，生成并执行工作流
> **预估**: 2 周

---

### Prompt 3.1 — 修复 LLM 流式实现

```
## 任务：实现 Claude 和 Gemini 的流式 API 调用

### 背景
`apps/backend/internal/agent/llm_provider.go` 中 Claude 和 Gemini 的流式方法返回 `"not implemented"` 错误。需要接入真实流式 API。

### 目标
移除占位错误，实现 Claude (Anthropic SDK) 和 Gemini (Google AI SDK) 的 SSE 流式调用。

### 文件清单

#### 修改文件

1. **`apps/backend/internal/agent/llm_provider.go`** — 实现流式方法

   **Claude 流式实现**（约 line 332，移除占位错误）：
   ```go
   func (p *ClaudeProvider) StreamChat(ctx context.Context, req ChatRequest, w http.ResponseWriter) error {
       // 1. 构建 Anthropic Messages API 请求体
       // 2. POST 到 https://api.anthropic.com/v1/messages
       //    设置 Headers: x-api-key, anthropic-version: 2023-06-01
       // 3. 设置 SSE 响应头：Content-Type: text/event-stream
       // 4. 逐行读取 SSE 事件流
       //    - event: content_block_delta → 提取 text delta → SSE 写入 w
       //    - event: message_stop → 写入 [DONE] 标记
       // 5. 错误处理：API 错误 → 写入 error 事件
       // 6. Flush 每个事件
   }
   ```

   **Gemini 流式实现**（约 line 452，移除占位错误）：
   ```go
   func (p *GeminiProvider) StreamChat(ctx context.Context, req ChatRequest, w http.ResponseWriter) error {
       // 1. 构建 Gemini API 请求体
       // 2. POST 到 https://generativelanguage.googleapis.com/v1beta/models/{model}:streamGenerateContent
       // 3. SSE 解析：处理 candidates[0].content.parts[0].text
       // 4. 使用 SSE 格式写入 w
       // 5. 错误处理
   }
   ```

   公共工具函数：
   - `writeSSEEvent(w, event, data)` — 写入 `event: {event}\ndata: {data}\n\n`
   - `parseSSEStream(resp *http.Response, handler func(event string, data []byte) error) error`

2. **`apps/backend/internal/agent/llm_provider_test.go`** — 添加流式测试

   - 创建 Mock HTTP 服务器模拟 Claude/Gemini API
   - 测试场景：
     - 正常流式响应
     - 网络超时
     - API 返回错误
     - 部分内容接收后中断

### 配置要求

确保 `config.yaml` 或环境变量中有：
```bash
LLM_CLAUDE_API_KEY=
LLM_CLAUDE_MODEL=claude-sonnet-4-20250514
LLM_GEMINI_API_KEY=
LLM_GEMINI_MODEL=gemini-2.0-flash
```

### 验收标准
- [ ] `ClaudeProvider.StreamChat()` 不再返回 "not implemented" 错误
- [ ] `GeminiProvider.StreamChat()` 不再返回 "not implemented" 错误
- [ ] 流式响应以正确的 `text/event-stream` 格式输出
- [ ] 每个 token/chunk 被独立 SSE 事件推送
- [ ] 流结束写入 `[DONE]` 标记
- [ ] Mock 测试覆盖正常和异常场景
- [ ] `go test ./apps/backend/internal/agent/...` 通过
```

---

### Prompt 3.2 — Agent 对话流式响应

```
## 任务：实现 Agent 对话流式 SSE 响应

### 背景
Agent Chat API `POST /api/agent/chat` 当前返回完整 JSON 响应。需要改为 SSE 流式响应，让前端能实时看到 LLM Token 流。

### 目标
修改 Agent Handler，支持 SSE 流式输出，前端支持流式增量渲染。

### 文件清单

#### 修改文件

1. **`apps/backend/internal/api/handlers/agent.go`** — 改为 SSE 流式响应

   ```go
   func (h *AgentHandler) HandleChat(c *gin.Context) {
       // 1. 设置 SSE 头
       c.Header("Content-Type", "text/event-stream")
       c.Header("Cache-Control", "no-cache")
       c.Header("Connection", "keep-alive")

       // 2. 解析请求 body
       var req ChatRequest
       if err := c.ShouldBindJSON(&req); err != nil {
           writeSSEError(c.Writer, "invalid_request", err.Error())
           return
       }

       // 3. 获取对应 LLM Provider
       provider := getProvider(req.Provider) // "claude" / "gemini" / "openai"

       // 4. 调用 StreamChat，传入 http.ResponseWriter
       err := provider.StreamChat(c.Request.Context(), req, c.Writer)
       if err != nil {
           writeSSEError(c.Writer, "stream_error", err.Error())
       }
   }
   ```

   SSE 事件类型：

   | 事件 | 数据 | 说明 |
   |------|------|------|
   | `token` | `{"text": "你好"}` | LLM 输出的文本片段 |
   | `action` | `{"type": "planning", "description": "..."}` | Agent 动作阶段 |
   | `done` | `{"reason": "stop"}` | 流结束 |
   | `error` | `{"code": "...", "message": "..."}` | 错误通知 |

2. **`apps/backend/internal/agent/agent.go`** — Agent 动作流式事件

   - 在执行各阶段时通过回调或 channel 推送事件：
     - 阶段：planning → executing → responding
     - 每个阶段携带描述文本
   - 集成到 Chat SSE 流中作为 `event: action`

3. **`apps/desktop/src/pages/AIChat/`** — 前端流式 Chat 对接

   - 修改 `MessageList.vue`：
     - 使用 `EventSource` 或 `fetch + ReadableStream` 接收 SSE
     - 逐 token 增量渲染（不是一次性替换）
     - 添加打字机动画效果（光标闪烁）
   - 修改 `ChatInput.vue`：
     - 发送消息后禁用输入，流完成后恢复
     - 支持中断流（发送 abort 信号）

### 验收标准
- [ ] `POST /api/agent/chat` 返回 `text/event-stream` 格式
- [ ] 前端 AI Chat 页面逐 token 显示流式响应
- [ ] Agent 各动作阶段（规划/执行/回应）显示进度提示
- [ ] 中断按钮能停止流式响应
- [ ] `go build` 通过
- [ ] 前端 `npm run build` 通过
```

---

### Prompt 3.3 — Agent 上下文感知与工作流生成

```
## 任务：使 Agent 感知项目上下文并自动生成工作流

### 背景
Agent 当前只能进行通用对话，缺乏项目上下文感知能力。需要注入当前项目的 Workflow、日志和错误信息，使 Agent 能理解用户项目状态并自动生成工作流。

### 目标
实现项目上下文注入、Agent 动作系统、记忆系统增强和自然语言到工作流的自动生成。

### 文件清单

#### 新建/修改文件

1. **`apps/backend/internal/agent/context.go`** — 项目上下文注入

   ```go
   type ProjectContext struct {
       ProjectID    string
       Name         string
       Workflows    []WorkflowSummary
       RecentLogs   []LogEntry
       RecentErrors []ErrorEntry
       ActiveTasks  []TaskSummary
   }
   ```

   方法：
   - `BuildProjectContext(ctx, projectID string) (*ProjectContext, error)` — 从 DB/LogCenter/Diagnostic 查询
   - `InjectToPrompt(ctx *ProjectContext, originalPrompt string) string` — 格式化为 Markdown 追加到系统提示词

2. **`apps/backend/internal/agent/memory.go`** — 记忆系统增强

   - 增强为基于向量相似度的语义检索：
     - 如果项目已有嵌入服务，调用它
     - 否则使用基于关键词的 TF-IDF 检索（不引入外部依赖）
   - 实现 `SemanticSearch(query string, k int) ([]MemoryItem, error)`

3. **`apps/backend/internal/service/agent_service.go`** — Workflow 自动生成

   实现 `GenerateWorkflowFromNL(ctx, projectID string, naturalLanguage string) (*workflow.Workflow, error)`：

   1. 构建 prompt，包含用户自然语言描述和可用的节点列表
   2. 调用 LLM Chat（非流式）获取 Workflow JSON
   3. 解析 JSON 为 `workflow.Workflow` 结构体
   4. 验证拓扑有效性（环检测、端口匹配）
   5. 返回生成的 Workflow

   系统提示词模板：
   ```
   你是一个 AI 工作流生成助手。根据用户的自然语言描述，生成一个 Workflow JSON。
   可用的节点类型有：[YOLOTrain, YOLOPredict, IF, Loop, Switch, ...]
   每个节点的输入输出端口已在类型定义中指定。
   请严格按照以下 JSON Schema 输出：
   {
     "nodes": [...],
     "edges": [...],
     "config": {...}
   }
   ```

   API 端点：`POST /api/agent/generate-workflow`
   - Request: `{"projectId": "...", "description": "..."}`
   - Response: 生成的 Workflow JSON

### 验收标准
- [ ] Agent Chat 中包含当前项目的工作流/日志/错误摘要
- [ ] Agent 能回答"我当前项目有哪些工作流？"等上下文问题
- [ ] `POST /api/agent/generate-workflow` 接受自然语言描述
- [ ] 生成的 Workflow JSON 语法正确，拓扑有效
- [ ] 前端可在 AI Chat 页面看到"生成工作流"的结果按钮
- [ ] `go test ./apps/backend/internal/agent/...` 通过
```

---

## Phase 4 — 数据管道与实时推送 (P2)

> **目标**: 完善的实时数据流、任务日志和可视化监控
> **预估**: 1-2 周

---

### Prompt 4.1 — 日志系统完善与诊断持久化

```
## 任务：诊断引擎持久化和日志实时搜索

### 背景
`apps/backend/internal/diagnostic/diagnostic.go` 中有 `// TODO: Implement history storage`，诊断结果未持久化。日志系统缺少实时搜索功能。

### 目标
移除 TODO，实现诊断历史持久化；增加日志全文搜索和 WebSocket 推送。

### 文件清单

#### 修改文件

1. **`apps/backend/internal/diagnostic/diagnostic.go`** — 实现持久化

   ```go
   type DiagnosticResult struct {
       ID          string    `gorm:"primaryKey"`
       TaskID      string    `gorm:"index"`
       WorkflowID  string    `gorm:"index"`
       ErrorType   string
       Severity    string    // "error", "warning", "info"
       Summary     string
       Details     string    // JSON 详细分析
       Suggestions string    // JSON 修复建议
       CreatedAt   time.Time
   }
   ```

   新增方法：
   - `SaveResult(ctx, result *DiagnosticResult) error`
   - `GetHistory(ctx, filter DiagnosticFilter) ([]DiagnosticResult, error)`
   - `GetByTaskID(ctx, taskID string) (*DiagnosticResult, error)`

   `DiagnosticFilter` 支持：按 Severity/ErrorType/TaskID/时间范围过滤

   在 `CreateDiagnostic()` 末尾自动调用 `SaveResult()`

2. **`apps/backend/internal/database/models/`** — 数据库迁移

   - 在 AutoMigrate 调用中添加 `DiagnosticResult` 模型

3. **`apps/backend/internal/logcenter/logcenter.go`** — 日志实时搜索

   ```go
   func Search(ctx, query LogQuery) (*LogSearchResult, error)
   ```

   - 全文搜索日志内容
   - 过滤条件：level (debug/info/warn/error)、时间范围、模块名、taskID
   - 分页返回

4. **`apps/backend/internal/api/ws/hub.go`** — 日志推送

   - LogCenter 事件通过 EventBus 转发到 WebSocket Hub
   - 日志事件推送到房间 `task:{taskID}`

5. **`apps/desktop/src/components/logs/LogViewer.vue`** — 前端日志对接

   - 显示实时日志流
   - 搜索框（支持全文搜索）
   - 过滤条件：level / module / 时间范围
   - 自动滚动（可锁定/解锁）

### 验收标准
- [ ] `DiagnosticResult.SaveResult()` 保存到数据库
- [ ] `GetHistory()` 返回历史诊断记录，支持过滤
- [ ] 日志支持全文搜索和模块/级别过滤
- [ ] LogViewer 显示实时日志推送
- [ ] `go test` 通过
```

---

## Phase 5 — 基础设施与生产化 (P2)

> **目标**: 开发者体验优化、CI/CD、Docker 化
> **预估**: 1-2 周

---

### Prompt 5.1 — 基础设施文件

```
## 任务：创建缺失的基础设施文件

### 背景
ROADMAP.md 引用了根目录 Makefile、.env.example，但文件不存在。Git 仓库可能缺少 .gitignore。

### 目标
创建 Makefile、.env.example、.gitignore 三个文件。

### 文件清单

#### 新建文件

1. **`Makefile`** — 根目录构建文件

   包含以下命令：

   | 命令 | 说明 |
   |------|------|
   | `make dev` | 启动所有开发服务（提示信息）|
   | `make dev-backend` | 启动 Go 后端 |
   | `make dev-frontend` | 启动前端开发服务器 |
   | `make dev-engine` | 启动 Python AI Engine |
   | `make build` | 构建所有 |
   | `make build-backend` | 构建后端 |
   | `make build-frontend` | 构建前端 |
   | `make test` | 运行所有测试 |
   | `make test-backend` | 运行后端测试 |
   | `make test-engine` | 运行 Engine 测试 |
   | `make docker-build` | 构建 Docker 镜像 |
   | `make docker-up` | 启动 Docker 环境 |
   | `make docker-down` | 停止 Docker 环境 |
   | `make clean` | 清理构建产物 |
   | `make help` | 显示帮助 |

2. **`.env.example`** — 环境变量模板

   ```bash
   # 应用配置
   APP_ENV=development
   APP_PORT=8080
   APP_LOG_LEVEL=debug

   # 数据库
   DB_PATH=data/aistudio.db

   # Python Engine
   ENGINE_URL=http://localhost:8082
   ENGINE_TIMEOUT=30s

   # LLM Providers
   LLM_OPENAI_API_KEY=
   LLM_OPENAI_MODEL=gpt-4
   LLM_CLAUDE_API_KEY=
   LLM_CLAUDE_MODEL=claude-sonnet-4-20250514
   LLM_GEMINI_API_KEY=
   LLM_GEMINI_MODEL=gemini-2.0-flash

   # JWT
   JWT_SECRET=change-this-to-a-random-secret
   JWT_EXPIRATION=24h

   # Plugins
   PLUGINS_DIR=plugins
   ```

3. **`.gitignore`** — Git 忽略规则

   覆盖以下类别：
   - Go 构建产物：`build/`, `*.exe`, `*.test`, `*.out`, `vendor/`
   - Node 依赖：`node_modules/`, `dist/`, `.cache/`
   - Python 缓存：`__pycache__/`, `*.py[cod]`, `env/`, `venv/`
   - 环境文件：`.env`, `.env.local`
   - 数据库：`*.db`, `*.sqlite`, `data/`
   - IDE 配置：`.vscode/`, `.idea/`, `*.swp`
   - 系统文件：`.DS_Store`, `Thumbs.db`

### 验收标准
- [ ] `make help` 显示所有命令
- [ ] `make dev-backend` 能启动后端
- [ ] `.env.example` 提供完整配置模板
- [ ] `.gitignore` 覆盖 Go/Node/Python/IDE/OS 文件
```

---

### Prompt 5.2 — Docker 开发环境

```
## 任务：创建 Docker 开发环境

### 背景
项目缺少 Docker 化配置，新开发者上手需手动安装 Go/Node/Python。

### 目标
创建 docker-compose.yml 和三个 Dockerfile，实现一键启动开发环境。

### 文件清单

#### 新建文件

1. **`docker-compose.yml`** — Docker Compose

   三个服务：

   | 服务 | 端口 | 构建目录 | 说明 |
   |------|------|---------|------|
   | `backend` | 8080 | `apps/backend/Dockerfile` | Go 后端 |
   | `engine` | 8082 | `Engine/Dockerfile` | Python AI Engine |
   | `frontend` | 3000:80 | `apps/desktop/Dockerfile` | Nginx 前端 |

   配置要点：
   - `volumes` 挂载源码目录用于热更新
   - `depends_on` 确保启动顺序
   - engine 服务添加 GPU 资源配置（`nvidia`）

2. **`apps/backend/Dockerfile`** — Go 后端多阶段构建

   - Stage 1 (builder)：`golang:1.22-alpine` → `go mod download` → `go build`
   - Stage 2 (runtime)：`alpine:3.19` → 复制构建产物 → `CMD ["./server"]`
   - 启用 `CGO_ENABLED=0` 静态编译

3. **`Engine/Dockerfile`** — Python 引擎镜像

   - Base: `python:3.11-slim`
   - 安装系统依赖 `gcc`
   - `pip install -r requirements.txt`
   - 复制 Engine 代码
   - `CMD ["python", "server.py", "--port", "8082"]`

4. **`apps/desktop/Dockerfile`** — 前端 Nginx 镜像

   - Stage 1 (builder)：`node:20-alpine` → `npm ci` → `npm run build`
   - Stage 2 (runtime)：`nginx:alpine` → 复制 dist → 复制 nginx.conf

### 验收标准
- [ ] `docker-compose build` 成功
- [ ] `docker-compose up` 启动所有服务
- [ ] 后端在 localhost:8080 可访问
- [ ] Engine 在 localhost:8082 可访问
- [ ] 前端在 localhost:3000 可访问
- [ ] `docker-compose logs` 显示各服务日志
```

---

### Prompt 5.3 — CI/CD 配置

```
## 任务：创建 GitHub Actions CI/CD 配置

### 背景
项目没有自动化构建、测试和部署流程。

### 目标
在 `.github/workflows/` 下创建 4 个 CI 配置文件。

### 文件清单

#### 新建文件

1. **`.github/workflows/backend-ci.yml`** — 后端 CI

   ```yaml
   name: Backend CI
   on:
     push:
       branches: [main, develop]
       paths: ['apps/backend/**', 'go.mod', 'go.sum']
     pull_request:
       branches: [main]
       paths: ['apps/backend/**']

   jobs:
     build:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version: '1.22'
             cache: true
         - uses: golangci/golangci-lint-action@v4
           with:
             working-directory: apps/backend
         - run: cd apps/backend && go test ./... -v -count=1 -race -coverprofile=coverage.out
         - run: cd apps/backend && go build ./cmd/
         - uses: codecov/codecov-action@v3
   ```

2. **`.github/workflows/frontend-ci.yml`** — 前端 CI

   - steps: checkout → setup-node → npm ci → npm run typecheck → npm run lint → npm run build

3. **`.github/workflows/engine-ci.yml`** — Engine CI

   - steps: checkout → setup-python → pip install → python -m pytest → flake8

4. **`.github/workflows/integration.yml`** — 集成测试

   - steps: checkout → docker-compose up → wait-for-it → go test ./tests/e2e/...

### 验收标准
- [ ] `.github/workflows/` 目录包含 4 个 yml 文件
- [ ] 后端 CI：lint → test → build
- [ ] 前端 CI：typecheck → lint → build
- [ ] Engine CI：pytest → lint
- [ ] 集成测试：启动 docker-compose → 运行 E2E 测试
- [ ] 每个 workflow 有正确的 paths 触发过滤
```

---

### Prompt 5.4 — 代码生成器注册统一

```
## 任务：在 main.go 注册全部 8 个代码生成器

### 背景
`apps/backend/cmd/main.go` 中的编译器注册处有 TODO，当前只注册了 Python 生成器。

### 目标
移除 TODO，注册全部 8 个代码生成器适配器。

### 修改文件

**`apps/backend/cmd/main.go`**

当前代码:
```go
compilerEngine.RegisterGenerator(compilerPython.NewGenerator())
// TODO: Register more generators
```

改为:
```go
compilerEngine.RegisterGenerator(compilerPython.NewGenerator())
compilerEngine.RegisterGenerator(compilerMATLAB.NewGenerator())
compilerEngine.RegisterGenerator(compilerROS2.NewGenerator())
compilerEngine.RegisterGenerator(compilerDocker.NewGenerator())
compilerEngine.RegisterGenerator(compilerSTM32.NewGenerator())
compilerEngine.RegisterGenerator(compilerCPP.NewGenerator())
compilerEngine.RegisterGenerator(compilerUnity.NewGenerator())
compilerEngine.RegisterGenerator(compilerJava.NewGenerator())
```

### 验收标准
- [ ] 8 个代码生成器全部在 main.go 注册
- [ ] `go build` 通过
- [ ] 所有生成器的 `NewGenerator()` 函数签名一致
```

---

## Phase 6 — E2E 联调与测试 (P3)

> **目标**: 完整端到端流程通过测试，bug 修复，文档对齐
> **预估**: 持续进行

---

### Prompt 6.1 — 关键 E2E 流程

```
## 任务：编写关键端到端测试流程

### 背景
需要验证从用户到工作流执行完成的全链路。

### 目标
编写 6 个关键 E2E 测试场景，覆盖完整产品流程。

### 测试文件

#### 新建文件

1. **`tests/e2e/auth_flow_test.go`** — 认证流程

   测试：用户注册 → 登录 → 创建项目
   验证：JWT token 返回、项目 CRUD

2. **`tests/e2e/workflow_persistence_test.go`** — 工作流持久化

   测试：创建 Workflow → 保存 JSON → 重新加载
   验证：节点和边完整

3. **`tests/e2e/workflow_execution_test.go`** — 工作流执行

   测试：拖拽节点 → 连接 → 配置 → 运行
   验证：节点依次执行、结果非空

4. **`tests/e2e/realtime_logs_test.go`** — 实时日志

   测试：建立 WebSocket → 运行工作流 → 接收日志

5. **`tests/e2e/task_result_test.go`** — 任务结果

   测试：任务完成/失败 → 验证终态 → WebSocket 通知

6. **`tests/e2e/plugin_lifecycle_test.go`** — 插件生命周期

   测试：安装 → 使用 → 卸载

### 验收标准
- [ ] 6 个 E2E 测试场景全部通过
- [ ] 测试可独立运行（无相互依赖）
- [ ] 测试环境自动初始化（清空数据库）
- [ ] `make test-e2e` 一键运行
```

---

### Prompt 6.2 — 测试覆盖提升

```
## 任务：补全核心模块单元测试

### 背景
当前 206 个测试函数，核心模块覆盖率低。需要为新增和原有代码补充单元测试。

### 目标
为 8 个核心模块补充测试，总计增加约 100 个测试函数。

### 测试目标

| 模块 | 新增测试数 | 覆盖内容 |
|------|-----------|---------|
| Workflow DAG | +20 | 拓扑排序、环检测、端口校验 |
| Engine Bridge (HTTP Client) | +10 | Mock Server、超时、重试 |
| Plugin Manager | +10 | Execute、Install、错误处理 |
| Agent System | +15 | Planner、Executor、Memory |
| LLM Provider | +8 | 各 Provider Chat + Stream |
| MCP Runtime | +10 | Connect、CallTool、ListTools |
| WebSocket Hub | +8 | 连接管理、广播、房间 |
| Python Engine (pytest) | +20 | YOLO、Dataset、Trainer |

### 验收标准
- [ ] 测试函数总数从 206 提升至 300+
- [ ] 新增模块覆盖率 ≥ 60%
- [ ] `go test ./... -count=1` 全部通过
- [ ] `cd Engine && python -m pytest` 全部通过
```

---

### Prompt 6.3 — 文档对齐

```
## 任务：文档对齐 — API 文档、README、ADR

### 背景
随着代码变更，架构文档可能与实际实现不一致。

### 目标
审查并更新 API 文档、README 和 ADR 文档。

### 任务清单

1. **API 文档对齐**
   - 对比 `Docs/api/` 中所有路由文档与实际 `apps/backend/internal/api/router.go`
   - 更新遗漏或不一致的路由
   - 确保请求/响应 schema 匹配

2. **更新 README**
   - 构建步骤（更新为使用 Makefile）
   - 依赖版本（Go 1.22+, Node 20+, Python 3.11+）
   - 配置说明（引用 .env.example）
   - Docker 开发流程

3. **ADR 对齐**
   - 检查 `Docs/ADR/ADR-*.md` 中的决策
   - 对与实现不一致的添加补充说明

### 验收标准
- [ ] Docs/api/ 与实际 router.go 一致
- [ ] README 包含真实构建步骤
- [ ] ADR 文档与当前实现无重大冲突
```

---

## 附录 A：文件索引

### 新建文件总览

```
Phase 1:
  apps/backend/internal/engine/                    # Go↔Python 通信
  apps/backend/internal/engine/client.go           # EngineClient 接口
  apps/backend/internal/engine/http_client.go      # HTTP 客户端实现
  apps/backend/internal/engine/config.go           # Engine 配置
  apps/backend/internal/engine/engine.go           # 包入口
  apps/backend/internal/api/ws/                    # WebSocket 包
  apps/backend/internal/api/ws/hub.go              # 连接管理
  apps/backend/internal/api/ws/client.go           # 连接读写
  apps/backend/internal/api/ws/messages.go         # 消息格式
  apps/backend/internal/workflow/executors/        # 节点执行器
  apps/backend/internal/workflow/executors/yolo_train.go
  apps/backend/internal/workflow/executors/yolo_predict.go
  apps/backend/internal/workflow/executors/condition.go
  apps/backend/internal/workflow/executors/loop.go
  apps/backend/internal/workflow/executors/switch.go
  apps/backend/internal/workflow/executors/retry.go
  apps/backend/internal/workflow/executors/data_loader.go
  apps/backend/internal/workflow/executors/nlp.go
  apps/backend/internal/workflow/executors/http_request.go
  apps/backend/internal/plugin/installer.go        # 插件安装

Phase 2:
  apps/backend/internal/plugin/executors/          # 插件执行器
  apps/backend/internal/plugin/executors/python_executor.go
  apps/backend/internal/plugin/executors/process_executor.go
  apps/backend/internal/plugin/executors/go_executor.go
  apps/backend/internal/plugin/executors/metrics.go

Phase 5:
  Makefile                                         # 根目录构建文件
  .env.example                                     # 环境变量模板
  .gitignore                                       # Git 忽略规则
  docker-compose.yml                               # 开发环境
  apps/backend/Dockerfile                          # 后端镜像
  Engine/Dockerfile                                # Python 引擎镜像
  apps/desktop/Dockerfile                          # 前端镜像
  .github/workflows/                              # CI/CD
    backend-ci.yml
    frontend-ci.yml
    engine-ci.yml
    integration.yml

Phase 6:
  tests/e2e/                                      # E2E 测试
    auth_flow_test.go
    workflow_persistence_test.go
    workflow_execution_test.go
    realtime_logs_test.go
    task_result_test.go
    plugin_lifecycle_test.go
```

### 修改文件总览

```
Phase 1:
  apps/backend/internal/workflow/builtin_nodes.go    # 节点 Factory 指向真实执行器
  apps/backend/cmd/main.go                           # 注册 Engine Bridge, WebSocket
  apps/backend/internal/api/router.go                # WebSocket 路由更新
  apps/backend/internal/service/service.go           # 添加 EngineService
  apps/backend/internal/plugin/manager.go            # 添加 Execute(), RegisterExecutor()

Phase 3:
  apps/backend/internal/agent/llm_provider.go        # Claude/Gemini 流式
  apps/backend/internal/agent/context.go             # 项目上下文
  apps/backend/internal/agent/memory.go              # 语义检索
  apps/backend/internal/api/handlers/agent.go        # SSE 流式
  apps/backend/internal/service/agent_service.go     # 工作流生成

Phase 4:
  apps/backend/internal/diagnostic/diagnostic.go     # 移除 TODO，实现持久化
  apps/backend/internal/database/models/             # 添加诊断记录模型
  apps/backend/internal/logcenter/logcenter.go       # 日志搜索

Phase 5:
  apps/backend/cmd/main.go                           # 注册全部 8 个生成器

Phase 6:
  Docs/api/*                                         # API 文档对齐
  README.md                                          # 更新构建步骤
  Docs/ADR/ADR-*.md                                  # ADR 对齐
```

---

## 附录 B：验收标准速查表

### Phase 1

| Step | 关键验收点 | 优先级 |
|------|-----------|--------|
| 1.1 | Engine Bridge 编译通过，Health() 可调用 Python Engine | P0 |
| 1.2 | 10 个节点执行器返回真实数据，非占位符 | P0 |
| 1.3 | WebSocket 推送 Task 状态到前端 | P0 |
| 1.4 | 插件可 Execute() 调用，安装管道可用 | P0 |
| 1.5 | 4 个集成测试场景全部通过 | P0 |

### Phase 2

| Step | 关键验收点 | 优先级 |
|------|-----------|--------|
| 2.1 | 依赖检查/模型下载/安装 UI 完整 | P1 |
| 2.2 | Go 插件加载 + 多语言执行器 + Metrics | P1 |

### Phase 3

| Step | 关键验收点 | 优先级 |
|------|-----------|--------|
| 3.1 | Claude/Gemini 流式实现，不再返回 "not implemented" | P1 |
| 3.2 | Agent Chat SSE 流式，前端逐 token 显示 | P1 |
| 3.3 | Agent 感知项目上下文，从自然语言生成 Workflow | P1 |

### Phase 4

| Step | 关键验收点 | 优先级 |
|------|-----------|--------|
| 4.1 | 诊断持久化 + 日志搜索 + 实时推送 | P2 |

### Phase 5

| Step | 关键验收点 | 优先级 |
|------|-----------|--------|
| 5.1 | Makefile/.env.example/.gitignore 就位 | P2 |
| 5.2 | docker-compose up 一键启动 | P2 |
| 5.3 | CI/CD 管道路由通过 | P2 |
| 5.4 | 8 个代码生成器全部注册 | P2 |

### Phase 6

| Step | 关键验收点 | 优先级 |
|------|-----------|--------|
| 6.1 | 6 个 E2E 场景全部通过 | P3 |
| 6.2 | 测试函数达 300+，覆盖率 ≥ 60% | P3 |
| 6.3 | 文档与实现一致 | P3 |

---

## 附录 C：常见问题

### 如何调试 Prompt 执行失败？

1. 确认文件路径在项目根目录下是否正确
2. 检查 Go 模块导入路径是否与 `go.mod` 中的 `module` 一致
3. 对照已有代码的命名风格（驼峰/下划线/缩写规范）
4. 先 `go build ./...` 确保编译通过，再运行测试

### 如何跳过已完成的 Step？

查验收标准中的 Checkbox，已勾选的步骤可以跳过。如果某个 Step 的代码已部分实现，直接跳到下一步。

### 如何并行执行？

Phase 1 内部 Step 1.1 ~ 1.4 可以并行。每个 Prompt 可以交给独立的 Claude Code 会话执行。确保在代码合并时解决导入路径冲突。

### 执行顺序优先级

- P0 标记的验收点必须在本 Step 完成
- P1 标记的验收点可以延迟到本 Phase 结束前完成
- 如某个 Step 遇到阻塞（如依赖的 API 不可用），可以挂起该 Step 先做其他 Steps
