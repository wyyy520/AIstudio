# 编排引擎

## 1. 执行模型

工作流引擎以**有向无环图（DAG）** 方式执行节点。核心步骤：

1. **拓扑排序**：将图转换为线性执行顺序
2. **并行分支**：无依赖的节点可并行执行
3. **数据传递**：通过 Edge 将上游输出映射到下游输入
4. **状态追踪**：实时推送每个节点的执行状态

## 2. 拓扑排序

```
原始图：
  n1 → n2 → n4
         ↘
          n3 → n5

拓扑排序结果：[n1, n2, n3, n4, n5]
```

### 算法

```go
func (g *DAG) TopologicalSort() []string {
    inDegree := make(map[string]int)
    for _, node := range g.Nodes {
        inDegree[node.ID] = 0
    }
    for _, edge := range g.Edges {
        inDegree[edge.To]++
    }
    
    queue := []string{}
    for _, node := range g.Nodes {
        if inDegree[node.ID] == 0 {
            queue = append(queue, node.ID)
        }
    }
    
    result := []string{}
    for len(queue) > 0 {
        node := queue[0]
        queue = queue[1:]
        result = append(result, node)
        
        for _, edge := range g.Edges {
            if edge.From == node {
                inDegree[edge.To]--
                if inDegree[edge.To] == 0 {
                    queue = append(queue, edge.To)
                }
            }
        }
    }
    return result
}
```

## 3. 并行执行

```
  n1 (图像读取)
  ├──→ n2 (YOLO 检测)     ← 可并行
  ├──→ n3 (OCR 识别)       ← 可并行
  └──→ n4 (元数据提取)     ← 可并行
  
  n5 ← 等待 n2/n3/n4 全部完成
```

### 并行调度器

```go
func (e *Executor) RunParallel(ctx context.Context, level []string) {
    var wg sync.WaitGroup
    for _, nodeID := range level {
        wg.Add(1)
        go func(id string) {
            defer wg.Done()
            e.executeNode(ctx, id)
        }(nodeID)
    }
    wg.Wait()
}
```

## 4. 数据路由器

上游节点的输出，通过 Edge 映射到下游节点的输入：

```go
func (e *Executor) routeData(fromNode string, output map[string]interface{}) {
    // 找到所有以 fromNode 为源的边
    for _, edge := range e.graph.Edges {
        if edge.From == fromNode {
            // 获取 from_port 的数据
            data := output[edge.FromPort]
            
            // 写入目标节点的输入缓存
            e.nodeInputs[edge.To][edge.ToPort] = data
        }
    }
}
```

## 5. 错误处理策略

| 策略 | 说明 | 配置方式 |
|------|------|----------|
| fail_fast | 任意节点失败立即停止 | 默认 |
| continue_on_error | 跳过失败节点继续执行 | 节点配置中设置 |
| retry | 自动重试 N 次 | 节点后接 Retry 节点 |

### 示例：带重试的工作流

```
[YOLO 检测] → [Retry: 3次] → [结果输出]
```

## 6. 断点续跑

任务失败后，保留已成功节点的结果，重新运行时跳过已完成节点：

```json
{
  "task_id": "task_20260707_001",
  "resume": true,
  "skip_nodes": ["n1", "n3"]  // 已成功的节点
}
```

## 7. 执行状态机

```
idle → queued → running → completed
               ↘          ↗
                → failed
                 ↘
                  → cancelled
```

| 状态 | 含义 |
|------|------|
| idle | 未启动 |
| queued | 已加入任务队列 |
| running | 执行中 |
| completed | 全部节点成功 |
| failed | 有节点失败且未配置重试 |
| cancelled | 用户手动停止 |

## 8. 性能优化

| 优化项 | 方法 |
|--------|------|
| 节点缓存 | 相同输入参数的已执行节点可复用结果 |
| 增量执行 | 仅重新执行修改过的节点及其下游 |
| 资源池 | 限制并发任务数，避免 GPU 溢出 |
| 模型预热 | 工作流启动前预加载常用模型 |
| 结果压缩 | 大型输出数据压缩后传输 |
