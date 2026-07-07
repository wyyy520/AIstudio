# AIStudio 系统架构设计文档

> 版本：V1.0
> 日期：2026-07-07
> 状态：设计阶段

---

## 目录

1. [后端整体架构](#1-后端整体架构)
2. [Workflow Engine 设计](#2-workflow-engine-设计)
3. [Plugin 系统设计](#3-plugin-系统设计)
4. [Task 系统设计](#4-task-系统设计)
5. [API 接口规范](#5-api-接口规范)
6. [Go 与 Python 通信方案](#6-go-与-python-通信方案)
7. [数据库设计](#7-数据库设计)
8. [开发路线图](#8-开发路线图)

---

## 1. 后端整体架构

### 1.1 系统分层

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend (Vue3 + Tauri)                   │
│   Dashboard │ Workflow Editor │ AI Chat │ Plugin Store │ Logs   │
└──────────────────────────┬──────────────────────────────────────┘
                           │ Tauri IPC / HTTP REST / WebSocket
┌──────────────────────────┴──────────────────────────────────────┐
│                         Backend (Go)                              │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    API Gateway (Gin)                       │   │
│  │   路由 │ 中间件 │ 认证 │ 限流 │ 日志 │ 错误恢复            │   │
│  └──────────────────────┬───────────────────────────────────┘   │
│                         │                                        │
│  ┌──────────┐ ┌────────┴──────┐ ┌──────────┐ ┌──────────┐     │
│  │ Project  │ │  Workflow      │ │  Agent   │ │  Plugin  │     │
│  │ Service  │ │  Engine        │ │  Manager │ │  Manager │     │
│  └────┬─────┘ └──────┬────────┘ └────┬─────┘ └────┬─────┘     │
│       │              │                │             │            │
│  ┌────┴──────────────┴────────────────┴─────────────┘          │
│  │                    Task Scheduler                            │   │
│  │         队列 │ 调度 │ 重试 │ 超时 │ 状态追踪                  │   │
│  └──────────────────────┬───────────────────────────────────┘   │
│                         │                                        │
│  ┌──────────┐ ┌────────┴──────┐ ┌──────────┐ ┌──────────┐     │
│  │ Logger   │ │  Environment  │ │   MCP    │ │  Config  │     │
│  │ Service  │ │  Manager      │ │  Manager │ │  Manager │     │
│  └──────────┘ └───────────────┘ └──────────┘ └──────────┘     │
│                         │                                        │
│  ┌──────────────────────┴───────────────────────────────────┐   │
│  │                    Engine Bridge                            │   │
│  │            gRPC Client │ 进程管理 │ 健康检查                  │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────────────────────┬──────────────────────────────────────┘
                           │ gRPC
┌──────────────────────────┴──────────────────────────────────────┐
│                      Engine (Python)                              │
│   Vision Handler │ NLP Handler │ Speech │ TimeSeries │ SDK       │
│   Model Manager │ Device Manager │ Runtime Cache │ Logger        │
└─────────────────────────────────────────────────────────────────┘
                           │
┌──────────────────────────┴──────────────────────────────────────┐
│                      Plugins                                      │
│   Vision (YOLO/SAM/OCR) │ NLP (LLM/Transformer) │ Logic │ System│
└─────────────────────────────────────────────────────────────────┘
                           │
┌──────────────────────────┴──────────────────────────────────────┐
│                      Storage / Runtime                            │
│   SQLite │ 文件系统 │ 模型缓存 │ 数据集 │ 日志 │ 临时工作空间      │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 目录结构

```
Backend/
├── cmd/
│   └── main.go                    # 程序入口，初始化并启动服务
├── internal/
│   ├── api/                       # HTTP 路由与 Handler
│   │   ├── router.go              # 路由注册
│   │   ├── middleware.go          # 中间件（认证/日志/恢复/跨域）
│   │   ├── handler_project.go     # 项目管理 Handler
│   │   ├── handler_workflow.go    # 工作流 Handler
│   │   ├── handler_plugin.go      # 插件管理 Handler
│   │   ├── handler_task.go        # 任务 Handler
│   │   ├── handler_agent.go       # Agent 对话 Handler
│   │   ├── handler_logs.go        # 日志 Handler
│   │   ├── handler_environment.go # 环境管理 Handler
│   │   └── handler_mcp.go         # MCP Handler
│   ├── workflow/                  # 工作流引擎核心
│   │   ├── engine.go              # 引擎主入口
│   │   ├── dag.go                 # DAG 图数据结构
│   │   ├── executor.go            # 节点执行器
│   │   ├── router.go              # 数据路由器（端口映射）
│   │   ├── validator.go           # 图校验（环检测/端口匹配）
│   │   └── types.go               # 工作流相关类型定义
│   ├── task/                      # 任务调度器
│   │   ├── scheduler.go           # 任务调度器
│   │   ├── queue.go               # 优先级队列
│   │   ├── worker.go              # Worker 池
│   │   ├── state.go               # 状态机
│   │   └── types.go               # 任务相关类型定义
│   ├── plugin/                    # 插件加载与管理
│   │   ├── manager.go             # 插件管理器
│   │   ├── loader.go              # 插件加载器（Go/Python/进程）
│   │   ├── registry.go            # 插件注册表
│   │   ├── installer.go           # 插件安装器
│   │   └── types.go               # 插件相关类型定义
│   ├── project/                   # 项目管理
│   │   ├── service.go             # 项目服务
│   │   └── types.go               # 项目相关类型定义
│   ├── environment/               # 运行环境管理
│   │   ├── manager.go             # 环境管理器
│   │   ├── python.go              # Python 环境检测与配置
│   │   ├── cuda.go                # CUDA/GPU 状态检测
│   │   └── types.go               # 环境相关类型定义
│   ├── logger/                    # 日志系统
│   │   ├── service.go             # 日志服务
│   │   ├── writer.go              # 日志写入器（支持实时推送）
│   │   ├── analyzer.go            # AI 日志分析器
│   │   └── types.go               # 日志相关类型定义
│   ├── agent/                     # Agent 管理
│   │   ├── manager.go             # Agent 管理器
│   │   ├── session.go             # 对话会话管理
│   │   ├── executor.go            # Agent 动作执行器
│   │   └── types.go               # Agent 相关类型定义
│   ├── mcp/                       # MCP 协议实现
│   │   ├── manager.go             # MCP 管理器
│   │   ├── client.go              # MCP 客户端
│   │   ├── server.go              # MCP 服务端
│   │   └── types.go               # MCP 相关类型定义
│   ├── database/                  # 数据库连接与迁移
│   │   ├── db.go                  # 数据库连接初始化
│   │   ├── migrate.go             # 自动迁移
│   │   └── models.go              # GORM 模型定义
│   ├── config/                    # 配置管理
│   │   ├── config.go              # 配置加载与解析
│   │   └── defaults.go            # 默认配置
│   └── common/                    # 公共工具
│       ├── response.go            # 统一响应封装
│       ├── errors.go              # 错误码定义
│       ├── utils.go               # 工具函数
│       └── id.go                  # ID 生成器
├── pkg/                           # 可外部引用的公共包
│   ├── plugin/                    # 插件接口定义（供 Go 插件实现）
│   │   └── plugin.go
│   └── engine/                    # Engine gRPC 客户端
│       └── client.go
└── go.mod
```

### 1.3 模块职责

| 模块 | 职责 | 对外接口 |
|------|------|----------|
| `api/` | HTTP 路由、请求校验、响应封装、中间件 | REST API + WebSocket |
| `workflow/` | 工作流解析、DAG 构建、拓扑排序、节点调度执行 | `Engine.Run()` |
| `task/` | 任务生命周期管理、队列调度、Worker 池、状态追踪 | `Scheduler.Submit()` |
| `plugin/` | 插件发现/加载/注册/卸载/安装、插件执行调度 | `Manager.Execute()` |
| `project/` | 项目 CRUD、项目工作空间管理、项目统计 | `Service.Create/Get/Update/Delete()` |
| `environment/` | Python 环境检测、CUDA 状态、依赖管理 | `Manager.Check/Detect()` |
| `logger/` | 日志收集、存储、实时推送、AI 错误分析 | `Service.Write/Stream()` |
| `agent/` | AI 对话会话管理、Agent 动作执行、LLM 调用 | `Manager.Chat()` |
| `mcp/` | MCP 协议客户端/服务端、工具注册与调用 | `Manager.CallTool()` |
| `database/` | 数据库连接、GORM 模型、自动迁移 | `db.DB` |
| `config/` | 配置文件加载、环境变量覆盖、默认值 | `config.C` |

### 1.4 依赖关系

```
api ──→ workflow ──→ task ──→ plugin ──→ pkg/engine (gRPC)
  │         │          │         │
  │         │          └──→ logger
  │         │
  │         └──→ environment
  │
  ├──→ project ──→ database
  ├──→ agent ──→ mcp
  ├──→ logger ──→ database
  └──→ config (全局)
```

核心原则：
- **单向依赖**：上层依赖下层，下层不依赖上层
- **接口隔离**：模块间通过接口通信，不直接依赖实现
- **数据库统一**：所有数据访问通过 `database/` 模块

---

## 2. Workflow Engine 设计

### 2.1 Workflow JSON 规范

```json
{
  "id": "wf_001",
  "name": "车辆检测工作流",
  "version": 3,
  "project_id": "proj_001",
  "graph": {
    "nodes": [
      {
        "id": "n1",
        "type": "input",
        "plugin": "image_input",
        "position": { "x": 100, "y": 200 },
        "config": {
          "source": "file"
        },
        "ports": {
          "inputs": [],
          "outputs": [
            { "name": "image", "type": "image", "description": "输出图像路径" }
          ]
        }
      },
      {
        "id": "n2",
        "type": "vision",
        "plugin": "yolo",
        "position": { "x": 400, "y": 200 },
        "config": {
          "model": "yolov8n.pt",
          "confidence": 0.5,
          "device": "auto"
        },
        "ports": {
          "inputs": [
            { "name": "image", "type": "image", "required": true }
          ],
          "outputs": [
            { "name": "detections", "type": "json" },
            { "name": "annotated_image", "type": "image" }
          ]
        }
      },
      {
        "id": "n3",
        "type": "logic",
        "plugin": "if",
        "position": { "x": 700, "y": 200 },
        "config": {
          "condition": "len(input.get('boxes', [])) > 0"
        },
        "ports": {
          "inputs": [
            { "name": "input", "type": "json", "required": true }
          ],
          "outputs": [
            { "name": "true", "type": "json" },
            { "name": "false", "type": "json" }
          ]
        }
      }
    ],
    "edges": [
      {
        "id": "e1",
        "from": "n1",
        "to": "n2",
        "from_port": "image",
        "to_port": "image"
      },
      {
        "id": "e2",
        "from": "n2",
        "to": "n3",
        "from_port": "detections",
        "to_port": "input"
      }
    ]
  }
}
```

### 2.2 Node 结构定义

```
Node
├── id: string              # 唯一标识，如 "n1"
├── type: NodeType          # input / vision / nlp / logic / system / output / note
├── plugin: string          # 对应插件名，如 "yolo"
├── position: {x, y}        # 前端画布位置
├── config: Map             # 节点配置参数（来自属性面板）
├── ports
│   ├── inputs: []Port      # 输入端口列表
│   │   ├── name            # 端口名
│   │   ├── type            # 数据类型 (image/text/number/json/file/tensor/stream)
│   │   ├── required        # 是否必填
│   │   └── default         # 默认值（可选端口）
│   └── outputs: []Port     # 输出端口列表
│       ├── name
│       ├── type
│       └── description
└── style: Map              # 样式覆盖（可选）
```

### 2.3 Edge 结构定义

```
Edge
├── id: string              # 唯一标识，如 "e1"
├── from: string            # 源节点 ID
├── to: string              # 目标节点 ID
├── from_port: string       # 源端口名
└── to_port: string         # 目标端口名
```

### 2.4 DAG 解析流程

```
前端 JSON
    │
    ▼
1. 反序列化 → Workflow 对象
    │
    ▼
2. 构建邻接表
    │  遍历所有 Edge，构建:
    │  - adjacency[from] = [{to, fromPort, toPort}]
    │  - reverseAdjacency[to] = [{from, fromPort, toPort}]
    │  - inDegree[to]++
    │
    ▼
3. 校验
    │  a. 环检测（DFS 三色标记法）
    │  b. 端口类型匹配（from_port.type == to_port.type）
    │  c. 必填端口检查（所有 required input 都有连入边）
    │  d. 孤立节点检测（无连入也无连出的非输入节点）
    │
    ▼
4. 拓扑排序（Kahn 算法）
    │  - 入度为 0 的节点入队
    │  - 逐个出队，将后继入度减 1
    │  - 入度为 0 的后继入队
    │  - 结果为分层执行顺序
    │
    ▼
5. 分层并行
    │  同一层（无相互依赖）的节点可并行执行
    │  [[n1], [n2, n3], [n4]]  ← 三层
    │
    ▼
6. 执行
```

### 2.5 执行引擎核心流程

```
Engine.Run(workflow, inputs)
│
├── 1. 创建执行上下文 WorkflowContext
│   ├── task_id
│   ├── project_id
│   ├── work_dir: /runtime/workspace/{task_id}/
│   ├── node_outputs: Map[nodeID]Map[portName]data  ← 节点输出缓存
│   └── variables: Map  ← 全局变量
│
├── 2. 注入初始输入
│   └── 将用户提供的 inputs 写入对应输入节点的输出缓存
│
├── 3. 按拓扑层遍历
│   └── for each level:
│       ├── 并行执行同层节点
│       │   └── executeNode(node)
│       │       ├── 从 node_outputs 读取上游数据
│       │       ├── 调用 PluginManager.Execute(plugin, input, config)
│       │       ├── 将输出写入 node_outputs[node.id]
│       │       ├── 推送 node_status 事件 (running → success/failed)
│       │       └── 推送 node_log 事件
│       │
│       └── 等待本层全部完成
│
├── 4. 数据路由
│   └── 节点执行完成后，根据 Edge 定义
│       将 output[from_port] 映射到下游 node 的 input[to_port]
│
├── 5. 错误处理
│   ├── 默认策略: fail_fast（任一节点失败，停止整个工作流）
│   ├── 节点级配置: continue_on_error（跳过失败节点）
│   └── 重试: 通过 Retry 节点包装
│
└── 6. 完成
    ├── 推送 workflow_done 事件
    ├── 收集最终输出节点数据
    └── 更新 Task 状态
```

### 2.6 条件分支与循环

**If 节点**：根据条件表达式，只激活 `true` 或 `false` 输出端口对应的下游边。

```
n2 (If) 执行后:
  - condition = true  → 只将数据路由到 from_port="true" 的边
  - condition = false → 只将数据路由到 from_port="false" 的边
  - 另一分支的下游节点标记为 skipped
```

**Loop 节点**：内部维护迭代计数器，每次迭代将当前元素通过 `output` 端口传出，迭代完成后通过 `completed` 端口传出汇总结果。

```
Loop 执行流程:
  for item in input_list:
    → route item to output port
    → execute downstream nodes
    → collect result
  → route collected_results to completed port
```

### 2.7 断点续跑

任务失败后保留已成功节点的输出缓存。重试时：

```
retry_task(task_id, skip_nodes=["n1", "n3"])
  → 从 node_outputs 恢复 n1、n3 的结果
  → 仅重新执行 n2（失败节点）及其下游
```

---

## 3. Plugin 系统设计

### 3.1 插件标准结构

```
Plugins/
└── Vision/
    └── YOLO/
        ├── plugin.json          # 插件清单（必需）
        ├── main.py              # 入口文件
        ├── requirements.txt     # Python 依赖
        ├── install.py           # 安装脚本（可选）
        ├── README.md            # 插件说明
        └── assets/
            └── icon.png         # 图标
```

### 3.2 plugin.json 规范

```json
{
  "name": "yolo-detector",
  "version": "1.0.0",
  "author": "AI Studio",
  "description": "YOLO 目标检测插件",
  "type": "vision",
  "icon": "assets/icon.png",
  "language": "python",
  "entry": "main.py",
  "ports": {
    "inputs": [
      { "name": "image", "type": "image", "required": true, "description": "输入图像" },
      { "name": "confidence", "type": "number", "required": false, "default": 0.5 }
    ],
    "outputs": [
      { "name": "detections", "type": "json", "description": "检测结果" },
      { "name": "annotated_image", "type": "image", "description": "标注图像" }
    ]
  },
  "config_schema": {
    "model": { "type": "string", "default": "yolov8n.pt", "description": "模型权重文件" },
    "device": { "type": "select", "default": "auto", "options": ["auto", "cpu", "cuda"] }
  },
  "dependencies": {
    "python": ">=3.10",
    "cuda": ">=11.8",
    "packages": ["ultralytics>=8.0.0", "opencv-python>=4.8.0"]
  },
  "install_steps": [
    { "action": "pip_install", "args": ["-r", "requirements.txt"] },
    { "action": "download", "url": "https://github.com/ultralytics/assets/releases/download/v8.1.0/yolov8n.pt", "dest": "models/yolov8n.pt" }
  ]
}
```

### 3.3 插件安装流程

```
用户点击"安装"
    │
    ▼
1. 下载/复制插件目录到 Plugins/{type}/{name}/
    │
    ▼
2. 解析 plugin.json，校验完整性
    │
    ▼
3. 检查依赖
    ├── Python 版本检查
    ├── CUDA 版本检查（如需要）
    └── 系统依赖检查
    │
    ▼
4. 执行安装步骤 (install_steps)
    ├── pip install -r requirements.txt
    ├── 下载模型权重
    ├── git clone（如需要）
    └── 执行 install.py（如存在）
    │
    ▼
5. 注册插件
    ├── 写入 plugin_registry 表
    ├── 生成 Workflow 节点模板
    └── 推送安装完成事件
    │
    ▼
6. 前端刷新节点面板，新节点可用
```

### 3.4 插件执行调度

```
Workflow Engine 请求执行节点
    │
    ▼
PluginManager.Execute(pluginName, input, config)
    │
    ├── 判断插件语言
    │   │
    │   ├── Python 插件
    │   │   └── 通过 Engine Bridge (gRPC) 调用 Python Engine
    │   │       ├── Engine 路由到对应 Handler
    │   │       ├── Handler 加载/缓存插件实例
    │   │       ├── 执行 plugin.execute(inputs, config)
    │   │       └── 返回结果
    │   │
    │   ├── Go 插件
    │   │   └── 直接调用 Go Plugin 接口（内存内）
    │   │       ├── 从 registry 获取插件实例
    │   │       ├── 调用 plugin.Execute(ctx, input)
    │   │       └── 返回结果
    │   │
    │   └── 进程隔离插件
    │       └── 启动独立进程，通过 stdin/stdout JSON 通信
    │           ├── 启动子进程
    │           ├── 写入输入 JSON
    │           ├── 读取输出 JSON
    │           └── 返回结果
    │
    ▼
返回 PluginOutput
    ├── data: Map[portName]value  ← 输出端口数据
    ├── status: success / error
    ├── error: string
    └── metrics: { duration_ms, memory_mb, device }
```

### 3.5 插件类型与注册节点映射

| 插件类型 | 注册为 Workflow 节点 | 节点颜色 | 示例 |
|----------|----------------------|----------|------|
| vision | VisionNode | 绿色 | YOLO, SAM, OCR, RT-DETR |
| nlp | NLPNode | 粉色 | LLM, Transformer, NER |
| logic | LogicNode | 黄色 | If, Switch, Loop, Retry |
| system | SystemNode | 紫色 | Python, Git, Terminal, Download |
| simulation | SimulationNode | 青色 | SUMO, MATLAB |
| mcp | MCPNode | 蓝色 | 外部 MCP 工具 |

---

## 4. Task 系统设计

### 4.1 任务生命周期

```
                    ┌──────────┐
                    │  Created  │
                    └─────┬────┘
                          │ 入队
                    ┌─────▼────┐
              ┌─────│  Queued   │─────┐
              │     └─────┬────┘     │
              │ 取消       │ 调度     │ 取消
              ▼           ▼          ▼
        ┌──────────┐ ┌──────────┐ ┌───────────┐
        │Cancelled │ │ Running  │ │ Cancelled  │
        └──────────┘ └─────┬────┘ └───────────┘
                     ┌─────┼──────┐
                     │     │      │
                  成功   失败   取消
                     │     │      │
                     ▼     ▼      ▼
              ┌────────┐ ┌────────┐ ┌───────────┐
              │Success │ │ Failed │ │ Cancelled  │
              └────────┘ └───┬────┘ └───────────┘
                              │
                              │ 重试
                              ▼
                        ┌──────────┐
                        │ Running  │ (新 Task，复用已完成节点)
                        └──────────┘
```

### 4.2 状态定义

| 状态 | 标识 | 含义 | 可转换到 |
|------|------|------|----------|
| Created | `created` | 任务已创建 | queued, cancelled |
| Queued | `queued` | 已加入调度队列 | running, cancelled |
| Running | `running` | 正在执行 | success, failed, cancelled |
| Success | `success` | 全部节点执行成功 | - (终态) |
| Failed | `failed` | 有节点执行失败 | running (重试) |
| Cancelled | `cancelled` | 用户手动取消 | - (终态) |

### 4.3 Task 数据结构

```
Task
├── id: string                    # "task_20260707_001"
├── workflow_id: string           # 关联的工作流
├── project_id: string            # 关联的项目
├── status: TaskStatus            # 当前状态
├── progress: float               # 0.0 ~ 1.0
├── inputs: Map[nodeID]Map        # 工作流输入数据
├── config: Map                   # 运行配置 (device, debug 等)
├── node_states: []NodeState      # 各节点执行状态
│   ├── node_id
│   ├── status: pending/running/success/failed/skipped
│   ├── started_at
│   ├── finished_at
│   ├── duration_ms
│   ├── output: Map[portName]data
│   └── error: string
├── result: Map                   # 最终输出（输出节点数据汇总）
├── error: string                 # 全局错误信息
├── created_at: timestamp
├── started_at: timestamp
├── finished_at: timestamp
└── duration_ms: int
```

### 4.4 调度器设计

```
TaskScheduler
├── queue: PriorityQueue          # 优先级队列（按创建时间 FIFO）
├── workers: WorkerPool           # Worker 池（默认 4 个并发）
├── running: Map[taskID]*Task     # 正在运行的任务
└── history: []TaskRecord         # 历史记录

调度策略:
  - 同一项目同一工作流，最多 1 个任务同时运行（避免 GPU 冲突）
  - 不同项目的任务可并行
  - 训练类任务独占 GPU 资源
  - 推理类任务可共享 GPU（通过 Engine 端模型缓存）
```

### 4.5 实时状态推送

```
Backend                              Frontend
  │                                     │
  │── WebSocket 连接 ──────────────────│
  │  ws://host/ws/task/{taskId}        │
  │                                     │
  │── {"type":"task_status",            │
  │    "status":"running"}              │
  │                                     │
  │── {"type":"node_status",            │
  │    "node_id":"n1",                  │
  │    "status":"running"}              │
  │                                     │
  │── {"type":"node_log",               │
  │    "node_id":"n1",                  │
  │    "level":"info",                  │
  │    "message":"Loading model..."}    │
  │                                     │
  │── {"type":"node_progress",          │
  │    "node_id":"n1",                  │
  │    "progress":0.45}                 │
  │                                     │
  │── {"type":"node_status",            │
  │    "node_id":"n1",                  │
  │    "status":"success",              │
  │    "output":{...}}                  │
  │                                     │
  │── {"type":"task_done",              │
  │    "status":"success",              │
  │    "result":{...}}                  │
```

---

## 5. API 接口规范

### 5.1 基础约定

| 项目 | 值 |
|------|-----|
| Base URL | `http://localhost:8080/api/v1` |
| 协议 | HTTP REST + WebSocket |
| 数据格式 | JSON |
| 认证 | Bearer Token（桌面版可跳过） |
| 编码 | UTF-8 |
| 时间格式 | ISO 8601 (RFC 3339) |

### 5.2 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 5.3 错误码体系

| 范围 | 模块 | 示例 |
|------|------|------|
| 0 | 成功 | 0 = 成功 |
| 1000-1999 | 通用错误 | 1001=参数错误, 1002=未认证, 1003=权限不足 |
| 2000-2999 | 工作流 | 2001=工作流不存在, 2002=校验失败, 2003=有环 |
| 3000-3999 | 插件 | 3001=插件不存在, 3002=执行失败, 3003=安装失败 |
| 4000-4999 | 任务 | 4001=任务不存在, 4002=任务超时, 4003=任务已取消 |
| 5000-5999 | 引擎 | 5001=连接失败, 5002=模型加载失败, 5003=CUDA错误 |
| 6000-6999 | 环境 | 6001=Python未找到, 6002=CUDA不匹配, 6003=依赖缺失 |

### 5.4 完整接口列表

#### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/login` | 用户登录 |
| POST | `/auth/logout` | 用户登出 |
| GET | `/auth/profile` | 获取当前用户 |

#### 项目管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/projects` | 项目列表 |
| POST | `/projects` | 创建项目 |
| GET | `/projects/:id` | 项目详情 |
| PUT | `/projects/:id` | 更新项目 |
| DELETE | `/projects/:id` | 删除项目 |

#### 工作流

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/workflows` | 工作流列表 |
| POST | `/workflows` | 创建工作流 |
| GET | `/workflows/:id` | 工作流详情（含图数据） |
| PUT | `/workflows/:id` | 更新工作流 |
| DELETE | `/workflows/:id` | 删除工作流 |
| POST | `/workflows/:id/run` | 运行工作流 |
| GET | `/workflows/:id/status` | 运行状态 |
| POST | `/workflows/:id/stop` | 停止运行 |
| GET | `/workflows/:id/nodes/:nodeId/output` | 节点输出 |

#### 插件管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/plugins` | 插件列表 |
| GET | `/plugins/:name` | 插件详情 |
| POST | `/plugins/install` | 安装插件 |
| DELETE | `/plugins/:name` | 卸载插件 |
| GET | `/plugins/:name/config-schema` | 配置 Schema |
| POST | `/plugins/:name/test` | 测试插件 |

#### 任务

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/tasks` | 任务列表 |
| GET | `/tasks/:id` | 任务详情 |
| GET | `/tasks/:id/logs` | 任务日志 |
| POST | `/tasks/:id/cancel` | 取消任务 |
| POST | `/tasks/:id/retry` | 重试任务 |

#### 日志

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/logs/tasks` | 日志任务列表 |
| GET | `/logs/:taskId` | 任务日志内容 |
| POST | `/logs/analyze` | AI 分析日志 |
| POST | `/logs/fix` | 执行修复 |

#### Agent

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/agent/chat` | 发送对话（SSE 流式响应） |
| GET | `/agent/sessions` | 对话列表 |
| POST | `/agent/sessions` | 创建对话 |
| GET | `/agent/sessions/:id` | 对话历史 |
| DELETE | `/agent/sessions/:id` | 删除对话 |

#### 环境

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/environment/status` | 环境状态（Python/CUDA/GPU） |
| POST | `/environment/check` | 重新检测环境 |
| GET | `/environment/dependencies` | 依赖列表 |
| POST | `/environment/install` | 安装依赖 |

#### MCP

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/mcp/tools` | 可用 MCP 工具列表 |
| POST | `/mcp/tools/:name/call` | 调用 MCP 工具 |

#### WebSocket

| 路径 | 说明 |
|------|------|
| `ws://host/ws/task/:taskId` | 任务实时状态推送 |
| `ws://host/ws/agent/:sessionId` | Agent 流式对话 |

---

## 6. Go 与 Python 通信方案

### 6.1 方案对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| `os/exec` | 零依赖、实现简单 | 每次启动新进程、无法复用模型、延迟高 | 一次性脚本执行 |
| HTTP | 简单通用、跨语言 | 无流式支持、JSON 序列化开销 | 简单请求-响应 |
| gRPC | 高性能、流式、强类型、代码生成 | 需维护 proto 文件、Python 端需运行服务 | 高频调用、流式推理 |

### 6.2 第一版方案：gRPC

**选择理由**：
1. AI 推理需要流式输出（LLM token 流、训练进度），gRPC 原生支持双向流
2. 模型加载后常驻内存，gRPC 服务端可缓存模型实例，避免重复加载
3. 强类型 proto 定义，前后端契约清晰
4. Go 和 Python 都有成熟的 gRPC 库

### 6.3 架构设计

```
Go Backend                              Python Engine
┌──────────────────┐                   ┌──────────────────┐
│  Engine Bridge    │                   │  gRPC Server      │
│  ┌──────────────┐ │    gRPC           │  ┌──────────────┐ │
│  │ gRPC Client  │─┼───────────────────┼─→│ EngineService │ │
│  └──────────────┘ │                   │  └──────┬───────┘ │
│                   │                   │         │          │
│  ┌──────────────┐ │                   │  ┌──────▼───────┐ │
│  │ Process Mgr  │ │  进程管理          │  │ Request      │ │
│  │ - Start()    │──────────────────→  │  │ Router       │ │
│  │ - Stop()     │ │  (启动/停止/重启)   │  └──────┬───────┘ │
│  │ - Health()   │ │                   │         │          │
│  └──────────────┘ │                   │  ┌──────▼───────┐ │
│                   │                   │  │ Model        │ │
│  ┌──────────────┐ │                   │  │ Manager      │ │
│  │ Connection   │ │                   │  └──────────────┘ │
│  │ Pool         │ │                   │                   │
│  └──────────────┘ │                   └──────────────────┘
└──────────────────┘
```

### 6.4 Proto 定义

```protobuf
syntax = "proto3";
package engine;

service EngineService {
  // 一次性推理
  rpc Infer(InferRequest) returns (InferResponse);
  
  // 流式推理（LLM token 流、训练进度）
  rpc InferStream(InferRequest) returns (stream InferChunk);
  
  // 加载模型到缓存
  rpc LoadModel(LoadModelRequest) returns (LoadModelResponse);
  
  // 卸载模型释放显存
  rpc UnloadModel(UnloadModelRequest) returns (UnloadModelResponse);
  
  // 执行插件
  rpc ExecutePlugin(PluginRequest) returns (PluginResponse);
  
  // 流式执行插件
  rpc ExecutePluginStream(PluginRequest) returns (stream PluginChunk);
  
  // 获取引擎状态
  rpc GetStatus(StatusRequest) returns (StatusResponse);
  
  // 健康检查
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}

message InferRequest {
  string model_type = 1;    // vision / nlp / speech / timeseries
  string model_name = 2;    // yolov8n.pt
  bytes input_data = 3;     // 输入数据
  string params = 4;        // JSON 格式参数
}

message InferResponse {
  string output_data = 1;   // JSON 格式输出
  string metadata = 2;      // 元数据
}

message InferChunk {
  string data = 1;          // 数据块
  bool done = 2;            // 是否结束
}

message PluginRequest {
  string plugin_name = 1;   // 插件名
  string inputs = 2;        // JSON 格式输入
  string config = 3;        // JSON 格式配置
  string context = 4;       // JSON 格式上下文
}

message PluginResponse {
  string output = 1;        // JSON 格式输出
  string status = 2;        // success / error
  string error = 3;         // 错误信息
  string metrics = 4;       // JSON 格式指标
}

message PluginChunk {
  string data = 1;
  bool done = 2;
}

message LoadModelRequest {
  string model_name = 1;
  string model_type = 2;
  string device = 3;        // auto / cpu / cuda
}

message LoadModelResponse {
  bool success = 1;
  string error = 2;
  string device = 3;        // 实际加载到的设备
}

message UnloadModelRequest {
  string model_name = 1;
}

message UnloadModelResponse {
  bool success = 1;
}

message StatusRequest {}

message StatusResponse {
  string device = 1;        // cuda / cpu
  int32 gpu_count = 2;
  repeated GPUInfo gpus = 3;
  repeated string loaded_models = 4;
  int64 memory_used_mb = 5;
  int64 memory_total_mb = 6;
}

message GPUInfo {
  int32 index = 1;
  string name = 2;
  int64 memory_total = 3;
  int64 memory_used = 4;
  int64 memory_cached = 5;
  float utilization = 6;
}

message HealthCheckRequest {}
message HealthCheckResponse {
  bool healthy = 1;
  string version = 2;
}
```

### 6.5 进程管理

```
Go Backend 启动流程:
  1. 检测 Python 环境
  2. 启动 Engine 进程: python -m engine.server --port 50051
  3. 等待健康检查通过（最多 30s）
  4. 建立 gRPC 连接池
  5. 预加载常用模型（可选）

进程监控:
  - 每 10s 健康检查
  - 进程崩溃自动重启
  - 重启后重新加载上次活跃的模型
```

### 6.6 降级方案

如果 gRPC 不可用（如 Python 端启动失败），降级为 `os/exec` 模式：

```
降级执行:
  1. 生成临时 Python 脚本
  2. os/exec 执行: python script.py --input xxx --output xxx
  3. 读取输出文件
  4. 返回结果

限制:
  - 无流式支持
  - 每次启动新进程，模型无法缓存
  - 性能较低，仅作为兜底
```

---

## 7. 数据库设计

### 7.1 技术选型

| 场景 | 数据库 | 理由 |
|------|--------|------|
| 桌面单机版 | SQLite | 零配置、单文件、嵌入式 |
| 服务部署版 | PostgreSQL | 并发、扩展、全文搜索 |

通过 GORM 抽象，应用层无需关心底层差异。

### 7.2 表结构设计

#### users（用户表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(36) | PK | UUID |
| username | VARCHAR(64) | UNIQUE, NOT NULL | 用户名 |
| password_hash | VARCHAR(256) | NOT NULL | 密码哈希 |
| avatar | VARCHAR(512) | | 头像路径 |
| role | VARCHAR(16) | DEFAULT 'user' | 角色 (admin/user) |
| settings | TEXT | | JSON 格式用户设置 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

#### projects（项目表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(36) | PK | UUID |
| user_id | VARCHAR(36) | FK, NOT NULL | 所属用户 |
| name | VARCHAR(128) | NOT NULL | 项目名称 |
| description | TEXT | | 项目描述 |
| path | VARCHAR(512) | NOT NULL | 项目文件路径 |
| template | VARCHAR(32) | DEFAULT 'blank' | 模板类型 |
| settings | TEXT | | JSON 格式项目设置 |
| created_at | DATETIME | NOT NULL | |
| updated_at | DATETIME | NOT NULL | |
| deleted_at | DATETIME | INDEX | 软删除 |

#### workflows（工作流表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(36) | PK | UUID |
| project_id | VARCHAR(36) | FK, INDEX, NOT NULL | 所属项目 |
| name | VARCHAR(128) | NOT NULL | 工作流名称 |
| description | TEXT | | 描述 |
| graph | TEXT | NOT NULL | JSON 格式图数据（nodes + edges） |
| version | INT | DEFAULT 1 | 版本号 |
| status | VARCHAR(16) | DEFAULT 'idle' | idle / running |
| created_at | DATETIME | NOT NULL | |
| updated_at | DATETIME | NOT NULL | |
| deleted_at | DATETIME | INDEX | 软删除 |

#### plugins（插件表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(36) | PK | UUID |
| name | VARCHAR(64) | UNIQUE, NOT NULL | 插件标识 |
| version | VARCHAR(16) | NOT NULL | 版本号 |
| type | VARCHAR(16) | NOT NULL | vision/nlp/logic/system/simulation/mcp |
| author | VARCHAR(64) | | 作者 |
| description | TEXT | | 描述 |
| language | VARCHAR(8) | NOT NULL | python/go |
| entry | VARCHAR(256) | NOT NULL | 入口文件 |
| path | VARCHAR(512) | NOT NULL | 插件目录路径 |
| icon | VARCHAR(256) | | 图标路径 |
| ports | TEXT | | JSON 格式端口定义 |
| config_schema | TEXT | | JSON 格式配置 Schema |
| dependencies | TEXT | | JSON 格式依赖信息 |
| install_steps | TEXT | | JSON 格式安装步骤 |
| enabled | BOOLEAN | DEFAULT TRUE | 是否启用 |
| loaded | BOOLEAN | DEFAULT FALSE | 是否已加载 |
| installed_at | DATETIME | | 安装时间 |
| created_at | DATETIME | NOT NULL | |
| updated_at | DATETIME | NOT NULL | |

#### tasks（任务表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(36) | PK | UUID |
| workflow_id | VARCHAR(36) | FK, INDEX, NOT NULL | 关联工作流 |
| project_id | VARCHAR(36) | FK, INDEX, NOT NULL | 关联项目 |
| status | VARCHAR(16) | INDEX, NOT NULL | created/queued/running/success/failed/cancelled |
| progress | REAL | DEFAULT 0 | 0.0 ~ 1.0 |
| inputs | TEXT | | JSON 格式输入数据 |
| config | TEXT | | JSON 格式运行配置 |
| node_states | TEXT | | JSON 格式各节点状态 |
| result | TEXT | | JSON 格式最终输出 |
| error | TEXT | | 错误信息 |
| retry_count | INT | DEFAULT 0 | 重试次数 |
| parent_task_id | VARCHAR(36) | FK | 重试时关联原任务 |
| created_at | DATETIME | NOT NULL | |
| started_at | DATETIME | | |
| finished_at | DATETIME | | |
| duration_ms | INT | | 执行耗时 |

#### logs（日志表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(36) | PK | UUID |
| task_id | VARCHAR(36) | FK, INDEX, NOT NULL | 关联任务 |
| node_id | VARCHAR(36) | | 关联节点 |
| level | VARCHAR(8) | INDEX, NOT NULL | info/warn/error/debug |
| message | TEXT | NOT NULL | 日志内容 |
| raw_message | TEXT | | 原始日志（含 traceback） |
| timestamp | DATETIME | INDEX, NOT NULL | 日志时间 |

#### models（模型表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(36) | PK | UUID |
| project_id | VARCHAR(36) | FK, INDEX, NOT NULL | 所属项目 |
| name | VARCHAR(128) | NOT NULL | 模型名称 |
| type | VARCHAR(32) | NOT NULL | yolo/transformer/lstm/cnn/... |
| framework | VARCHAR(16) | NOT NULL | pytorch/onnx/... |
| path | VARCHAR(512) | NOT NULL | 模型文件路径 |
| size_bytes | INT | | 文件大小 |
| metrics | TEXT | | JSON 格式评估指标 |
| config | TEXT | | JSON 格式训练配置 |
| version | VARCHAR(16) | | 模型版本 |
| created_at | DATETIME | NOT NULL | |

#### agent_sessions（Agent 对话会话表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(36) | PK | UUID |
| user_id | VARCHAR(36) | FK, NOT NULL | 所属用户 |
| project_id | VARCHAR(36) | FK | 关联项目 |
| title | VARCHAR(256) | | 会话标题 |
| message_count | INT | DEFAULT 0 | 消息数 |
| created_at | DATETIME | NOT NULL | |
| updated_at | DATETIME | NOT NULL | |

#### agent_messages（Agent 对话消息表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(36) | PK | UUID |
| session_id | VARCHAR(36) | FK, INDEX, NOT NULL | 所属会话 |
| role | VARCHAR(16) | NOT NULL | user/assistant/system |
| content | TEXT | NOT NULL | 消息内容 |
| actions | TEXT | | JSON 格式 Agent 动作 |
| timestamp | DATETIME | NOT NULL | |

#### mcp_tools（MCP 工具注册表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(36) | PK | UUID |
| name | VARCHAR(64) | UNIQUE, NOT NULL | 工具名 |
| description | TEXT | | 工具描述 |
| server_url | VARCHAR(512) | | MCP 服务端地址 |
| input_schema | TEXT | | JSON 格式输入 Schema |
| enabled | BOOLEAN | DEFAULT TRUE | 是否启用 |
| created_at | DATETIME | NOT NULL | |

### 7.3 索引策略

| 表 | 索引 | 用途 |
|------|------|------|
| workflows | idx_project_id | 按项目查询工作流 |
| tasks | idx_workflow_id | 按工作流查询任务 |
| tasks | idx_status | 按状态筛选任务 |
| tasks | idx_project_id | 按项目查询任务 |
| logs | idx_task_id | 按任务查询日志 |
| logs | idx_timestamp | 按时间范围查询 |
| logs | idx_level | 按级别筛选 |
| agent_messages | idx_session_id | 按会话查询消息 |

---

## 8. 开发路线图

### V1：核心执行引擎（8 周）

**目标**：用户能创建项目、拖拽工作流、运行并查看结果。

```
Week 1-2: 基础框架
├── Go 项目初始化（Gin + GORM + SQLite）
├── 配置管理模块
├── 数据库迁移与模型定义
├── 统一响应/错误码封装
├── 健康检查接口
└── Tauri IPC 桥接

Week 3-4: 项目与工作流
├── 项目 CRUD API
├── 工作流 CRUD API
├── 工作流图数据存储
├── DAG 解析与校验
├── 拓扑排序算法
└── 前端 API 对接（替换 Mock）

Week 5-6: 任务与执行
├── Task 调度器
├── Worker 池
├── Python Engine 进程管理
├── gRPC 通信层
├── Engine 基础服务（健康检查、模型管理）
├── 内置插件：image_input, yolo, if, json_output
└── WebSocket 实时状态推送

Week 7-8: 日志与集成
├── 日志收集与存储
├── 日志查询 API
├── AI 日志分析（基础版）
├── 环境检测（Python/CUDA）
├── 前后端完整联调
└── 端到端测试：创建项目 → 拖拽工作流 → 运行 → 查看结果
```

**V1 交付物**：
- 可运行的桌面应用（Tauri 打包）
- 完整的 Workflow 创建→运行→查看结果 流程
- 4 个内置插件（image_input, yolo, if, json_output）
- 实时日志推送
- SQLite 数据持久化

---

### V2：插件生态（6 周）

**目标**：第三方开发者可编写、安装、发布插件，扩展平台能力。

```
Week 9-10: 插件系统
├── plugin.json 解析与校验
├── 插件安装流程（pip install / 模型下载）
├── 插件注册表
├── 插件加载器（Python/Go/进程隔离）
├── 插件市场 API
└── 前端插件管理页面对接

Week 11-12: 更多内置插件
├── Vision: SAM, OCR, RT-DETR
├── NLP: Transformer, LLM
├── Logic: Switch, Loop, Retry
├── System: Python Script, Git, Terminal, Download
└── 插件测试框架

Week 13-14: 环境与部署
├── Python 虚拟环境管理（venv/conda）
├── CUDA 版本检测与兼容性检查
├── 依赖自动安装
├── 模型导出（ONNX/TensorRT）
├── 断点续跑
└── 插件 SDK 文档完善
```

**V2 交付物**：
- 完整的插件安装/卸载/管理流程
- 15+ 内置插件
- 插件 SDK + 开发文档
- Python 环境自动管理
- 断点续跑能力

---

### V3：Agent 自动化（6 周）

**目标**：AI Agent 能理解用户意图，自动生成和优化工作流。

```
Week 15-16: Agent 基础
├── Agent 会话管理
├── LLM 集成（OpenAI API / 本地模型）
├── SSE 流式对话
├── Agent 动作系统
│   ├── create_workflow
│   ├── modify_node
│   ├── run_workflow
│   ├── analyze_error
│   └── suggest_fix
└── 前端 AI Chat 对接

Week 17-18: Agent 深度集成
├── 上下文感知（当前项目/工作流/错误日志）
├── 工作流自动生成（从自然语言描述）
├── 错误自动分析与修复建议
├── MCP 协议支持
├── MCP 工具注册与调用
└── Agent 记忆与历史

Week 19-20: 优化与发布
├── 性能优化（模型缓存、增量执行）
├── PostgreSQL 支持
├── 多用户与权限
├── 国际化
├── 完整文档
└── V1.0 正式发布
```

**V3 交付物**：
- AI Agent 对话式工作流创建
- 错误自动分析与修复
- MCP 工具生态
- 生产级部署能力
- 完整文档与 SDK

---

## 附录 A：关键设计决策

| 决策 | 选择 | 理由 |
|------|------|------|
| Backend 语言 | Go | 高性能、并发友好、单二进制部署 |
| AI 执行语言 | Python | 生态成熟（PyTorch/Transformers/Ultralytics） |
| 通信协议 | gRPC | 流式支持、强类型、双语言代码生成 |
| 数据库 | SQLite → PostgreSQL | 桌面版零配置，服务版可扩展 |
| HTTP 框架 | Gin | 成熟、高性能、中间件丰富 |
| ORM | GORM | Go 生态最成熟、自动迁移 |
| 前端框架 | Vue3 + Tauri | 已确定，不更改 |

## 附录 B：配置文件规范

```yaml
# config.yaml
server:
  host: "127.0.0.1"
  port: 8080
  mode: "debug"  # debug / release

database:
  driver: "sqlite"  # sqlite / postgres
  dsn: "./data/aistudio.db"
  # dsn: "host=localhost port=5432 user=aistudio dbname=aistudio sslmode=disable"

engine:
  python_path: "python3"
  grpc_port: 50051
  startup_timeout: 30
  health_check_interval: 10
  max_reconnect: 3

task:
  max_workers: 4
  default_timeout: 3600
  max_retry: 3

storage:
  data_dir: "./data"
  projects_dir: "./data/projects"
  plugins_dir: "./data/plugins"
  models_dir: "./data/models"
  runtime_dir: "./data/runtime"
  logs_dir: "./data/logs"

agent:
  llm_provider: "openai"  # openai / local
  api_key: ""
  model: "gpt-4"
  base_url: "https://api.openai.com/v1"
  max_tokens: 4096
  temperature: 0.7

log:
  level: "info"
  format: "json"
  output: "stdout"
```

## 附录 C：错误处理策略

```
错误分类:
├── 用户错误 (4xx)
│   ├── 参数校验失败 → 1001
│   ├── 资源不存在 → 2001/3001/4001
│   └── 权限不足 → 1003
│
├── 系统错误 (5xx)
│   ├── Engine 连接失败 → 5001
│   ├── 模型加载失败 → 5002
│   └── 数据库错误 → 内部处理
│
└── 执行错误
    ├── 节点执行失败 → 记录日志，推送事件
    ├── 插件执行失败 → 记录日志，标记节点 failed
    └── 超时 → 取消任务，标记 timeout

错误恢复:
├── Engine 崩溃 → 自动重启 + 重新加载模型
├── 任务失败 → 保留已完成节点，支持断点续跑
├── 插件崩溃 → 进程隔离，不影响主进程
└── 数据库错误 → 事务回滚 + 日志记录
```