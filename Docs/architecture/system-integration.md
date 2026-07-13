# AIStudio 系统整体集成架构设计

## 1. 概述

### 1.1 目标

将分散的 Frontend(Vue)、Backend(Go)、Engine(Python) 三个独立模块整合成**一个完整的桌面应用**，实现：

- 用户点击 `AIStudio.exe` 一键启动
- 所有模块自动启动、自动连接
- 进程生命周期统一管理
- 通信协议统一规范
- 数据格式统一标准

### 1.2 架构原则

| 原则 | 说明 |
|------|------|
| **Monorepo** | 所有代码在一个仓库，统一版本管理 |
| **分层解耦** | Launcher → UI → Backend → Engine 清晰分层 |
| **统一生命周期** | 启动-运行-关闭全流程统一控制 |
| **统一配置** | 全局配置中心化管理 |
| **协议标准化** | API、Event、Data 格式统一 |
| **一键构建** | 从源码到 exe 全自动化 |

---

## 2. Monorepo 目录结构设计

```
AIStudio/
├── .github/
│   └── workflows/              # CI/CD 流水线
│       ├── build.yml
│       └── release.yml
├── .vscode/                    # VSCode 工作区配置
│   └── settings.json
├── Launcher/                   # 🆕 统一启动器 (Rust/Tauri)
│   ├── src/
│   │   ├── main.rs             # 入口：顺序启动所有模块
│   │   ├── process_manager.rs  # 进程管理
│   │   ├── config_manager.rs   # 配置管理
│   │   ├── lifecycle.rs        # 生命周期处理
│   │   └── tray.rs             # 系统托盘
│   ├── build.rs
│   └── Cargo.toml
├── Frontend/                   # 桌面 UI (Vue 3 + Tauri)
│   ├── src/
│   │   ├── components/
│   │   ├── views/
│   │   ├── stores/
│   │   ├── router/
│   │   ├── api/                # API 客户端：调用 Backend HTTP
│   │   └── main.ts
│   ├── public/
│   ├── index.html
│   ├── package.json
│   └── vite.config.ts
├── Backend/                    # 业务后端 (Go)
│   ├── cmd/
│   │   └── main/
│   │       └── main.go         # Backend 入口
│   ├── internal/               # 业务模块（已实现）
│   │   ├── agent/
│   │   ├── api/
│   │   ├── common/
│   │   ├── config/
│   │   ├── database/
│   │   ├── engine/
│   │   ├── environment/
│   │   ├── mcp/
│   │   ├── plugin/
│   │   ├── service/
│   │   ├── task/
│   │   └── workflow/
│   ├── pkg/                    # 公共工具库
│   ├── config/                 # 配置文件（模板）
│   │   ├── default.yaml
│   │   ├── development.yaml
│   │   ├── production.yaml
│   │   └── mcp.json
│   ├── go.mod
│   └── go.sum
├── Engine/                     # AI 计算引擎 (Python)
│   ├── runtime/                # Python 运行时核心
│   ├── sdk/                    # 插件开发 SDK
│   ├── models/                 # 内置模型
│   ├── vision/                 # 视觉能力 (YOLO 等)
│   ├── nlp/                    # NLP 能力
│   ├── requirements.txt        # Python 依赖
│   └── runner.py               # 独立运行入口（供 Backend 调用）
├── Plugins/                    # 插件目录（运行时动态加载）
│   ├── YOLO/
│   │   ├── plugin.json
│   │   └── ...
│   ├── CNN/
│   │   ├── plugin.json
│   │   └── ...
│   └── ...
├── Engine-Packages/            # 🆕 Python 包缓存目录
│   └── ...                     # 避免重复下载
├── Data/                       # 🆕 运行时数据目录
│   ├── projects/               # 用户项目
│   ├── database/               # SQLite 数据库
│   ├── logs/                   # 各模块日志
│   │   ├── launcher.log
│   │   ├── backend.log
│   │   └── engine.log
│   └── temp/                   # 临时文件
├── Docs/                       # 文档（已存在）
│   ├── architecture/
│   │   └── system-integration.md  # 本文档
│   ├── api/
│   ├── plugin-sdk/
│   └── protocol/
├── Resources/                  # 静态资源
│   ├── icons/
│   │   └── app-icon.ico
│   ├── splash/
│   └── ...
├── Build/                      # 🆕 构建输出目录
│   └── ...
├── scripts/                    # 🆕 构建脚本
│   ├── build.ps1               # Windows 全量构建
│   ├── build.sh                # Linux/macOS 全量构建
│   ├── package.ps1             # 打包成安装包
│   └── dev-start.ps1           # 开发环境启动
├── .gitignore
├── CHANGELOG.md
└── README.md
```

**设计要点：**

- **Launcher** 新增，负责统筹所有模块启动
- **Plugins** 独立目录，用户安装的插件放这里，不混入源码
- **Data** 独立目录，存储运行时数据、数据库、日志
- **Engine-Packages** 缓存 Python 依赖，加速安装
- **scripts** 统一构建脚本，一键构建所有模块

---

## 3. Launcher 启动系统设计

### 3.1 角色定位

Launcher 是**唯一入口进程**，用户只启动它。它负责：

1. 读取全局配置
2. 按顺序启动 Frontend → Backend → Engine
3. 监控所有子进程健康状态
4. 处理异常自动重启
5. 提供系统托盘控制
6. 统一处理退出关闭

### 3.2 启动流程

```
用户双击 AIStudio.exe
         ↓
Launcher 启动
         ↓
[1] 初始化日志系统
[2] 加载全局配置 (config.yaml)
[3] 检查数据目录，不存在则创建
[4] 环境检查
    ├─ Python 是否存在？
    ├─ 依赖是否安装？
    └─ Go 后端是否存在？
[5] 启动 Backend (Go HTTP 服务器)
    ├─ 等待端口监听成功 (默认 8081)
    └─ 健康检查 /api/health
[6] 启动 Python Engine
    ├─ 启动 gRPC/HTTP 服务
    └─ 等待就绪
[7] 启动 Tauri/Vue UI
    └─ 窗口显示
[8] 进入监控循环
    ├─ 定期检查各进程状态
    ├─ 异常崩溃自动重启
    └─ 转发信号给子进程
[9] 等待退出信号
    ├─ 按逆序关闭：UI → Engine → Backend
    ├─ 等待优雅关闭完成（超时强制 kill）
    └─ 退出 Launcher
```

### 3.3 进程状态机

```
┌─────────┐
│  Idle   │ 初始状态
└────┬────┘
     │ start
     ↓
┌─────────┐
│ Starting│ 正在启动
└────┬────┘
     │ success
     ↓
┌─────────┐
│ Running │ 正常运行
└────┬────┘
     │ crash / stop
     ↓
┌─────────┐
│Stopping │ 正在停止
└────┬────┘
     │ done
     ↓
┌─────────┐
│ Stopped │ 已停止
└─────────┘
```

### 3.4 健康检查机制

| 模块 | 检查方式 | 间隔 | 失败策略 |
|------|----------|------|---------|
| Backend | GET `/api/health` | 10s | 3 次失败 → 自动重启 |
| Engine | HTTP/GRPC 心跳 | 10s | 2 次失败 → 自动重启 |
| Frontend | 进程存在检查 | 1s | UI 退出 → 触发整体退出 |

---

## 4. 进程生命周期管理

### 4.1 进程树结构

```
AIStudio.exe (Launcher)
├── aistudio-backend.exe (Go Backend)
│   └── [可能 spawn 子进程]
├── python.exe (Python Engine)
│   └── runner.py
└── [Tauri 集成 UI 进程]
```

### 4.2 端口分配策略

| 模块 | 默认端口 | 端口冲突处理 |
|------|---------|-------------|
| Backend HTTP | 8081 | 自动递增寻找空闲端口，更新配置 |
| Engine RPC | 50051 | 自动递增 |
| Frontend Dev | 5173 | 仅开发环境需要 |

Launcher 启动前检测端口占用，自动分配可用端口，并写入**运行时配置**供各模块读取。

### 4.3 启动顺序与依赖

```
Launcher
  ↓
[等待] Backend 健康检查通过
  ↓
[等待] Engine 就绪检查通过
  ↓
Frontend 加载完成 → 显示界面
```

**超时配置：**

- Backend 启动超时：30 秒
- Engine 启动超时：60 秒
- 超时提示用户，提供日志链接

### 4.4 关闭流程

```
用户点击退出 / 收到关闭信号
         ↓
Launcher 开始关闭
         ↓
[1] 关闭 Frontend (UI 先消失，体验好)
         ↓
[2] 停止所有任务 (通知 Backend)
         ↓
[3] 关闭 Engine (优雅关闭，等待处理中任务保存)
         ↓
[4] 关闭 Backend
         ↓
[5] 刷新日志，关闭 Launcher
```

**关闭超时处理：** 5 秒未退出 → 强制 kill

---

## 5. 全局配置系统

### 5.1 配置加载优先级（高 → 低）

1. **环境变量** (`AISTUDIO_*`) - 最高优先级
2. **用户配置** (`Data/config/user.yaml`) - 用户自定义
3. **当前环境配置** (`Backend/config/{env}.yaml`)
4. **默认配置** (`Backend/config/default.yaml`) - 最低优先级

### 5.2 配置结构

```yaml
# config.yaml 全局配置
app:
  name: AIStudio
  version: 1.0.0
  environment: development  # development / production

server:
  host: 0.0.0.0
  port: 8081  # 运行时会被 Launcher 更新

database:
  type: sqlite
  url: Data/database/aistudio.db

engine:
  python_path: auto  # auto: 自动检测，或者指定路径
  engine_dir: Engine
  grpc_port: 50051

plugin:
  directory: Plugins

mcp:
  config_path: Backend/config/mcp.json
  auto_connect: true
  default_timeout: 30000

llm:
  provider: openai  # openai / claude / gemini / mock
  api_key: ""
  base_url: ""
  model: gpt-4o

logging:
  level: info  # debug / info / warn / error
  directory: Data/logs

ui:
  theme: auto
  language: zh-CN
```

### 5.3 配置管理流程

```
Launcher
  ↓
读取默认配置
  ↓
合并环境配置
  ↓
合并用户配置
  ↓
合并环境变量
  ↓
检测端口占用，分配端口
  ↓
写入运行时配置文件 (Data/config/runtime.yaml)
  ↓
Backend 启动时读取运行时配置
  ↓
Engine 启动时读取对应配置
```

### 5.4 存储位置

| 配置类型 | 路径 |
|----------|------|
| 内置默认 | `Backend/config/default.yaml` |
| 环境模板 | `Backend/config/{env}.yaml` |
| 用户配置 | `Data/config/user.yaml` |
| 运行时配置 | `Data/config/runtime.yaml` - **动态生成** |

---

## 6. Frontend/Backend/Engine 统一管理方案

### 6.1 通信架构

```
┌─────────────┐
│  Frontend   │  Vue 3 + Tauri
│  (UI Layer) │
└──────┬──────┘
       │ HTTP REST + WebSocket
       ↓
┌─────────────┐
│  Backend    │  Go
│  (Service Layer)
└──────┬──────┘
       │ 1. HTTP / GRPC
       ↓
┌─────────────┐
│   Engine    │  Python
│ (Compute Layer)
└─────────────┘
```

### 6.2 各层职责划分

| 层 | 职责 | 技术栈 | 入口 |
|---|------|--------|------|
| **Launcher** | 进程启动、生命周期管理、托盘 | Rust + Tauri | `AIStudio.exe` |
| **Frontend** | 用户交互、可视化、Workflow 编辑 | Vue 3 + Vite + Tauri | Tauri 集成 |
| **Backend** | 业务逻辑、API、Workflow 编排、Plugin 管理、Task 调度 | Go | `aistudio-backend.exe` |
| **Engine** | AI 模型执行、Python 计算、模型推理 | Python | `runner.py` |

### 6.3 通信协议

#### Frontend → Backend

- **REST API**：CRUD 操作，请求-响应
- **WebSocket**：实时推送任务进度、日志、事件

**API 统一路径前缀：** `/api/{module}/`

- `/api/health` - 健康检查
- `/api/projects/...` - 项目管理
- `/api/workflows/...` - 工作流
- `/api/tasks/...` - 任务
- `/api/plugins/...` - 插件
- `/api/agent/...` - Agent
- `/api/mcp/...` - MCP
- `/api/environment/...` - 环境

#### Backend → Engine

两种模式：

**模式 A：长驻进程 + RPC** （推荐）

- Engine 启动后常驻
- Backend 通过 gRPC 调用
- 低延迟，适合频繁调用
- 状态保持

**模式 B：每次调用 spawn 进程** （简单任务）

- Backend 每次调用 fork python 进程
- 输入通过 STDIN 传递
- 输出通过 STDOUT 返回
- 适合一次性脚本

当前方案：混合模式，默认长驻 gRPC。

### 6.4 地址发现

Launcher 启动时分配端口，写入 `Data/config/runtime.yaml`：

```yaml
server:
  port: 8081

engine:
  grpc_port: 50051
```

Backend 读取后，Engine 地址就是 `localhost:{grpc_port}`，直接连接。

**不需要**网络发现，因为都在本地。

---

## 7. API 协议规范

### 7.1 统一响应格式

所有 REST API 响应格式统一：

```json
{
  "code": 0,
  "message": "success",
  "data": { ... },
  "request_id": "uuid-xxx",
  "timestamp": 1690000000000
}
```

**Code 约定：**

| code | 含义 |
|------|------|
| `0` | 成功 |
| `-1` | 业务错误 / 参数错误 |
| `-2` | 认证失败 / 未授权 |
| `-3` | 权限不足 |
| `-4` | 资源不存在 |
| `-5` | 系统内部错误 |

### 7.2 分页约定

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [ ... ],
    "total": 100,
    "page": 1,
    "page_size": 20,
    "total_pages": 5
  }
}
```

### 7.3 错误响应格式

```json
{
  "code": -1,
  "message": "Plugin installation failed: dependency not found",
  "error_type": "DependencyError",
  "details": {
    "dependency": "torch>=2.0",
    "current_version": "1.9"
  },
  "request_id": "uuid-xxx",
  "timestamp": 1690000000000
}
```

### 7.4 HTTP 状态码约定

| HTTP 状态码 | 对应 code | 场景 |
|-------------|----------|------|
| 200 OK | 0 / -1 | 正常响应（包括业务错误） |
| 400 Bad Request | -1 | 参数解析失败 |
| 401 Unauthorized | -2 | 未登录 |
| 403 Forbidden | -3 | 无权限 |
| 404 Not Found | -4 | 资源不存在 |
| 500 Internal Error | -5 | 系统异常 |

---

## 8. Event 事件协议

### 8.1 用途

- 实时推送任务进度
- 日志更新
- 系统状态变化
- 插件安装进度

### 8.2 WebSocket 事件格式

所有事件通过 WebSocket 推送，统一格式：

```json
{
  "event_type": "task:progress",
  "topic": "task:123e4567-e89b-12d3-a456-426614174000",
  "data": {
    "task_id": "123e4567-e89b-12d3-a456-426614174000",
    "status": "running",
    "progress": 0.65,
    "message": "Training epoch 5/10"
  },
  "timestamp": 1690000000000
}
```

### 8.3 事件类型约定

| 前缀 | 事件类型 | 说明 |
|------|---------|------|
| `system:` | `system:started` | 系统启动完成 |
| `system:` | `system:error` | 系统错误 |
| `system:` | `system:shutdown` | 系统即将关闭 |
| `task:` | `task:created` | 任务创建 |
| `task:` | `task:started` | 任务开始 |
| `task:` | `task:progress` | 任务进度更新 |
| `task:` | `task:completed` | 任务完成 |
| `task:` | `task:failed` | 任务失败 |
| `task:` | `task:cancelled` | 任务取消 |
| `log:` | `log:append` | 新增日志行 |
| `plugin:` | `plugin:install:progress` | 插件安装进度 |
| `plugin:` | `plugin:installed` | 插件安装完成 |
| `workflow:` | `workflow:node:started` | 节点开始 |
| `workflow:` | `workflow:node:completed` | 节点完成 |
| `agent:` | `agent:thinking` | Agent 思考中 |
| `agent:` | `agent:step:complete` | Agent 一步完成 |
| `agent:` | `agent:complete` | Agent 任务完成 |
| `mcp:` | `mcp:connected` | MCP 服务器连接 |
| `mcp:` | `mcp:disconnected` | MCP 服务器断开 |

### 8.4 订阅机制

Frontend 可以通过 WebSocket 发送订阅指令：

```json
{
  "action": "subscribe",
  "topic": "task:*"  // 所有任务
  // 或 "topic": "task:12345"  // 特定任务
}
```

取消订阅：

```json
{
  "action": "unsubscribe",
  "topic": "task:*"
}
```

---

## 9. Task 数据统一格式

### 9.1 Task 基础结构

```typescript
interface Task {
  id: string;             // UUID
  project_id: string;
  name: string;
  type: TaskType;        // "workflow" | "agent" | "plugin" | "engine" | "mcp"
  status: TaskStatus;    // "pending" | "running" | "completed" | "failed" | "cancelled"
  handler: string;       // 处理模块
  priority: number;      // 1-5，越大优先级越高
  payload: any;          // 输入参数
  progress: number;      // 0-1
  result?: any;          // 输出结果
  error?: string;        // 错误信息
  logs: TaskLog[];
  started_at?: number;   // timestamp
  finished_at?: number;
  duration_ms?: number;
  created_at: number;
}

interface TaskLog {
  level: "debug" | "info" | "warn" | "error";
  message: string;
  timestamp: number;
}

type TaskType = 
  | "workflow"    // 工作流执行
  | "agent"       // Agent 对话
  | "plugin"      // 插件安装/操作
  | "engine"      // Python 引擎任务
  | "mcp"         // MCP 工具调用
  | "environment" // 环境检查/修复
  ;

type TaskStatus =
  | "pending"
  | "running"
  | "completed"
  | "failed"
  | "cancelled"
  ;
```

### 9.2 结果存储约定

- 成功任务：`result` 存储输出数据
- 失败任务：`error` 存储错误信息，`logs` 保存完整日志
- 所有时间戳使用 **Unix 毫秒时间戳**

### 9.3 MCP Task 结果格式

```json
{
  "id": "task-uuid",
  "type": "mcp",
  "status": "completed",
  "payload": {
    "server_name": "SUMO",
    "tool_name": "run_simulation",
    "input": { "vehicle_count": 200 }
  },
  "result": {
    "success": true,
    "output": {
      "simulation_id": "sumo-1234",
      "average_speed_kmh": 47.2,
      "congestion_level": "moderate"
    },
    "duration_ms": 512
  },
  "progress": 1.0
}
```

---

## 10. Plugin 加载机制

### 10.1 插件目录结构

```
Plugins/
└── YOLO/
    ├── plugin.json      # 插件清单（必填）
    ├── manifest.yaml    # 元数据
    ├── __init__.py
    ├── train.py
    ├── predict.py
    ├── nodes/           # Workflow 节点定义
    │   └── detection.py
    └── requirements.txt # Python 依赖
```

### 10.2 插件清单 (plugin.json)

```json
{
  "name": "YOLO",
  "version": "8.2.0",
  "type": "vision",
  "description": "YOLOv8 object detection",
  "author": "AIStudio",
  "license": "MIT",
  "dependencies": [
    {
      "name": "torch",
      "version": ">=2.0",
      "optional": false
    },
    {
      "name": "ultralytics",
      "version": ">=8.2.0",
      "optional": false
    }
  ],
  "nodes": [
    {
      "id": "yolo_detector",
      "name": "YOLO 检测",
      "type": "vision",
      "description": "Detect objects in images using YOLOv8"
    }
  ],
  "metadata": {
    "image_url": "assets/icon.png",
    "tags": ["detect", "vision", "object"]
  }
}
```

### 10.3 加载流程

```
Backend 启动
  ↓
Scan Plugins 目录
  ↓
For each directory:
  ├─ 读取 plugin.json
  ├─ 校验格式
  ├─ 检查依赖（Python 包）
  ├─ 注册到 PluginRegistry
  ├─ 注册节点到 Workflow 注册表
  └─ 标记为 enabled / disabled
  ↓
Discovery 完成，暴露给 Frontend 列表
```

### 10.4 运行时调用

**Workflow 中调用插件节点：**

```
Workflow Engine
  ↓
MCPNode / PluginNode
  ↓
PluginExecutor
  ↓
Forward to Python Engine
  ↓
Execute plugin code
  ↓
Return result
```

### 10.5 插件安装流程

1. 用户在 UI 点击安装
2. Backend 创建 `plugin:install` 任务
3. 下载插件包到临时目录
4. 解析 `plugin.json`
5. 安装 Python 依赖到 `Engine-Packages/`
6. 移动到 `Plugins/{name}/`
7. 注册到注册表
8. 通知 UI 完成

---

## 11. 构建和 Release 方案

### 11.1 开发环境构建

开发者本地构建：

```powershell
# 1. 编译 Backend (Go)
cd Backend
go build -o ../Build/aistudio-backend.exe ./cmd/main

# 2. 构建 Frontend (Vue)
cd ../Frontend
npm install
npm run build

# 3. 编译 Launcher (Rust + Tauri)
cd ../Launcher
cargo build

# 输出: Build/debug/AIStudio.exe
```

### 11.2 生产打包流程

**目标：输出单个安装包 `AIStudio_1.0.0_x64_Setup.exe`**

```
CI 流水线触发构建
  ↓
[1] 版本信息生成
  └─ 写入 version.json
[2] 编译 Backend
  ├─ GOOS=windows GOARCH=amd64
  └─ Output: Build/distribution/aistudio-backend.exe
[3] 构建 Frontend
  ├─ npm ci (锁版本)
  ├─ npm run build
  └─ Output: Frontend/dist/
[4] 编译 Launcher (Tauri 自动集成 Frontend)
  ├─ cargo build --release
  └─ Output: Build/distribution/AIStudio.exe
[5] 收集资源
  ├─ 复制 Backend/config/ → 安装目录
  ├─ 复制 Engine/ → 安装目录
  ├─ 创建 Plugins/ 空目录
  ├─ 创建 Data/ 目录结构
  └─ 复制默认 mcp.json
[6] 使用 Inno Setup 打包
  ├─ 生成安装向导
  ├─ 添加快捷方式到开始菜单
  └─ Output: AIStudio_{version}_x64_Setup.exe
[7] 生成便携版 zip
  └─ 直接压缩整个目录结构
[8] 上传到 GitHub Release
```

### 11.3 目录结构（安装后）

```
C:\Program Files\AIStudio\
├── AIStudio.exe               # Launcher（主入口）
├── aistudio-backend.exe       # Backend
├── Backend\
│   └── config/                # 配置模板
├── Frontend\
│   └── dist/                  # 编译后的 UI 资源
├── Engine\                    # Python 引擎源码
│   ├── runtime/
│   ├── sdk/
│   ├── vision/
│   └── requirements.txt
├── Plugins\                   # 用户安装的插件（初始为空）
├── Resources\                 # 图标等资源
└── uninstall.exe              # 卸载程序
```

**用户数据位置：** `%APPDATA%\AIStudio\Data\`

- 数据库
- 日志
- 用户项目
- 配置

这样设计：程序和数据分离，升级不影响用户数据。

### 11.4 依赖处理

**Python 依赖：**

- 首次启动时，Launcher/Backend 检查 `Engine/requirements.txt`
- 自动 `pip install -r` 到 `Engine-Packages/`（用户数据目录）
- 不影响系统 Python，隔离依赖

**Go 静态编译：**

- Backend 完全静态编译，不需要安装 Go 环境
- 所有依赖打包进 `aist-backend.exe`

### 11.5 更新机制

- 检查 GitHub Release 获取新版本
- 下载新的安装包
- 关闭运行中的 AIStudio
- 覆盖安装程序文件
- 保留用户数据
- 重启完成更新

---

## 12. 错误处理和日志

### 12.1 分层日志

| 模块 | 日志文件 |
|------|---------|
| Launcher | `Data/logs/launcher.log` |
| Backend | `Data/logs/backend.log` |
| Engine | `Data/logs/engine.log` |
| Frontend | 浏览器控制台（开发）|

### 12.2 日志格式

```
2024-01-01 12:00:00 [INFO] [launcher] Starting AIStudio...
2024-01-01 12:00:01 [INFO] [launcher] Backend started successfully on port 8081
2024-01-01 12:00:02 [WARN] [backend] Plugin YOLO dependencies not installed, prompt user
```

### 12.3 用户可见错误

- 启动失败：Launcher 弹出对话框，显示错误和日志路径
- 运行失败：UI 显示错误详情，提供"查看日志"按钮
- 网络错误：明确提示是后端未启动还是 API 错误

---

## 13. 用户体验

### 13.1 首次启动流程

```
用户双击 AIStudio.exe
  ↓
显示启动画面 (Splash Screen)
  ↓
[1] 检测 Python 环境
  ├─ 已安装 → 检查版本
  ├─ 未安装 → 提示用户下载安装 / 或者使用内置捆绑？
  ↓
[2] 安装 Python 依赖 (requirements.txt)
  ↓
[3] 启动 Backend
  ↓
[4] 启动 Engine
  ↓
[5] 检查插件
  ↓
[6] 进入主界面
```

### 13.2 系统托盘

```
AIStudio 图标
  ├─ 打开主窗口
  ├─ 检查更新
  ├─ 查看日志
  ├─ 设置
  └─ 退出
```

---

## 14. 总结

### 14.1 整合后的优点

| 方面 | 改进 |
|------|------|
| **用户体验** | 双击一键启动，所有模块自动连接 |
| **架构清晰** | Launcher → UI → Backend → Engine 四层清晰 |
| **生命周期** | 统一管理启动-监控-关闭，自动重启异常进程 |
| **协议统一** | API、Event、Task 格式标准化 |
| **便于分发** | 一键构建出安装包，用户容易安装 |
| **数据隔离** | 程序文件和用户数据分离，升级安全 |

### 14.2 核心架构图

```
┌─────────────────────────────────────────────────────┐
│                    用户双击                            │
└────────────────┬────────────────────────────────────┘
                 ↓
        ┌───────────────────┐
        │    Launcher       │  Rust 唯一入口
        │  进程/配置管理    │
        └────────┬──────────┘
                 ↓
        ┌───────────────────┐
        │   Frontend (UI)  │  Vue + Tauri
        └────────┬──────────┘
                 ↓  HTTP + WebSocket
        ┌───────────────────┐
        │    Backend        │  Go 业务层
        │ API/Workflow/Agent│
        └────────┬──────────┘
                 ↓  GRPC
        ┌───────────────────┐
        │     Engine        │  Python 计算层
        │  Model Inference  │
        └───────────────────┘
```

### 14.3 实现 roadmap

1. **Phase 1** - 创建 Launcher 目录结构，实现基本进程启动和监控
2. **Phase 2** - 集成 Backend 和 Engine，实现端口分配和配置管理
3. **Phase 3** - Tauri 集成 Frontend，实现系统托盘
4. **Phase 4** - 编写构建脚本，测试安装包生成
5. **Phase 5** - 添加自动更新和错误处理优化

---

**文档版本**: 1.0
**最后更新**: 2026-07-08