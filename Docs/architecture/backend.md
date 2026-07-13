# 后端架构 (Go Backend)

## 1. 技术选型

| 技术 | 用途 |
|------|------|
| Go 1.21+ | 后端主语言 |
| Gin / Echo | HTTP 框架 |
| GORM | ORM |
| SQLite / PostgreSQL | 数据库 |
| goroutine + channel | 并发任务调度 |
| gRPC | 与 Python Engine 通信 |

## 2. 目录结构说明

```
Backend/
├── cmd/                    # 入口程序
│   └── main.go
├── internal/               # 内部包（不可外部引用）
│   ├── api/               # HTTP 路由与 Handler
│   ├── workflow/          # 工作流引擎核心
│   ├── task/              # 任务调度器
│   ├── agent/             # Agent 管理
│   ├── plugin/            # 插件加载与管理
│   ├── environment/       # 运行环境管理
│   ├── project/           # 项目管理
│   ├── logger/            # 日志系统
│   ├── mcp/               # MCP 协议实现
│   ├── config/            # 配置管理
│   ├── common/            # 公共工具
│   └── database/          # 数据库连接与迁移
├── pkg/                    # 可外部引用的公共包
└── go.mod
```

## 3. 分层架构

```
┌─────────────────────────────────────┐
│           API Layer (api/)           │  HTTP 路由、请求校验、响应封装
├─────────────────────────────────────┤
│         Service Layer                │  业务逻辑
│  ┌──────────┐ ┌──────────┐         │
│  │ Workflow  │ │  Agent   │         │
│  │ Engine    │ │ Manager  │         │
│  └─────┬────┘ └────┬─────┘         │
│        │            │                │
│  ┌─────┴────────────┴─────┐        │
│  │    Task Scheduler       │        │  任务队列、调度
│  └───────────┬────────────┘        │
│              │                      │
│  ┌───────────┴────────────┐        │
│  │   Plugin Manager        │        │  插件加载/调用
│  └───────────┬────────────┘        │
│              │                      │
│  ┌───────────┴────────────┐        │
│  │   Engine Bridge (gRPC)  │        │  Python 引擎桥接
│  └───────────┴────────────┘        │
├─────────────────────────────────────┤
│        Data Layer (database/)        │  数据持久化
└─────────────────────────────────────┘
```

## 4. 核心模块设计

### 4.1 工作流引擎 (workflow/)

```go
// 工作流定义
type Workflow struct {
    ID     string
    Name   string
    Nodes  []Node
    Edges  []Edge
    Status WorkflowStatus
}

// 节点定义
type Node struct {
    ID       string
    Type     NodeType    // vision / nlp / logic / system / ...
    Plugin   string      // 对应插件名
    Config   map[string]interface{}
    Inputs   []Port
    Outputs  []Port
}

// 执行器：拓扑排序后按序执行
type Executor struct {
    graph    *DAG
    scheduler *TaskScheduler
}

func (e *Executor) Run(ctx context.Context, wf *Workflow) error {
    // 1. 拓扑排序
    order := e.graph.TopologicalSort()
    // 2. 逐节点执行（支持并行分支）
    for _, node := range order {
        result := e.executeNode(ctx, node)
        e.graph.SetNodeResult(node.ID, result)
    }
    return nil
}
```

### 4.2 任务调度器 (task/)

```go
type TaskScheduler struct {
    queue   chan *Task
    workers int
}

// 任务状态机
// pending → running → completed / failed / cancelled
type TaskStatus string

const (
    StatusPending   TaskStatus = "pending"
    StatusRunning   TaskStatus = "running"
    StatusCompleted TaskStatus = "completed"
    StatusFailed    TaskStatus = "failed"
    StatusCancelled TaskStatus = "cancelled"
)
```

### 4.3 插件管理器 (plugin/)

```go
type PluginManager struct {
    registry map[string]Plugin
    loader   PluginLoader
}

type Plugin interface {
    Name() string
    Type() PluginType
    Execute(ctx context.Context, input *PluginInput) (*PluginOutput, error)
    Config() PluginConfig
}

// 插件类型
type PluginType string
const (
    PluginVision     PluginType = "vision"
    PluginNLP        PluginType = "nlp"
    PluginLogic      PluginType = "logic"
    PluginSystem     PluginType = "system"
    PluginSimulation PluginType = "simulation"
    PluginMCP        PluginType = "mcp"
)
```

### 4.4 Engine 桥接

```go
// 通过 gRPC 调用 Python Engine
type EngineClient struct {
    conn *grpc.ClientConn
}

func (c *EngineClient) Infer(ctx context.Context, req *InferRequest) (*InferResponse, error) {
    // 调用 Python Engine 的推理接口
    // req: model_name, input_data, params
    // resp: output_data, metadata
}
```

## 5. 数据库模型

```go
// 项目
type Project struct {
    ID        string    `gorm:"primaryKey"`
    Name      string
    Path      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 工作流
type Workflow struct {
    ID        string    `gorm:"primaryKey"`
    ProjectID string    `gorm:"index"`
    Name      string
    Data      string    // JSON 序列化的图数据
    Version   int
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 任务记录
type TaskRecord struct {
    ID         string    `gorm:"primaryKey"`
    WorkflowID string    `gorm:"index"`
    Status     string
    Result     string
    Error      string
    StartedAt  time.Time
    FinishedAt time.Time
}
```
