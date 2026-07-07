# AIStudio Backend

AIStudio 是一个 AI Agent 工作流开发平台的后端服务。

## 技术栈

- **语言**: Go 1.21+
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: SQLite (开发) / PostgreSQL (生产)
- **任务调度**: 内置优先级队列 + Worker Pool

## 快速启动

### 前置条件

- Go 1.21+
- 网络连接（首次运行需下载依赖）

### 启动

```bash
# 1. 下载依赖
go mod tidy

# 2. 运行服务
go run cmd/main.go
```

服务默认在 `http://localhost:8081` 启动。

### 配置

通过 `.env` 文件或环境变量配置：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DATABASE_TYPE` | 数据库类型 (`sqlite` / `postgres`) | `sqlite` |
| `DATABASE_URL` | 数据库连接字符串 | `aistudio.db` |
| `SERVER_PORT` | HTTP 服务端口 | `8081` |

### 数据库

默认使用 SQLite，首次启动自动创建表结构。

```bash
# 使用 PostgreSQL (需安装 postgres driver)
DATABASE_TYPE=postgres DATABASE_URL="host=localhost user=postgres password=xxx dbname=aistudio port=5432" go run cmd/main.go
```

## 项目结构

```
Backend/
├── cmd/
│   └── main.go              # 启动入口
├── internal/
│   ├── api/
│   │   ├── handlers/         # HTTP 处理器
│   │   ├── middleware/        # 中间件 (Logger, CORS, Recovery)
│   │   └── router.go         # 路由配置
│   ├── database/
│   │   ├── models/           # 数据模型
│   │   ├── config.go         # 数据库配置
│   │   ├── database.go       # 数据库初始化
│   │   └── migrate.go        # 自动迁移
│   ├── plugin/
│   │   ├── executor.go       # 插件执行器
│   │   ├── interfaces.go     # 插件接口定义
│   │   ├── loader.go         # 插件加载器
│   │   ├── manager.go        # 插件管理器
│   │   ├── models.go         # 插件数据模型
│   │   └── registry.go       # 插件注册中心
│   ├── task/
│   │   ├── interfaces.go     # 任务处理器接口
│   │   ├── manager.go        # 任务管理器
│   │   ├── models.go         # 任务数据模型
│   │   ├── queue.go          # 优先级队列
│   │   ├── scheduler.go      # 调度器 (超时回收)
│   │   ├── state.go          # 状态机
│   │   └── worker.go         # Worker Pool
│   └── workflow/             # 工作流引擎 (已有)
├── .env                      # 环境配置
├── go.mod
└── README.md
```

## 系统架构

```
Frontend (Vue3 + TypeScript)
    ↓ REST API
Backend (Gin)
    ├── Task Scheduler (优先级队列 + Worker Pool)
    ├── Plugin Manager (发现/注册/执行)
    ├── Workflow Engine (DAG 执行)
    └── Database (GORM + SQLite/PostgreSQL)
```

## API 文档

查看 [docs/API.md](docs/API.md) 获取完整 API 文档。

## 运行测试

```bash
go test ./...
```