# 数据流转架构

## 1. 数据流概览

```
用户拖拽编排工作流 → 保存到数据库 → 点击运行
    → Backend 解析工作流图 → 拓扑排序 → 逐节点执行
    → 节点间数据传递 → 结果回传前端 → 实时渲染
```

## 2. 工作流执行数据流

### 2.1 工作流定义 → 执行图

```
前端 JSON 定义:
{
  "nodes": [
    {"id": "n1", "type": "vision", "plugin": "yolo", "config": {...}},
    {"id": "n2", "type": "logic", "plugin": "if", "config": {...}},
    {"id": "n3", "type": "system", "plugin": "terminal", "config": {...}}
  ],
  "edges": [
    {"from": "n1", "to": "n2", "port": "detections"},
    {"from": "n2", "to": "n3", "port": "result"}
  ]
}

         转换为 DAG (有向无环图)

     n1 (YOLO检测)
      │
     n2 (条件判断)
      │
     n3 (终端执行)
```

### 2.2 节点间数据传递

```
Node A 执行完毕
    │
    ├── output: { "detections": [...] }
    │
    ▼
数据路由器 (Data Router)
    │
    ├── 根据 Edge 定义将 output 端口数据
    │   映射到下游 Node 的 input 端口
    │
    ▼
Node B 接收 input: { "detections": [...] }
    │
    ├── 执行插件逻辑
    │
    ▼
Node B output: { "result": true }
```

### 2.3 数据类型系统

| 类型 | 标识 | 说明 | 示例 |
|------|------|------|------|
| image | `image` | 图像数据 | base64 / 文件路径 |
| text | `text` | 文本字符串 | "Hello World" |
| number | `number` | 数值 | 0.95 |
| json | `json` | 结构化数据 | `{"boxes": [...], "scores": [...]}` |
| file | `file` | 文件引用 | `/storage/datasets/img001.jpg` |
| tensor | `tensor` | 张量数据 | numpy array 序列化 |
| stream | `stream` | 流式数据 | LLM token 流 |

## 3. 运行时数据存储

### 3.1 节点运行时状态

```go
type NodeRuntime struct {
    NodeID    string
    Status    string    // idle / running / success / error
    Input     map[string]interface{}   // 输入快照
    Output    map[string]interface{}   // 输出快照
    Error     string
    StartTime time.Time
    EndTime   time.Time
    Duration  time.Duration
}
```

### 3.2 执行上下文流转

```
WorkflowContext
├── workflow_id: "wf_001"
├── task_id: "task_20260707_001"
├── project_id: "proj_001"
├── work_dir: "/runtime/workspace/task_20260707_001/"
├── shared: {}                    # 节点间共享数据
│   ├── "n1.detections" → [...]  # Node 1 的输出缓存
│   └── "n2.result" → true       # Node 2 的输出缓存
└── variables: {}                 # 用户自定义变量
```

## 4. 实时数据推送

```
Backend                    Frontend
  │                           │
  │── WebSocket 连接 ────────│
  │                           │
  │── node_status: running ──→│  节点变黄
  │── node_log: "loading..."──→│  日志面板
  │── node_status: success ──→│  节点变绿
  │── node_output: {...} ─────→│  数据预览
  │                           │
  │── workflow_done ─────────→│  完成提示
```

## 5. 大数据处理策略

| 场景 | 策略 | 说明 |
|------|------|------|
| 图像/视频 | 文件引用 | 不直接传 base64，传文件路径 |
| 大批量数据 | 分块处理 | 流式处理，避免内存溢出 |
| 模型推理结果 | 压缩传输 | JSON 压缩或二进制序列化 |
| 中间结果 | 磁盘缓存 | 落盘到 Runtime/workspace |

## 6. 错误处理与重试

```
节点执行失败
    │
    ├── 检查是否有 Retry 节点
    │   ├── 是 → 重试 N 次后继续
    │   └── 否 → 标记错误，停止下游执行
    │
    ├── 错误信息推送到前端
    │
    └── 已执行节点结果保留（可断点续跑）
```
