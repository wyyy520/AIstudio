# AIStudio 项目实施路线图

> **版本**: v1.0  
> **日期**: 2026-07-13  
> **状态**: 规划阶段  
> **目标**: 将所有架构文档中的设计变为可运行的端到端产品

---

## 目录

1. [项目全景状态](#1-项目全景状态)
2. [核心缺失总览](#2-核心缺失总览)
3. [Phase 1 — 引擎打通：让工作流真正运行 (P0)](#3-phase-1--引擎打通让工作流真正运行-p0)
4. [Phase 2 — 插件系统：让节点可扩展 (P1)](#4-phase-2--插件系统让节点可扩展-p1)
5. [Phase 3 — Agent 完善与 LLM 集成 (P1)](#5-phase-3--agent-完善与-llm-集成-p1)
6. [Phase 4 — 数据管道与实时推送 (P2)](#6-phase-4--数据管道与实时推送-p2)
7. [Phase 5 — 基础设施与生产化 (P2)](#7-phase-5--基础设施与生产化-p2)
8. [Phase 6 — E2E 联调与测试 (P3)](#8-phase-6--e2e-联调与测试-p3)
9. [参考文档](#9-参考文档)

---

## 1. 项目全景状态

### 1.1 完成度总览

```
层 级              模块                     完成度    状态
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
后端基础           配置/数据库/事件总线/日志    100%     ✅ 已完成
认证系统           JWT/API Key/Session/权限    100%     ✅ 已完成
                   用户 CRUD / 配额管理

API 层             37 个路由组 / 60+ 路由      100%     ✅ 已完成
                   中间件 (CORS/限流/恢复)
                   统一响应封装

Workflow DSL       类型定义 / 解析/序列化      100%     ✅ 已完成
                   Schema 迁移 / 文件 IO

Workflow DAG       拓扑排序 (Kahn)             100%     ✅ 已完成
                   上下流分析 / 分层分组
                   环检测 / 端口校验

Workflow 节点定义   40+ 内置节点 (IF/Loop/     100%     ✅ 已完成
                   YOLO/BERT/LLM/ASR/...)
                   端口 / 数据类型 / 约束

Workflow 执行引擎   Engine.RunWorkflow()        90%     ✅ 已完成
                   NodeRegistry / Factory       ⚠️ 节点均返回占位结果
                   按拓扑序逐节点执行

Task 调度系统       Manager/Queue/WorkerPool    100%    ✅ 已完成
                   状态机 / EventBus

代码编译器          Compiler + Generator 接口   90%     ✅ 已完成
                   8 个目标适配器               ⚠️ 仅 Python 在 main.go 注册
                   (Python/MATLAB/ROS2/...)

Runtime 执行器      本地执行 / Bundle 管理       80%    ✅ 已完成
                   进程生命周期

Python AI 引擎     HTTP Server (server.py)     70%     ✅ 已存在
                   YOLO 训练/预测                ⚠️ HTTP-only，无 gRPC
                   数据集/训练器/推理/结果导出   ⚠️ 仅 YOLO，无 NLP/语音

Plugin Manager V2   Manifest 发现/解析          70%     ✅ 已存在
                   注册表/启用禁用               ❌ 无 Execute() 方法
                   桥接到 Workflow Registry      ❌ 无安装管道

MCP 运行时          Client/Server/Registry      70%     ✅ 已存在
                                               ⚠️ 仅本地/Mock 模式

Agent 系统          规划→执行→回应 全流程        80%    ✅ 已存在
                   ToolRegistry / Memory         ❌ 需对接真实 LLM
                   6 个 LLM Provider              ⚠️ Claude/Gemini 流式未实现

Environment         环境检测/依赖管理            80%    ✅ 已存在
                   安装/修复

Diagnostic          错误分析框架                 60%    ✅ 已存在
                                               ⚠️ 历史存储未实现
Frontend (Vue3)     7个完整页面 / 170+文件       90%    ✅ 已完成
                   11个Pinia Store / 路由
                   完整UI组件库

Tauri 桌面壳        Rust 配置 / 图标 / 打包      90%    ✅ 已完成
                   子进程管理

文档体系           架构/API/UI/Plugin-SDK/ADR   90%    ✅ 已完成
                   安全/性能/集成审查报告
```

### 1.2 分层架构

```
┌─────────────────────────────────────────────────────────────────┐
│                   Tauri Desktop (Vue3 + TS)                       │
│   Dashboard │ AI Chat │ Workflow │ PluginStore │ Logs │ Settings │
│   11 Pinia Stores │ WebSocket Client │ API Client (12 modules)   │
├─────────────────────────────────────────────────────────────────┤
│                     Go Backend (Gin + GORM)                       │
│                                                                    │
│   ┌─────────────────────────────────────────────────────────┐    │
│   │                  API Router (37 route groups)             │    │
│   │   Auth/User/Project/Workflow/Task/Plugin/Agent/MCP/Env   │    │
│   │   Logs/Settings/ErrorAnalysis/WebSocket                   │    │
│   └───────────────────────────┬──────────────────────────────┘    │
│                               │                                     │
│   ┌──────────┐ ┌─────────────┴──────┐ ┌──────────┐ ┌──────────┐ │
│   │Workflow  │ │   Compiler        │ │  Agent   │ │  Plugin  │ │
│   │ Engine   │ │   8 Generators    │ │  6 LLM   │ │  V2 Mgr │ │
│   │(DAG+Exec)│ │  Python/MATLAB... │ │Provider  │ │(No Exec) │ │
│   └────┬─────┘ └──────┬───────────┘ └────┬─────┘ └────┬─────┘ │
│        │              │                  │             │        │
│        └──────┬───────┴──────────┬───────┴─────────────┘        │
│          Task Scheduler │ MCP Runtime │ Diagnostic Engine       │
├──────────────────────────────────────────────────────────────┤
│                  Python AI Engine (HTTP)                        │
│   YOLO Train │ YOLO Predict │ Dataset │ Trainer │ Result       │
│   ❌ 无 gRPC  ❌ 无 NLP  ❌ 无 Speech                           │
├──────────────────────────────────────────────────────────────┤
│                     Plugins (Declared Only)                      │
│   ❌ 无 Execute ❌ 无 Install                                    │
├──────────────────────────────────────────────────────────────┤
│              Storage: SQLite / Filesystem                        │
└──────────────────────────────────────────────────────────────┘
```

---

## 2. 核心缺失总览

### P0 — 立刻需要（影响端到端可用性）

| # | 缺失 | 影响 | 涉及文件 |
|---|------|------|----------|
| 1 | **Workflow 节点真实执行** | 40+ 节点全是 `noOpExecutor`，工作流"运行"返回占位结果 | `internal/workflow/builtin_nodes.go` |
| 2 | **Go ↔ Python 通信桥梁** | Engine HTTP 服务器可用但 Go 端未调用它 | 新建 `internal/engine/` |
| 3 | **插件执行系统** | Plugin V2 能发现注册但 `manager.go` 无 `Execute()` | `internal/plugin/manager.go` |
| 4 | **WebSocket 实时推送** | Task EventBus → 前端管道未打通 | `internal/api/handlers/websocket.go` |

### P1 — 短期内需要

| # | 缺失 | 影响 | 涉及文件 |
|---|------|------|----------|
| 5 | **gRPC 协议定义** | Python 引擎只有 HTTP，无法流式推理 | 新建 `proto/` |
| 6 | **Python gRPC 服务端** | 无法利用 gRPC 双向流做 LLM Token 输出 | 新建 `Engine/server_grpc.py` |
| 7 | **Go gRPC 客户端** | Engine Bridge 不存在 | 新建 `internal/engine/bridge.go` |
| 8 | **Claude/Gemini 流式** | 返回错误 "not implemented" | `internal/agent/llm_provider.go` |
| 9 | **诊断引擎持久化** | `// TODO: Implement history storage` | `internal/diagnostic/diagnostic.go` |
| 10 | **根目录 Makefile** | README 引用但不存在 | 新建 `Makefile` |
| 11 | **技能模板扩展** | 只有 `yolo_detection.json` | `internal/skill/templates/` |

### P2 — 优化与完善

| # | 缺失 | 影响 | 涉及文件 |
|---|------|------|----------|
| 12 | **CI/CD 配置** | 无自动化构建/测试/部署 | 新建 `.github/workflows/` |
| 13 | **.env 示例文件** | 配置加载器引用但不存在 | 新建 `.env.example` |
| 14 | **测试覆盖率** | 206 个测试函数，覆盖率低 | 各模块 `*_test.go` |
| 15 | **Docker 开发环境** | 缺少 docker-compose 配置 | 新建 `docker-compose.yml` |
| 16 | **代码生成器统一注册** | 8 个适配器，仅 Python 在 `main.go` 注册 | `apps/backend/cmd/main.go` |

---

## 3. Phase 1 — 引擎打通：让工作流真正运行 (P0)

> **目标**：用户创建项目 → 设计工作流 → 节点能调用真实的 Python 引擎 → 看到真实结果  
> **预估**：2-3 周  
> **并行度**：Step 1.1 ~ 1.4 可以并行

### 3.1 Step 1.1 — 实现 Go→Python HTTP Bridge

Python Engine 已有 HTTP Server (`Engine/server.py`)，Go 端需创建客户端调用它。

#### 具体任务

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 1.1.1 | 创建 Engine Bridge 包 | `apps/backend/internal/engine/` (新建) | 专门负责 Go ↔ Python 通信 |
| 1.1.2 | 定义 EngineClient 接口 | `apps/backend/internal/engine/client.go` | `Infer()`, `Train()`, `Health()`, `LoadModel()` 等 |
| 1.1.3 | 实现 HTTP Engine Client | `apps/backend/internal/engine/http_client.go` | 调用 `Engine/server.py` 的 `/task` 和 `/health` 端点 |
| 1.1.4 | 定义配置 | `apps/backend/internal/engine/config.go` | Engine URL, 超时, 重试策略 |
| 1.1.5 | 注册到 Service Container | `apps/backend/internal/service/service.go` | 添加 `EngineService()` 到 Container |

#### 参考

- Python Engine HTTP 端点：`Engine/server.py` line 43-56 → `/health` (GET), `/task` (POST)
- 现有 Service 模式：`apps/backend/internal/service/runtime_service.go`
- Container 注册：`apps/backend/internal/service/service.go`

---

### 3.2 Step 1.2 — 实现 Workflow 节点真实执行器

将 40+ 个 `noOpExecutor` 中的关键节点替换为真实引擎调用。

#### 具体任务

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 1.2.1 | 创建节点执行器目录 | `apps/backend/internal/workflow/executors/` (新建) | 存放节点执行逻辑 |
| 1.2.2 | 实现 YOLO 训练执行器 | `executors/yolo_train.go` | 调用 EngineClient.Train() |
| 1.2.3 | 实现 YOLO 推理执行器 | `executors/yolo_predict.go` | 调用 EngineClient.Infer() |
| 1.2.4 | 实现 IF 条件节点 | `executors/condition.go` | 解析 Expression，返回 true/false 分支 |
| 1.2.5 | 实现 Loop 循环节点 | `executors/loop.go` | 维护迭代计数器，路由数据 |
| 1.2.6 | 实现 Switch 分支节点 | `executors/switch.go` | 多路分支选择 |
| 1.2.7 | 实现 Retry 重试节点 | `executors/retry.go` | 失败重试 + 退避策略 |
| 1.2.8 | 修改 `BuiltInNodeDefinitions()` | `internal/workflow/builtin_nodes.go` | 将 Factory 指向真实执行器而非 noOp |
| 1.2.9 | 实现 Data Loader 节点 | `executors/data_loader.go` | 读取本地文件/数据集 |
| 1.2.10 | 实现文本处理节点 | `executors/nlp.go` | 调用 LLM Provider |

#### 关键变更

```go
// 当前 — builtin_nodes.go line 49-58
func noOpExecutor(...) { return mock result }

// 改为
func yoloTrainExecutor(engine *engine.Client) ExecutableNode {
    return executableFunc(func(ctx, inputs, config) (map[string]interface{}, error) {
        return engine.Train(ctx, TrainRequest{
            Dataset: inputs["dataset"],
            Config:  config,
        })
    })
}
```

---

### 3.3 Step 1.3 — 实现 WebSocket 实时状态推送

Task EventBus 已有事件系统，需推送到前端的 WebSocket 连接。

#### 具体任务

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 1.3.1 | 实现 WebSocket Hub | `apps/backend/internal/api/ws/hub.go` (新建) | 管理连接池，房间订阅 |
| 1.3.2 | 实现 WebSocket Client | `apps/backend/internal/api/ws/client.go` (新建) | 单个连接读写 |
| 1.3.3 | 订阅 Task EventBus | `apps/backend/cmd/main.go` | `setupEventSubscriptions()` 中添加 WebSocket 转发 |
| 1.3.4 | 定义推送消息格式 | `apps/backend/internal/api/ws/messages.go` (新建) | `TaskStatus`, `NodeStatus`, `NodeLog`, `TaskDone` |
| 1.3.5 | 更新 WebSocket 路由 | `apps/backend/internal/api/router.go` | 连接到新的 Hub |
| 1.3.6 | 前端 WebSocket 客户端对接 | `apps/desktop/src/api/websocket.ts` | 确保前端正确处理推送事件 |

#### 消息格式

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

---

### 3.4 Step 1.4 — 实现插件执行系统

Plugin V2 Manager 能发现和注册插件，但无法执行。需添加 Execute 方法和安装管道。

#### 具体任务

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 1.4.1 | 定义 PluginExecutor 接口 | `apps/backend/internal/plugin/interfaces.go` | `Execute(ctx, input, config) → output, error` |
| 1.4.2 | 实现 Python 插件执行器 | `apps/backend/internal/plugin/executors/python_executor.go` (新建) | 通过 `os/exec` 调用 Python 脚本 |
| 1.4.3 | 实现进程隔离执行器 | `apps/backend/internal/plugin/executors/process_executor.go` (新建) | 独立进程，stdin/stdout JSON 通信 |
| 1.4.4 | 在 Manager 中注册执行器 | `apps/backend/internal/plugin/manager.go` | `RegisterExecutor(language, executor)` |
| 1.4.5 | 实现 Execute() 方法 | `apps/backend/internal/plugin/manager.go` | `func (m *Manager) Execute(ctx, pluginName, input, config)` |
| 1.4.6 | 实现插件安装管道 | `apps/backend/internal/plugin/installer.go` (新建) | pip install / 模型下载 / 依赖检查 |
| 1.4.7 | 添加安装 API 路由 | `apps/backend/internal/api/router.go` | `POST /api/plugins/install`, `DELETE /api/plugins/:name` |

#### 现有参考

- `apps/backend/internal/plugin/models.go` — Plugin struct 定义
- `apps/backend/internal/plugin/registry.go` — 注册表
- `apps/backend/internal/plugin/interfaces.go` — 现有接口

---

### 3.5 Phase 1 验收标准

```
□ 启动后端后，Go 能成功调用 Python Engine 的 /health 端点
□ 创建一个 YOLO 训练工作流 → 运行 → 节点状态显示 running/success
□ WebSocket 实时推送节点状态到前端
□ IF/Switch/Loop/Retry 控制节点返回真实结果（非占位）
□ 插件可被 Execute() 调用（Python 子进程模式）
□ 前端 Workflow Console 面板显示实时日志
```

---

## 4. Phase 2 — 插件系统：让节点可扩展 (P1)

> **目标**：第三方开发者能编写、安装、发布插件  
> **预估**：2 周  
> **前置**：Phase 1 Step 1.4

### 4.1 Step 2.1 — 完善插件安装流程

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 2.1.1 | 实现依赖检查 | `apps/backend/internal/plugin/installer.go` | Python 版本 / CUDA / 系统依赖 |
| 2.1.2 | 实现 pip 安装 | `apps/backend/internal/plugin/installer.go` | `pip install -r requirements.txt` |
| 2.1.3 | 实现模型权重下载 | `apps/backend/internal/plugin/installer.go` | 下载到 `models/` 目录 |
| 2.1.4 | 实现 install.py 执行 | `apps/backend/internal/plugin/installer.go` | 自定义安装脚本 |
| 2.1.5 | 安装状态追踪 | `apps/backend/internal/plugin/installer.go` | 进度 / 日志 / 错误报告 |
| 2.1.6 | 前端安装 UI 对接 | `apps/desktop/src/components/plugin/` | 确保 PluginInstallTask 正确显示进度 |

---

### 4.2 Step 2.2 — 多语言插件支持

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 2.2.1 | Go 插件加载器 | `apps/backend/internal/plugin/executors/go_executor.go` (新建) | `plugin.Open()` 方式加载 |
| 2.2.2 | 进程隔离执行器完善 | `apps/backend/internal/plugin/executors/process_executor.go` | 超时 / 资源限制 / 日志 |
| 2.2.3 | 插件执行 Metrics | `apps/backend/internal/plugin/executors/metrics.go` (新建) | 耗时 / 内存 / 设备 |

---

### 4.3 Phase 2 验收标准

```
□ 可通过 API 安装一个 Python 插件（pip install + 模型下载）
□ 已安装的插件可在工作流中拖拽使用
□ 插件执行有完整的错误报告
□ 插件卸载会清理相关文件
□ 前端插件商店页面能展示安装/卸载进度
```

---

## 5. Phase 3 — Agent 完善与 LLM 集成 (P1)

> **目标**：AI Agent 能真正通过 LLM 理解用户意图，生成并执行工作流  
> **预估**：2 周

### 5.1 Step 3.1 — 修复 LLM 流式实现

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 3.1.1 | 实现 Claude 流式 | `apps/backend/internal/agent/llm_provider.go` | 移除 `return fmt.Errorf("...")`，实现真实 SSE 流式 |
| 3.1.2 | 实现 Gemini 流式 | `apps/backend/internal/agent/llm_provider.go` | 同上 |
| 3.1.3 | 添加流式测试 | `apps/backend/internal/agent/llm_provider_test.go` | Mock 服务器 + 集成测试 |

#### 当前代码 (llm_provider.go)

```go
// line 332 — Claude
return fmt.Errorf("Claude streaming is not yet implemented; use non-streaming Chat instead")

// line 452 — Gemini
return fmt.Errorf("Gemini streaming is not yet implemented; use non-streaming Chat instead")
```

---

### 5.2 Step 3.2 — 实现 Agent 对话流式响应

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 3.2.1 | 实现 Agent SSE 流 | `apps/backend/internal/api/handlers/agent.go` | `POST /api/agent/chat` 改为 SSE 流式响应 |
| 3.2.2 | Agent 动作流式事件 | `apps/backend/internal/agent/agent.go` | 规划/执行/完成 各阶段推送事件 |
| 3.2.3 | 前端流式 Chat 对接 | `apps/desktop/src/pages/AIChat/` | MessageList 支持流式增量渲染 |

---

### 5.3 Step 3.3 — Agent 上下文感知

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 3.3.1 | 项目上下文注入 | `apps/backend/internal/agent/context.go` | 当前项目的 Workflow / 日志 / 错误信息 |
| 3.3.2 | Agent 动作系统完善 | `apps/backend/internal/agent/` | `create_workflow`, `modify_node`, `run_workflow`, `analyze_error` |
| 3.3.3 | 记忆系统增强 | `apps/backend/internal/agent/memory.go` | 语义检索（非简单 LIKE 匹配） |
| 3.3.4 | Workflow 自动生成 | `apps/backend/internal/service/agent_service.go` | 从自然语言 → Workflow JSON |

---

### 5.4 Phase 3 验收标准

```
□ Agent Chat 支持流式 SSE 响应（OpenAI + Claude + Gemini）
□ 前端 AI Chat 页面显示流式 Token
□ Agent 可读取当前项目上下文并生成工作流
□ /api/agent/generate-workflow 端点返回有效 Workflow JSON
```

---

## 6. Phase 4 — 数据管道与实时推送 (P2)

> **目标**：完善的实时数据流、任务日志和可视化监控  
> **预估**：1-2 周

### 6.1 Step 4.1 — 日志系统完善

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 4.1.1 | 诊断引擎持久化 | `apps/backend/internal/diagnostic/diagnostic.go` | 移除 TODO，实现 `GetHistory()` |
| 4.1.2 | 日志实时搜索 | `apps/backend/internal/logcenter/logcenter.go` | 全文搜索 / 过滤 |
| 4.1.3 | 日志 WebSocket 推送 | `apps/backend/internal/api/ws/hub.go` | 将 LogCenter 事件推送到前端 |
| 4.1.4 | 前端日志对接 | `apps/desktop/src/components/logs/` | LogViewer 接收实时日志流 |

---

### 6.2 Step 4.2 — 任务监控面板

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 4.2.1 | 任务执行时间线 | `apps/backend/internal/task/manager.go` | 每个节点的 start/end/duration 追踪 |
| 4.2.2 | 前端 TrainingMonitor 对接 | `apps/desktop/src/components/logs/TrainingMonitor.vue` | 展示训练进度 / loss 曲线 |
| 4.2.3 | 错误自动分析 | `apps/backend/internal/diagnostic/diagnostic.go` | 任务失败后自动执行 Analyze() |

---

### 6.3 Phase 4 验收标准

```
□ 任务运行状态实时推送到前端
□ 日志支持全文搜索和过滤
□ 诊断分析结果持久化并可查询
□ 前端 TrainingMonitor 显示真实训练进度
```

---

## 7. Phase 5 — 基础设施与生产化 (P2)

> **目标**：开发者体验优化、CI/CD、Docker 化  
> **预估**：1-2 周

### 7.1 Step 5.1 — 缺失基础设施文件

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 5.1.1 | 创建根目录 Makefile | `Makefile` (新建) | 统一构建 / 测试 / 开发命令 |
| 5.1.2 | 创建 .env.example | `.env.example` (新建) | 配置模板 |
| 5.1.3 | 创建 .gitignore | `.gitignore` (若缺失) | 忽略 build/ / .env / *.db |

#### Makefile 参考内容

```makefile
.PHONY: all dev build test clean

# Development
dev:
	cd apps/backend && go run ./cmd/

dev-frontend:
	cd apps/desktop && npm run dev

dev-engine:
	cd Engine && python server.py --port 8082

# Build
build-backend:
	cd apps/backend && go build -o ../../build/bin/ ./cmd/

build-frontend:
	cd apps/desktop && npm run build

# Test
test-backend:
	cd apps/backend && go test ./... -v -count=1

test-engine:
	cd Engine && python -m pytest

# Docker
docker-build:
	docker-compose build

docker-up:
	docker-compose up
```

---

### 7.2 Step 5.2 — Docker 开发环境

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 5.2.1 | 创建 docker-compose.yml | `docker-compose.yml` (新建) | backend + frontend + engine |
| 5.2.2 | 创建 Backend Dockerfile | `apps/backend/Dockerfile` (新建) | 多阶段构建 |
| 5.2.3 | 创建 Engine Dockerfile | `Engine/Dockerfile` (新建) | Python 依赖安装 |
| 5.2.4 | 创建 Frontend Dockerfile (服务模式) | `apps/desktop/Dockerfile` (新建) | Nginx 静态文件 |

---

### 7.3 Step 5.3 — CI/CD 配置

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 5.3.1 | GitHub Actions 后端 CI | `.github/workflows/backend-ci.yml` (新建) | lint → test → build |
| 5.3.2 | GitHub Actions 前端 CI | `.github/workflows/frontend-ci.yml` (新建) | typecheck → lint → build |
| 5.3.3 | GitHub Actions Engine CI | `.github/workflows/engine-ci.yml` (新建) | pytest → lint |
| 5.3.4 | GitHub Actions 集成测试 | `.github/workflows/integration.yml` (新建) | E2E 测试 |

---

### 7.4 Step 5.4 — 代码生成器注册统一

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| 5.4.1 | 在 main.go 注册全部 8 个生成器 | `apps/backend/cmd/main.go` | 移除 TODO，注册 MATLAB/ROS2/Docker 等 |

```go
// cmd/main.go — 当前只有 Python
compilerEngine.RegisterGenerator(compilerPython.NewGenerator())
// TODO: Register more generators

// 改为全部注册
compilerEngine.RegisterGenerator(compilerPython.NewGenerator())
compilerEngine.RegisterGenerator(compilerMATLAB.NewGenerator())
compilerEngine.RegisterGenerator(compilerROS2.NewGenerator())
compilerEngine.RegisterGenerator(compilerDocker.NewGenerator())
compilerEngine.RegisterGenerator(compilerSTM32.NewGenerator())
compilerEngine.RegisterGenerator(compilerCPP.NewGenerator())
compilerEngine.RegisterGenerator(compilerUnity.NewGenerator())
compilerEngine.RegisterGenerator(compilerJava.NewGenerator())
```

---

### 7.5 Phase 5 验收标准

```
□ make dev 能一键启动所有服务
□ docker-compose up 启动完整开发环境
□ CI 管道路由通过（lint + test + build）
□ 8 个代码生成器全部注册
□ .env.example 提供完整配置模板
```

---

## 8. Phase 6 — E2E 联调与测试 (P3)

> **目标**：完整端到端流程通过测试，bug 修复，文档对齐  
> **预估**：持续进行

### 8.1 Step 6.1 — 关键 E2E 流程

| # | 场景 | 涉及组件 | 说明 |
|---|------|----------|------|
| 6.1.1 | 用户注册 → 登录 → 创建项目 | Frontend → Auth → Database | 认证 + 项目 CRUD |
| 6.1.2 | 创建 Workflow → 保存 → 加载 | Frontend → API → Filesystem | Workflow JSON 持久化 |
| 6.1.3 | 拖拽节点 → 连接 → 配置 → 运行 | Frontend → Workflow Engine → Python Engine | 完整执行链路 |
| 6.1.4 | 运行中查看实时日志和进度 | WebSocket → Frontend LogViewer | 实时推送 |
| 6.1.5 | 任务完成/失败 → 查看结果 | Engine → Task System → WebSocket | 终态展示 |
| 6.1.6 | 安装插件 → 在工作流中使用 | Plugin Manager → Workflow Engine | 插件生命周期 |

### 8.2 Step 6.2 — 测试覆盖

| # | 模块 | 当前测试 | 目标 | 说明 |
|---|------|----------|------|------|
| 6.2.1 | Workflow DAG | ⚠️ 部分 | +20 测试 | 拓扑排序 / 环检测 / 端口校验 |
| 6.2.2 | Engine Bridge | ❌ 无 | +10 测试 | HTTP 客户端 / Mock Server |
| 6.2.3 | Plugin Manager | 已有基础 | +10 测试 | Execute / Install / 错误处理 |
| 6.2.4 | Agent System | ⚠️ 部分 | +15 测试 | Planner / Executor / Memory |
| 6.2.5 | LLM Provider | ❌ 无 | +8 测试 | 各 Provider 的 Chat + Stream |
| 6.2.6 | MCP Runtime | ❌ 无 | +10 测试 | Connect / CallTool / ListTools |
| 6.2.7 | WebSocket Hub | ❌ 无 | +8 测试 | 连接管理 / 广播 / 房间 |
| 6.2.8 | Python Engine | ⚠️ 部分 | +20 测试 | YOLO / Dataset / Trainer |

### 8.3 Step 6.3 — 文档对齐

| # | 任务 | 说明 |
|---|------|------|
| 6.3.1 | API 文档与真实路由对齐 | 确保 `Docs/api/` 与实际 `router.go` 一致 |
| 6.3.2 | 更新 README | 真实构建步骤、依赖版本、配置说明 |
| 6.3.3 | 补充缺失的 v2 架构文档引用 | 确保 ADR 与实际实现一致 |

---

## 9. 时间线汇总

```
周次    Phase                           里程碑
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1-2     Phase 1: 引擎打通                 工作流可真实调用 Python
                                             WebSocket 实时推送
                                             插件可执行
3-4     Phase 2: 插件系统                 插件安装/卸载/管理完整流程
                                             多语言执行器
5-6     Phase 3: Agent 完善               Claude/Gemini 流式可用
                                             Agent 自动生成 Workflow
7-8     Phase 4: 数据管道                  诊断持久化
                                             训练监控面板
                                             日志实时搜索
9-10    Phase 5: 基础设施                  Docker 开发环境
                                             CI/CD 管道
                                             Makefile
11+     Phase 6: E2E 测试                  测试覆盖 ≥ 60%
                                             所有 E2E 场景通过
```

---

## 10. 详细文件变更清单

### 新建文件

```
Phase 1:
  apps/backend/internal/engine/                 # Go↔Python 通信
  apps/backend/internal/engine/client.go         # EngineClient 接口
  apps/backend/internal/engine/http_client.go    # HTTP 客户端实现
  apps/backend/internal/engine/config.go         # Engine 配置
  apps/backend/internal/api/ws/                  # WebSocket 包
  apps/backend/internal/api/ws/hub.go            # 连接管理
  apps/backend/internal/api/ws/client.go         # 连接读写
  apps/backend/internal/api/ws/messages.go       # 消息格式
  apps/backend/internal/workflow/executors/      # 节点执行器
  apps/backend/internal/workflow/executors/yolo_train.go
  apps/backend/internal/workflow/executors/yolo_predict.go
  apps/backend/internal/workflow/executors/condition.go
  apps/backend/internal/workflow/executors/loop.go
  apps/backend/internal/workflow/executors/switch.go
  apps/backend/internal/workflow/executors/retry.go
  apps/backend/internal/plugin/installer.go      # 插件安装

Phase 2:
  apps/backend/internal/plugin/executors/        # 插件执行器
  apps/backend/internal/plugin/executors/python_executor.go
  apps/backend/internal/plugin/executors/process_executor.go

Phase 5:
  Makefile                                       # 根目录构建文件
  .env.example                                   # 环境变量模板
  .gitignore                                     # Git 忽略规则
  docker-compose.yml                             # 开发环境
  apps/backend/Dockerfile                        # 后端镜像
  Engine/Dockerfile                              # Python 引擎镜像
  .github/workflows/                             # CI/CD

Phase 6:
  tests/e2e/                                     # E2E 测试
```

### 修改文件

```
Phase 1:
  apps/backend/internal/workflow/builtin_nodes.go     # 节点 Factory 指向真实执行器
  apps/backend/cmd/main.go                            # 注册 Engine Bridge, WebSocket
  apps/backend/internal/api/router.go                 # WebSocket 路由更新
  apps/backend/internal/service/service.go            # 添加 EngineService
  apps/backend/internal/plugin/manager.go             # 添加 Execute(), RegisterExecutor()

Phase 3:
  apps/backend/internal/agent/llm_provider.go         # Claude/Gemini 流式
  apps/backend/internal/agent/context.go              # 项目上下文
  apps/backend/internal/agent/memory.go               # 语义检索
  apps/backend/internal/api/handlers/agent.go          # SSE 流式

Phase 4:
  apps/backend/internal/diagnostic/diagnostic.go       # 移除 TODO，实现持久化
  apps/backend/internal/database/models/               # 添加诊断记录模型

Phase 5:
  apps/backend/cmd/main.go                             # 注册全部 8 个生成器
```

---

## 11. 参考文档索引

| 文档 | 路径 | 适用 Phase |
|------|------|-----------|
| 系统架构设计 | `Docs/architecture/system-design.md` | Phase 1-6 |
| 后端架构 | `Docs/architecture/backend.md` | Phase 1-6 |
| v2 架构总览 | `Docs/architecture/v2/OVERVIEW.md` | Phase 1-6 |
| Workflow DSL | `Docs/architecture/v2/WORKFLOW_DSL.md` | Phase 1 |
| 编译器架构 | `Docs/architecture/v2/COMPILER.md` | Phase 5 |
| 运行时设计 | `Docs/architecture/v2/RUNTIME.md` | Phase 1 |
| 插件系统 | `Docs/architecture/v2/PLUGIN_SYSTEM.md` | Phase 2 |
| Agent 设计 | `Docs/architecture/v2/AGENT.md` | Phase 3 |
| Engine 架构 | `Docs/architecture/engine.md` | Phase 1 |
| API 接口 | `Docs/api/README.md` | Phase 1-6 |
| 集成检查 | `Docs/architecture/integration-check.md` | Phase 6 |
| ADR 文档 | `Docs/ADR/ADR-*.md` | 各 Phase 参考 |

---

## 12. 附录

### 12.1 关键代码位置快速索引

| 模块 | 路径 |
|------|------|
| 后端入口 | `apps/backend/cmd/main.go` |
| 路由定义 | `apps/backend/internal/api/router.go` |
| 服务容器 | `apps/backend/internal/service/service.go` |
| Workflow 类型 | `apps/backend/internal/workflow/types.go` |
| DAG 操作 | `apps/backend/internal/workflow/dag.go` |
| 节点定义 | `apps/backend/internal/workflow/builtin_nodes.go` |
| 执行引擎 | `apps/backend/internal/workflow/types.go` (Engine struct) |
| 任务管理器 | `apps/backend/internal/task/manager.go` |
| 插件管理器 | `apps/backend/internal/plugin/manager.go` |
| Agent 核心 | `apps/backend/internal/agent/agent.go` |
| LLM Provider | `apps/backend/internal/agent/llm_provider.go` |
| MCP 管理器 | `apps/backend/internal/mcp/manager.go` |
| 环境管理器 | `apps/backend/internal/environment/checker.go` |
| 诊断引擎 | `apps/backend/internal/diagnostic/diagnostic.go` |
| 事件总线 | `apps/backend/internal/eventbus/eventbus.go` |
| Python Server | `Engine/server.py` |
| Python Runner | `Engine/runner.py` |
| YOLO 训练 | `Engine/vision/yolo/train.py` |
| 前端入口 | `apps/desktop/src/App.vue` |
| 前端路由 | `apps/desktop/src/router/index.ts` |
| Tauri 配置 | `apps/desktop/src-tauri/tauri.conf.json` |

### 12.2 依赖关系

```
Phase 1 ──→ Phase 2 ──→ Phase 3 ──→ Phase 4 ──→ Phase 5 ──→ Phase 6
  │                                                    │
  └────────────────────────────────────────────────────┘
                       ↑ 可并行

Phase 1 内部:
  Step 1.1 Engine Bridge ──→ Step 1.2 节点执行器
  Step 1.3 WebSocket ────── 独立，可并行
  Step 1.4 插件执行 ────── 独立，可并行

Phase 5 内部:
  Makefile ── 独立，可随时进行
  Docker ── 依赖于 Phase 1 完成
  CI/CD ── 可早期开始
```

### 12.3 技术债务清单

| 债务 | 优先级 | 说明 |
|------|--------|------|
| `noOpExecutor` 占位节点 | P0 | 所有 40+ 节点需要真实实现 |
| `// TODO: Implement history storage` | P1 | 诊断引擎缺失持久化 |
| `Claude/Gemini streaming not implemented` | P1 | 两个主流 LLM 流式缺失 |
| `// TODO: Register more generators` | P2 | 8 个适配器仅注册 1 个 |
| 206 个测试覆盖率低 | P2 | 核心模块缺少单元测试 |
| MCP 仅 Mock 模式 | P2 | 无真实远程 MCP 服务器连接 |
| 根目录 Makefile 缺失 | P2 | README 引用但不存在 |
| .env 文件缺失 | P2 | 配置加载器引用但未提供模板 |
| 无 CI/CD | P2 | 无自动化质量门禁 |
| 文档与实现可能不一致 | P3 | 需对齐架构文档与实际代码 |
